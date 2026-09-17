# Crush Domain Glossary

## Core concepts

| Term | Definition |
|---|---|
| **Auto Resume Session** | A feature that automatically compresses conversation history before sending the next user prompt, triggered when the session's token usage exceeds a configurable threshold. |
| **Pre-send trigger** | A check that runs in the coordinator layer before each user message is dispatched to the LLM. If the token threshold is met, summarization runs first. |
| **Prompt-guided compression** | Compression that receives the user's pending input text and instructs the summarizer to preserve context relevant to that upcoming question. |
| **Greedy chunking** | A chunking strategy for summarization when the small model's context window is too small: pull messages oldest-to-newest into a chunk until the next message would exceed the window, summarize that chunk, replace it with the summary, and repeat until the full conversation fits. |
| **Accumulated summary** | After each greedy chunk is summarized, the summary from all prior chunks is prepended to the next chunk so each compression pass sees the full condensed history. |

## Design decisions

| Term | Definition |
|---|---|
| **Global-only config** | Dialog changes to auto-resume settings always write to persistent global config via typed ConfigStore mutators. No per-session override mechanism. |
| **Coordinator-layer trigger** | The pre-send summarization check lives in the coordinator layer (not the UI), keeping the UI unaware of summarization internals. |
| **StopWhen coexistence** | When auto-resume is active (threshold < 100%), the old `StopWhen` mid-turn summarization is suppressed via an explicit gate rather than removed. |
| **Small model fallback** | If the small model fails mid-chunking, summarization retries with the large model. If both fail, the session is restored from a pre-summarization backup. |
| **Enabled but inert** | `auto-resume` defaults to enabled, but the threshold defaults to 100% (never triggers). Users activate it by lowering the threshold. |