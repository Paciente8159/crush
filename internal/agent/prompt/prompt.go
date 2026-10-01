package prompt

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/filepathext"
	"github.com/charmbracelet/crush/internal/gitutil"
	"github.com/charmbracelet/crush/internal/home"
	"github.com/charmbracelet/crush/internal/shell"
	"github.com/charmbracelet/crush/internal/skills"
)

// dynamicTailMarker separates a prompt template's static body from its
// session-specific tail (environment, LSP, skills, context files). It is a
// template comment so it renders to nothing. SYSTEM.md and SYSTEM_APPEND.md
// patch the body only; the tail is always appended afterwards.
const dynamicTailMarker = "{{/* dynamic-tail */}}"

// System prompt override files, looked up in each of the project, global
// config, and global data directories.
const (
	systemPromptFile       = "SYSTEM.md"
	systemPromptAppendFile = "SYSTEM_APPEND.md"
)

// Prompt represents a template-based prompt generator.
type Prompt struct {
	name       string
	template   string
	now        func() time.Time
	platform   string
	workingDir string
	patches    bool
}

type PromptDat struct {
	Provider           string
	Model              string
	Config             config.Config
	WorkingDir         string
	IsGitRepo          bool
	Platform           string
	Date               string
	GitStatus          string
	ContextFiles       []ContextFile
	GlobalContextFiles []ContextFile
	AvailSkillXML      string
}

type ContextFile struct {
	Path    string
	Content string
}

type Option func(*Prompt)

func WithTimeFunc(fn func() time.Time) Option {
	return func(p *Prompt) {
		p.now = fn
	}
}

func WithPlatform(platform string) Option {
	return func(p *Prompt) {
		p.platform = platform
	}
}

func WithWorkingDir(workingDir string) Option {
	return func(p *Prompt) {
		p.workingDir = workingDir
	}
}

// WithSystemPromptPatches enables SYSTEM.md and SYSTEM_APPEND.md patching of
// this prompt's static body.
func WithSystemPromptPatches() Option {
	return func(p *Prompt) {
		p.patches = true
	}
}

func NewPrompt(name, promptTemplate string, opts ...Option) (*Prompt, error) {
	p := &Prompt{
		name:     name,
		template: promptTemplate,
		now:      time.Now,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *Prompt) Build(ctx context.Context, provider, model string, store *config.ConfigStore) (string, error) {
	d, err := p.promptData(ctx, provider, model, store)
	if err != nil {
		return "", err
	}

	bodySrc, tailSrc, _ := strings.Cut(p.template, dynamicTailMarker)

	body, err := renderPrompt(p.name, bodySrc, d)
	if err != nil {
		return "", err
	}

	if p.patches && !store.Config().Options.DisableSystemPromptFiles {
		body = applySystemPromptFiles(body, cmp.Or(p.workingDir, store.WorkingDir()))
	}

	tail, err := renderPrompt(p.name+"_tail", tailSrc, d)
	if err != nil {
		return "", err
	}

	return body + tail, nil
}

func renderPrompt(name, src string, d PromptDat) (string, error) {
	if src == "" {
		return "", nil
	}
	t, err := template.New(name).Parse(src)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}
	var sb strings.Builder
	if err := t.Execute(&sb, d); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return sb.String(), nil
}

// systemPromptPatch is one override or append file in the patch order.
type systemPromptPatch struct {
	path   string
	append bool
}

// systemPromptFiles returns the SYSTEM.md and SYSTEM_APPEND.md files to
// apply, lowest priority first so later entries win. Priority 1 is the
// project's .crush directory, then the global config directory, then the
// global data directory; SYSTEM_APPEND.md files append while SYSTEM.md files
// replace.
func systemPromptFiles(workingDir string) []systemPromptPatch {
	configDir := filepath.Dir(config.GlobalConfig())
	dataDir := filepath.Dir(config.GlobalConfigData())
	projectDir := filepath.Join(workingDir, ".crush")

	return []systemPromptPatch{
		{filepath.Join(dataDir, systemPromptAppendFile), true},
		{filepath.Join(dataDir, systemPromptFile), false},
		{filepath.Join(configDir, systemPromptAppendFile), true},
		{filepath.Join(configDir, systemPromptFile), false},
		{filepath.Join(projectDir, systemPromptAppendFile), true},
		{filepath.Join(projectDir, systemPromptFile), false},
	}
}

// applySystemPromptFiles patches body with the SYSTEM.md and
// SYSTEM_APPEND.md files found on disk. SYSTEM_APPEND.md files append,
// SYSTEM.md files replace the accumulated text, and the project directory
// wins over the global ones. Missing or empty files are skipped.
func applySystemPromptFiles(body, workingDir string) string {
	for _, f := range systemPromptFiles(workingDir) {
		content, err := os.ReadFile(f.path)
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(content))
		if text == "" {
			continue
		}
		if f.append {
			body = strings.TrimRight(body, "\n") + "\n\n" + text + "\n\n"
		} else {
			body = text + "\n\n"
		}
	}
	return body
}

func processFile(filePath string) *ContextFile {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}
	return &ContextFile{
		Path:    filePath,
		Content: string(content),
	}
}

