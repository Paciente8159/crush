# Auto Resume Session

Status: ready-for-agent

## Problem Statement

Crush users with long coding sessions frequently hit LLM context limits mid-task. The existing auto-summarize (`StopWhen`) fires mid-turn with a hardcoded threshold and always uses the expensive large model, giving users no control over when or how compression happens. Users who want fine-grained control — a custom context threshold, a cheaper model for summarization, or compression that preserves what they're about to ask — have no way to configure it.

## Solution

An **Auto Resume Session** feature that compresses conversation history before the next user message is sent to the LLM, triggered when session token usage exceeds a user-configurable threshold. Compression uses a user-selected model (small or large) and is guided by the user's pending input to preserve relevant context. The feature is accessible from the command palette (Ctrl+P) and configurable via shell builtins in `crushrc`.

## User Stories

1. As a user with long coding sessions, I want conversation history to be automatically compressed before my next prompt when context is running low, so that I don't hit context limits mid-task.

2. As a cost-conscious user, I want to use my configured small model for summarization, so that compression uses minimal API cost.

3. As a user about to ask a specific question, I want the summarizer to know what I'm about to ask, so that it preserves relevant context and drops irrelevant details.

4. As a user who wants predictable behavior, I want to set a specific context threshold at which compression triggers, so that I control exactly when the summarizer runs.

5. As a user who prefers the old behavior, I want to disable auto-resume entirely, so that the old StopWhen auto-summarize handles compression as before.

6. As a user who uses the existing auto-summarize toggle, I want auto-resume to be independent from auto-summarize, so that disabling one doesn't affect the other.

7. As a user adjusting settings mid-session, I want to open the auto-resume dialog from the command palette (Ctrl+P), adjust the threshold or model, and have changes persist for future sessions.

8. As a user running a local model that may go down, I want compression to fall back to the large model if the small model is unavailable or fails, so that compression still works.

9. As a user with a small model that has limited context, I want the summarizer to chunk the conversation into pieces that fit the small model's context window, so that compression works with any model size.

10. As a user waiting for compression to finish, I want to see a brief status message ("Auto-resuming…") when the summarizer runs, so that I know why there's a delay.

11. As a user who uses multiple agents, I want auto-resume model selection to use my existing configured small/large models, so that I don't have to configure summarization models separately.

12. As a user whose prompt triggered compression, I want my original prompt sent unmodified after compression completes, so that my intent is preserved exactly.

13. As a user who set the threshold to 100%, I want auto-resume to never trigger, so that the feature is effectively disabled without changing multiple settings.

14. As a user who accidentally set the threshold to 0%, I want it to behave the same as disabled (never summarize), so that I can't accidentally trigger summarization on every message.

15. As a user whose small model fails partway through chunked summarization, I want the remaining chunks compressed by the large model and the session restored to a valid state, so that summarization recovers gracefully without data loss.

## Implementation Decisions

### Architecture

- **Pre-send trigger in coordinator layer**: The threshold check lives in `coordinator.Run`, before dispatching to `agent.Run`. The coordinator has access to config (threshold, model), the session (token counts), and can call `agent.Summarize(prompt)`. The UI layer (`sendMessageInternal`) remains unaware of summarization details.

- **Prompt-guided compression**: The user's current input is passed as a parameter to `Summarize` and injected into `buildSummaryPrompt`. The summarizer receives context like "The user is about to ask: <prompt>. Keep this in mind when summarizing."

- **Greedy oldest-to-newest chunking**: When the small model's context window is smaller than the conversation to compress, messages are accumulated oldest-first into a chunk until adding the next message would exceed 90% of the small model's context window. That chunk is summarized, replaced with the summary, and the process repeats until the full conversation fits.

- **Accumulated summaries**: After each chunk is summarized, the accumulated summary from all prior chunks is prepended before the next chunk. Each compression pass sees the full condensed history of everything older.

- **Session backup**: Before the first chunk is compressed, the session's raw message history is backed up. If both small and large model summarization fail, the backup is restored and an error is surfaced to the user.

- **StopWhen coexistence gate**: When auto-resume is enabled (threshold < 100%), the existing `StopWhen` auto-summarize closure returns false via an explicit gate. When disabled or at 100%, the old behavior runs unchanged.

### Config Storage

New fields on the `Config.Options` struct:

