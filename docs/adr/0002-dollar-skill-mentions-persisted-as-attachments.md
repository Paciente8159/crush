# ADR 0002: `$` skill mentions persisted as attachments

## Status

Accepted.

## Context

When a user selects a skill from the `$` mention popup, the resulting
`$skill-name` literal in the buffer is not enough for the agent to know which
skill was attached. The skill's instructions must be re-injected on every turn
(session replay, summarization, etc.). This requires durable persistence in
the message store, not just in-memory state.

## Decision

Skill selections create a **first-class `message.BinaryContent` part** with
`Skill *message.SkillInfo` set. This part:

- Uses `MIMEType: "text/markdown"` and carries the full SKILL.md content as
  raw `Data`, so existing code paths (text attachment rendering) handle it
  naturally.
- Carries the parsed `SkillInfo` (name, description, location, instructions)
  for efficient re-injection.
- Is serialized with `json:"skill,omitempty"` so legacy messages without a
  skill attachment are byte-identical on the wire.
- Is skipped from `FilePart` conversion in `ToAIMessage` (the skill
  instructions are folded into the user text via `PromptWithTextAttachments`
  instead).

The persistent `$skill-name` text stays in the editor buffer as a literal.

## Consequences

- Skill attachments survive session save/load, JSON round-trip, and server
  ↔ client sync.
- The agent receives **both** the literal mention (`$skill-name` in the prompt
  text) and the `<loaded_skill>` block (injected by `PromptWithTextAttachments`
  before the text part).
- Multiple skills per message are supported; re-selecting an already-attached
  skill is a no-op (dedup in `attachments.Attachments`).