# 03: Coordinator pre-send check and StopWhen gate

**What to build:** In `coordinator.Run`, before dispatching to `agent.Run`, read the auto-resume config and check the current session's token usage against the threshold. If `(prompt_tokens + completion_tokens) / context_window * 100 >= threshold`, call `Summarize(prompt)` first, then proceed to `agent.Run` with the original prompt unmodified. If auto-resume is disabled (`DisableAutoResume` is false) or threshold is 0 or 100, skip summarization. In the existing `StopWhen` closure in `agent.Run`, add a gate: when auto-resume is active (threshold < 100), return false so the old mid-turn auto-summarize doesn't fire. When disabled or at 100%, the old behavior runs unchanged. The coordinator passes the user's prompt text to `Summarize` for prompt-guided compression.

**Blocked by:** 01 (config options), 02 (Summarize with prompt parameter)

**Status:** ready-for-agent

- [ ] `coordinator.Run` reads auto-resume config before dispatching to agent
- [ ] Threshold check: `(prompt_tokens + completion_tokens) / context_window * 100 >= threshold` triggers summarization
- [ ] Threshold 0 or 100 → skip summarization (disable)
- [ ] Auto-resume disabled → skip summarization
- [ ] `Summarize(prompt)` called when threshold met, user's prompt passed through
- [ ] Original user prompt sent unmodified after compression completes
- [ ] StopWhen gate: auto-resume active (threshold < 100) → StopWhen returns false
- [ ] Auto-resume disabled or 100% → StopWhen runs old behavior
- [ ] Coordinator-level tests cover: threshold met → Summarize called, threshold not met → skip, disabled → skip, 0% → skip, 100% → skip