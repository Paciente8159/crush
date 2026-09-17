# 02: Summarize with prompt guidance, chunking, and model fallback

**What to build:** The agent's `Summarize` method gains an optional `prompt` parameter. When non-empty, the user's pending input is injected into `buildSummaryPrompt` as guidance ("The user is about to ask: <prompt>. Keep this in mind when summarizing."). When empty, behavior is identical to today (backward compatible for the manual "Summarize Session" command). When the small model is selected for summarization, implement greedy oldest-to-newest chunking: accumulate messages into a chunk until the next would exceed 90% of the small model's context window, summarize that chunk, replace it with the summary, and prepend the accumulated summary from prior chunks before each subsequent pass. Before the first chunk, backup the session's raw message history. If the small model fails mid-chunking, retry remaining chunks with the large model. If both fail, restore the backup and surface the error.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] `Summarize` method signature gains optional `prompt` parameter
- [ ] All stub implementations across test files updated to match new signature
- [ ] `buildSummaryPrompt` injects prompt guidance when parameter is non-empty
- [ ] `buildSummaryPrompt` behavior unchanged when parameter is empty
- [ ] Greedy oldest-to-newest chunking: messages accumulated until next exceeds 90% of small model's context window
- [ ] Accumulated summary from prior chunks prepended before each new chunk
- [ ] Session backup before first chunk
- [ ] Model fallback: small model failure → retry with large model
- [ ] Both models fail → restore backup, surface error
- [ ] Small model context window is 0 → fall back to large immediately
- [ ] Agent-level tests cover prompt injection, chunk boundary, accumulated summary, model fallback, session restore