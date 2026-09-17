package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOptionsGetAutoResumeThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  *Options
		expected int
	}{
		{
			name:     "nil options returns 100 (disabled)",
			options:  nil,
			expected: 100,
		},
		{
			name:     "unset field returns 100 (disabled)",
			options:  &Options{},
			expected: 100,
		},
		{
			name:     "zero returns 100 (disabled)",
			options:  &Options{AutoResumeThreshold: 0},
			expected: 100,
		},
		{
			name:     "negative returns 100 (disabled)",
			options:  &Options{AutoResumeThreshold: -5},
			expected: 100,
		},
		{
			name:     "100 returns 100 (disabled)",
			options:  &Options{AutoResumeThreshold: 100},
			expected: 100,
		},
		{
			name:     "above 100 returns 100 (disabled)",
			options:  &Options{AutoResumeThreshold: 150},
			expected: 100,
		},
		{
			name:     "50 is used as-is",
			options:  &Options{AutoResumeThreshold: 50},
			expected: 50,
		},
		{
			name:     "1 is used as-is",
			options:  &Options{AutoResumeThreshold: 1},
			expected: 1,
		},
		{
			name:     "99 is used as-is",
			options:  &Options{AutoResumeThreshold: 99},
			expected: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.options.GetAutoResumeThreshold())
		})
	}
}

func TestOptionsIsAutoResumeEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  *Options
		expected bool
	}{
		{
			name:     "nil options is not enabled",
			options:  nil,
			expected: false,
		},
		{
			name:     "default options is not enabled (threshold 0 = 100)",
			options:  &Options{},
			expected: false,
		},
		{
			name:     "disabled with threshold set is not enabled",
			options:  &Options{DisableAutoResume: true, AutoResumeThreshold: 80},
			expected: false,
		},
		{
			name:     "threshold at 100 is not enabled",
			options:  &Options{AutoResumeThreshold: 100},
			expected: false,
		},
		{
			name:     "enabled with threshold 80",
			options:  &Options{AutoResumeThreshold: 80},
			expected: true,
		},
		{
			name:     "enabled with threshold 1",
			options:  &Options{AutoResumeThreshold: 1},
			expected: true,
		},
		{
			name:     "enabled with threshold 99",
			options:  &Options{AutoResumeThreshold: 99},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.options.IsAutoResumeEnabled())
		})
	}
}

func TestOptionsAutoResumeThresholdFromJSON(t *testing.T) {
	t.Parallel()

	var cfg Config
	require.NoError(t, json.Unmarshal([]byte(`{"options":{"auto_resume_threshold":80}}`), &cfg))
	require.Equal(t, 80, cfg.Options.AutoResumeThreshold)
	require.Equal(t, 80, cfg.Options.GetAutoResumeThreshold())
	require.True(t, cfg.Options.IsAutoResumeEnabled())
}

func TestOptionsAutoResumeDisabledFromJSON(t *testing.T) {
	t.Parallel()

	var cfg Config
	require.NoError(t, json.Unmarshal([]byte(`{"options":{"disable_auto_resume":true}}`), &cfg))
	require.True(t, cfg.Options.DisableAutoResume)
	require.False(t, cfg.Options.IsAutoResumeEnabled())
}