# 01: Config layer — options, shell builtins, ConfigStore mutators

**What to build:** Add the three new auto-resume options to the config system. `DisableAutoResume` (inverted bool), `AutoResumeThreshold` (int 0–100, default 100), and `AutoResumeModel` (enum "large"/"small", default "large") go on `Config.Options`. Corresponding shell builtins (`auto-resume`, `auto-resume-threshold`, `auto-resume-model`) register in the shellconfig options table. Typed `ConfigStore` mutators (`SetAutoResumeThreshold`, `SetAutoResumeModel`, `SetDisableAutoResume`) follow the existing copy-on-write + targeted JSON write pattern. The coordinator reads these options when `coordinator.Run` is called. `crush info` displays the values.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] `DisableAutoResume`, `AutoResumeThreshold`, `AutoResumeModel` fields added to `Config.Options` struct
- [ ] Shell builtins registered for all three options in `shellconfig/options.go`
- [ ] Typed `ConfigStore` mutators implemented following `UpdatePreferredModel` pattern
- [ ] Config field defaults match spec table
- [ ] `crush info` displays new option values
- [ ] Shell option parsing tests pass (following `options_test.go` pattern)
- [ ] Config field displays correctly in `crush_info_test.go` (following `TestCrushInfo_AutoSummarizeInversion`)