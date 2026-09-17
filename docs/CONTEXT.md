# Context: Skill Mentions and Invocability

## Glossary

- **skill mention**: A `$`-prefixed skill name typed in the editor prompt (e.g.
  `$grilling`). It stays in the message buffer as literal text and is sent
  verbatim alongside the injected loaded-skill block.

- **skill attachment**: A first-class content part attached to a user message
  that carries a `SkillInfo` struct (name, description, location, instructions).
  Skill attachments are persisted via `message.BinaryContent.Skill`, rendered
  in the attachment toolbar as a removable chip, and injected into every turn
  as a `<loaded_skill>` XML block.

- **user-invocable**: A tri-state YAML flag (`user-invocable`) in the skill's
  frontmatter. Omitted = invocable (true). Explicit `false` hides the skill
  from the `$` mention popup and the `/` command palette. Explicit `true` is
  a no-op but documents the intent.

- **model-invocation-disabled**: The YAML flag `disable-model-invocation: true`
  keeps the skill out of the system prompt's `<available_skills>` list but
  leaves it invocable via `$` or `/`.

- **Invocation catalog**: The set of skills visible to the user through the
  `$` popup or the `/` palette — entries where `UserInvocable && ModelInvocable`
  (both resolved to `true`).

## ADRs

- [ADR 0001](adr/0001-user-invocable-defaults-to-true.md) — Tri-state
  `user-invocable` defaults to true.
- [ADR 0002](adr/0002-dollar-skill-mentions-persisted-as-attachments.md) —
  Skill attachments persist as first-class message parts.