func processContextPath(p string, store *config.ConfigStore) []ContextFile {
	var contexts []ContextFile
	fullPath := filepathext.SmartJoin(store.WorkingDir(), p)
	info, err := os.Stat(fullPath)
	if err != nil {
		return contexts
	}
	if info.IsDir() {
		filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				if result := processFile(path); result != nil {
					contexts = append(contexts, *result)
				}
			}
			return nil
		})
	} else {
		result := processFile(fullPath)
		if result != nil {
			contexts = append(contexts, *result)
		}
	}
	return contexts
}

// expandPath expands ~ and environment variables in file paths
func expandPath(path string, store *config.ConfigStore) string {
	path = home.Long(path)
	// Handle environment variable expansion using the same pattern as config
	if strings.HasPrefix(path, "$") {
		if expanded, err := store.Resolver().ResolveValue(path); err == nil {
			path = expanded
		}
	}

	return path
}

// loadContextFiles loads and deduplicates context files from a list of paths.
func loadContextFiles(paths []string, store *config.ConfigStore) map[string][]ContextFile {
	files := map[string][]ContextFile{}
	for _, pth := range paths {
		expanded := expandPath(pth, store)
		pathKey := strings.ToLower(expanded)
		if _, ok := files[pathKey]; ok {
			continue
		}
		files[pathKey] = processContextPath(expanded, store)
	}
	return files
}

func (p *Prompt) promptData(ctx context.Context, provider, model string, store *config.ConfigStore) (PromptDat, error) {
	workingDir := cmp.Or(p.workingDir, store.WorkingDir())
	platform := cmp.Or(p.platform, runtime.GOOS)

	cfg := store.Config()
	contextFiles := loadContextFiles(cfg.Options.ContextPaths, store)
	globalContextFiles := loadContextFiles(cfg.Options.GlobalContextPaths, store)

	// Discover and load skills metadata.
	var availSkillXML string

	// Start with builtin skills.
	allSkills := skills.DiscoverBuiltin()
	builtinNames := make(map[string]bool, len(allSkills))
	for _, s := range allSkills {
		builtinNames[s.Name] = true
	}

	// Discover user skills from configured paths.
	if len(cfg.Options.SkillsPaths) > 0 {
		expandedPaths := make([]string, 0, len(cfg.Options.SkillsPaths))
		for _, pth := range cfg.Options.SkillsPaths {
			expandedPaths = append(expandedPaths, expandPath(pth, store))
		}
		for _, userSkill := range skills.Discover(expandedPaths) {
			if builtinNames[userSkill.Name] {
				slog.Warn("User skill overrides builtin skill", "name", userSkill.Name)
			}
			allSkills = append(allSkills, userSkill)
		}
	}

	// Deduplicate: user skills override builtins with the same name.
	allSkills = skills.Deduplicate(allSkills)

	// Filter out disabled skills.
	allSkills = skills.Filter(allSkills, cfg.Options.DisabledSkills)

	if len(allSkills) > 0 {
		availSkillXML = skills.ToPromptXML(allSkills)
	}

	isGit := isGitRepo(store.WorkingDir())
	data := PromptDat{
		Provider:      provider,
		Model:         model,
		Config:        *cfg,
		WorkingDir:    filepath.ToSlash(workingDir),
		IsGitRepo:     isGit,
		Platform:      platform,
		Date:          p.now().Format("1/2/2006"),
		AvailSkillXML: availSkillXML,
	}
	if isGit {
		var err error
		data.GitStatus, err = getGitStatus(ctx, store.WorkingDir())
		if err != nil {
			return PromptDat{}, err
		}
	}

	for _, files := range contextFiles {
		data.ContextFiles = append(data.ContextFiles, files...)
	}
	for _, files := range globalContextFiles {
		data.GlobalContextFiles = append(data.GlobalContextFiles, files...)
	}
	return data, nil
}

func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func getGitStatus(ctx context.Context, dir string) (string, error) {
	sh := shell.NewShell(&shell.Options{
		WorkingDir: dir,
	})
	branch := getGitBranch(dir)
	status, err := getGitStatusSummary(ctx, sh)
	if err != nil {
		return "", err
	}
	commits, err := getGitRecentCommits(ctx, sh)
	if err != nil {
		return "", err
	}
	return branch + status + commits, nil
}

// getGitBranch reads the checked-out branch straight from the repository.
// This runs on every prompt build, so it stays off the shell: spawning git
// each turn is wasted work, and a detail of prompt assembly has no business
// passing through command blocking and permission policy.
func getGitBranch(dir string) string {
	branch := gitutil.CurrentBranch(dir)
	if branch == "" {
		return ""
	}
	return fmt.Sprintf("Current branch: %s\n", branch)
}

func getGitStatusSummary(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git status --short 2>/dev/null | head -20")
	if err != nil {
		return "", nil
	}
	out = strings.TrimSpace(out)
	if out == "" {
		return "Status: clean\n", nil
	}
	return fmt.Sprintf("Status:\n%s\n", out), nil
}

func getGitRecentCommits(ctx context.Context, sh *shell.Shell) (string, error) {
	out, _, err := sh.Exec(ctx, "git log --oneline -n 3 2>/dev/null")
	if err != nil || out == "" {
		return "", nil
	}
	out = strings.TrimSpace(out)
	return fmt.Sprintf("Recent commits:\n%s\n", out), nil
}

func (p *Prompt) Name() string {
	return p.name
}