| Option | Config field | Type | Default | Notes |
|---|---|---|---|---|
| `auto-resume` | `DisableAutoResume` | bool (inverted) | `true` | Inverted: `true` in config = feature enabled |
| `auto-resume-threshold` | `AutoResumeThreshold` | int | `100` | 0–100; floor at 1; 0 or 100 = disabled |
| `auto-resume-model` | `AutoResumeModel` | string | `"large"` | `"large"` or `"small"`; maps to existing `SelectedModelType` |

All changes persist globally via typed `ConfigStore` mutators following the existing copy-on-write + targeted JSON write pattern (same as `UpdatePreferredModel`). The dialog always writes to global config; there is no per-session override.

### Dialog

- New entry in the command palette: **"Auto Resume Session"** (Ctrl+A shortcut)
- Opens a config dialog with:
  - **Resume Model**: toggle between "Small" and "Large" (default: Large)
  - **Context Threshold**: integer slider/input 0–100 (default: 100%)
- Changes persist immediately to global config via `ConfigStore` typed mutators
- No "Save" button needed — changes take effect on dialog close

### Model Fallback

When "Small" is selected:
- Availability is checked at config load time (warn if unavailable, fall back to large)
- At trigger time, if the small model call fails mid-chunking, retry with the large model
- If both fail, restore pre-summarization session backup and surface error

### Summarize Method Changes

`Summarize` gains an optional `prompt` parameter. When non-empty, it is injected into `buildSummaryPrompt` as guidance. When empty (manual "Summarize Session" from command palette), behavior is identical to today.

## Testing Decisions

### What makes a good test

Test external behavior and decisions, not implementation details. For the pre-send check, test that summarization is or isn't called based on threshold. For chunking, test the boundary at which a chunk is cut. Don't test which exact LLM response was generated — test the orchestration.

### Primary seam: Agent-level Summarize tests

Test the modified `Summarize` method using a mock fantasy backend (following the existing `mockSessionAgent` pattern in `coordinator_test.go`). Tests cover:

- Prompt guidance injected into summary prompt when prompt parameter is non-empty
- No prompt injection when prompt parameter is empty (backward compatibility)
- Chunk boundary computation: conversation exceeding 90% of small model's context window triggers a chunk cut
- Accumulated summary prepended to each subsequent chunk
- Model fallback: small model fails → large model retried on remaining chunks
- Session backup restored when both models fail

Prior art: `coordinator_test.go` uses `mockSessionAgent` with configurable `runFunc`. A similar mock fantasy agent pattern exists for testing agent behavior without real LLM calls.

### Secondary seam: Coordinator pre-send check

Test the trigger logic in `coordinator.Run`:

- Threshold met (usage >= threshold) → `Summarize` is called before `Run`
- Threshold not met → `Run` is called without `Summarize`
- Auto-resume disabled → no summarization, regardless of threshold
- Threshold at 100% → no summarization (disabled even if enabled)
- Threshold at 0% → no summarization (same as disabled)

Prior art: `coordinator_test.go` tests `coordinator.Run` with customizable `mockSessionAgent` run functions.

### Config loading

Test the three new shell builtin options (following `internal/shellconfig/options_test.go`):

- `auto-resume` parses as inverted bool
- `auto-resume-threshold` parses as int with 0–100 range
- `auto-resume-model` parses with enum validation ("large"/"small")
- Default values match spec table

## Out of Scope

- **Per-session config overrides**: All auto-resume settings are global only. Per-session overrides are not supported.
- **Custom chunking strategies**: Only greedy oldest-to-newest chunking is implemented. Overlapping windows, semantic chunk boundaries, or ML-based chunking are not supported.
- **Custom summary prompt templates**: The summary prompt is fixed (though prompt-guided compression injects the user's pending input). Users cannot customize the summary instruction.
- **Post-hoc summarization**: Compression only happens pre-send. No background/scheduled summarization.
- **UI feedback beyond status message**: No progress bar for multi-chunk compression. A single-line "Auto-resuming…" message is shown while compression runs.
- **Third-party model chunking**: Chunking is designed for the small model only. The large model is expected to have sufficient context for any conversation.

## Further Notes

- This feature replaces the hardcoded `StopWhen` auto-summarize only when enabled (threshold < 100%). The old `StopWhen` logic remains unchanged for users who don't use auto-resume.
- The existing `auto-summarize` (`disable_auto_summarize`) option is independent from `auto-resume`. Users can have auto-summarize off and auto-resume on, or vice versa, or both.
- The `Ctrl+A` shortcut for opening the dialog is available (no conflict with existing bindings) and serves as a mnemonic for "Auto."
- When the small model is selected but its context window is unknown (0), the summarizer falls back to the large model immediately rather than attempting chunking with an unknown boundary.