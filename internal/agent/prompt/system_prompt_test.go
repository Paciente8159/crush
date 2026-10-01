package prompt

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/stretchr/testify/require"
)

// writeFile writes content to dir/name, creating parent directories.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
}

func TestSystemPromptFilesOrder(t *testing.T) {
	configDir := t.TempDir()
	dataDir := t.TempDir()
	t.Setenv("CRUSH_GLOBAL_CONFIG", configDir)
	t.Setenv("CRUSH_GLOBAL_DATA", dataDir)
	workingDir := t.TempDir()

	files := systemPromptFiles(workingDir)
	require.Len(t, files, 6)

	want := []systemPromptPatch{
		{filepath.Join(dataDir, systemPromptAppendFile), true},
		{filepath.Join(dataDir, systemPromptFile), false},
		{filepath.Join(configDir, systemPromptAppendFile), true},
		{filepath.Join(configDir, systemPromptFile), false},
		{filepath.Join(workingDir, ".crush", systemPromptAppendFile), true},
		{filepath.Join(workingDir, ".crush", systemPromptFile), false},
	}
	require.Equal(t, want, files)
}

func TestApplySystemPromptFiles(t *testing.T) {
	t.Run("no files leaves body untouched", func(t *testing.T) {
		t.Setenv("CRUSH_GLOBAL_CONFIG", t.TempDir())
		t.Setenv("CRUSH_GLOBAL_DATA", t.TempDir())
		require.Equal(t, "builtin", applySystemPromptFiles("builtin", t.TempDir()))
	})

	t.Run("project replace wins over everything", func(t *testing.T) {
		configDir := t.TempDir()
		dataDir := t.TempDir()
		t.Setenv("CRUSH_GLOBAL_CONFIG", configDir)
		t.Setenv("CRUSH_GLOBAL_DATA", dataDir)
		workingDir := t.TempDir()

		writeFile(t, dataDir, systemPromptAppendFile, "data append")
		writeFile(t, configDir, systemPromptFile, "config replace")
		writeFile(t, filepath.Join(workingDir, ".crush"), systemPromptFile, "project replace")

		got := applySystemPromptFiles("builtin", workingDir)
		require.Equal(t, "project replace\n\n", got)
	})

	t.Run("append stacks on the builtin body", func(t *testing.T) {
		configDir := t.TempDir()
		dataDir := t.TempDir()
		t.Setenv("CRUSH_GLOBAL_CONFIG", configDir)
		t.Setenv("CRUSH_GLOBAL_DATA", dataDir)
		workingDir := t.TempDir()

		writeFile(t, dataDir, systemPromptAppendFile, "data append")
		writeFile(t, configDir, systemPromptAppendFile, "config append")
		writeFile(t, filepath.Join(workingDir, ".crush"), systemPromptAppendFile, "project append")

		got := applySystemPromptFiles("builtin", workingDir)
		require.Equal(t, "builtin\n\ndata append\n\nconfig append\n\nproject append\n\n", got)
	})

	t.Run("replace discards earlier appends", func(t *testing.T) {
		configDir := t.TempDir()
		dataDir := t.TempDir()
		t.Setenv("CRUSH_GLOBAL_CONFIG", configDir)
		t.Setenv("CRUSH_GLOBAL_DATA", dataDir)
		workingDir := t.TempDir()

		writeFile(t, dataDir, systemPromptAppendFile, "data append")
		writeFile(t, configDir, systemPromptFile, "config replace")
		writeFile(t, filepath.Join(workingDir, ".crush"), systemPromptAppendFile, "project append")

		got := applySystemPromptFiles("builtin", workingDir)
		require.Equal(t, "config replace\n\nproject append\n\n", got)
	})

	t.Run("empty files are skipped", func(t *testing.T) {
		configDir := t.TempDir()
		dataDir := t.TempDir()
		t.Setenv("CRUSH_GLOBAL_CONFIG", configDir)
		t.Setenv("CRUSH_GLOBAL_DATA", dataDir)
		workingDir := t.TempDir()

		writeFile(t, filepath.Join(workingDir, ".crush"), systemPromptFile, "   \n\n")

		require.Equal(t, "builtin", applySystemPromptFiles("builtin", workingDir))
	})
}

func TestBuildSystemPromptPatches(t *testing.T) {
	workingDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(workingDir, ".crush"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(workingDir, ".crush", systemPromptFile), []byte("CUSTOM BODY"), 0o644))

	t.Setenv("CRUSH_GLOBAL_CONFIG", t.TempDir())
	t.Setenv("CRUSH_GLOBAL_DATA", t.TempDir())

	const tmpl = "BUILTIN BODY\n\n" + dynamicTailMarker + "<env>\nWork: {{.WorkingDir}}\n</env>\n"

	newStore := func() *config.ConfigStore {
		return config.NewTestStore(&config.Config{Options: &config.Options{}})
	}

	t.Run("override replaces the body but keeps the dynamic tail", func(t *testing.T) {
		p, err := NewPrompt("test", tmpl, WithWorkingDir(workingDir), WithSystemPromptPatches())
		require.NoError(t, err)

		out, err := p.Build(context.Background(), "provider", "model", newStore())
		require.NoError(t, err)
		require.True(t, strings.HasPrefix(out, "CUSTOM BODY"), out)
		require.NotContains(t, out, "BUILTIN BODY")
		require.Contains(t, out, "<env>")
		require.Contains(t, out, filepath.ToSlash(workingDir))
	})

	t.Run("disable flag ignores the files", func(t *testing.T) {
		store := config.NewTestStore(&config.Config{Options: &config.Options{DisableSystemPromptFiles: true}})

		p, err := NewPrompt("test", tmpl, WithWorkingDir(workingDir), WithSystemPromptPatches())
		require.NoError(t, err)

		out, err := p.Build(context.Background(), "provider", "model", store)
		require.NoError(t, err)
		require.Contains(t, out, "BUILTIN BODY")
		require.NotContains(t, out, "CUSTOM BODY")
	})

	t.Run("patches disabled without the option", func(t *testing.T) {
		p, err := NewPrompt("test", tmpl, WithWorkingDir(workingDir))
		require.NoError(t, err)

		out, err := p.Build(context.Background(), "provider", "model", newStore())
		require.NoError(t, err)
		require.Contains(t, out, "BUILTIN BODY")
		require.NotContains(t, out, "CUSTOM BODY")
	})
}
