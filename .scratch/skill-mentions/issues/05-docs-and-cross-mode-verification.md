# 05 — Agent template note, docs, cross-mode verification

**What to build:** The closing slice that makes the feature legible to the model and to humans. The coder system prompt gains a short note that a loaded-skill block in a user message means the user explicitly requested that skill and it should be followed without loading the skill file itself. A domain glossary file records the terms "skill mention", "skill attachment", "user-invocable", and "model-invocation-disabled". Two architecture decision records are written: one for flipping user-invocable to default-on (hard-to-reverse semantics change affecting skills in the wild), and one for persisting `$` mentions as skill attachments with per-turn injection (against the considered alternatives of ephemeral send-time text, a plain file attachment, or literal command text). The built-in configuration skill's documentation of skill invocation and the user-facing docs are updated to describe the `$` trigger. Finally, the feature is verified end to end in client/backend mode and across a session reload.

**Blocked by:** 04 ($ trigger end to end).

**Status:** ready-for-agent

- [ ] The coder system prompt explains that a loaded-skill block is an explicit user request to follow that skill
- [ ] The glossary file defines all four terms without implementation detail
- [ ] ADR for the default-on user-invocable flip, recording the trade-off and the impact on existing skills
- [ ] ADR for persisted skill mentions with per-turn injection, recording the rejected alternatives
- [ ] The built-in configuration skill docs and user-facing docs describe invoking a skill with `$`
- [ ] Verified end to end in client/backend mode: selector, chip, injection, and persistence all work
- [ ] Verified across a session reload: attached skills persist and remain injected on later turns
