Status: ready-for-agent

# Skill mentions: invoke a skill from the prompt with $

## Problem Statement

Skills exist to encode repeatable expert workflows, but getting one onto the
agent is clumsy. Today the only user-driven path is the `/` command palette,
and it only lists skills that bother to opt in with an explicit frontmatter
flag — so for most skills the list is effectively empty. Everything else
relies on the model noticing that a task matches a skill's description and
loading it on its own. A user who wants a *specific* skill followed for a
specific request has no direct way to say "run this one", and has to either
describe the task in the hope the model picks the right skill, or hunt the
skill out of the palette and attach its raw file, which is a weak signal that
the model may treat as mere context.

## Solution

Typing `$` in the prompt input — at the start of the line or after
whitespace — opens a skill selector that behaves like the existing `@` file
selector: an autocomplete popup above the cursor, fuzzy-filtered as you type,
navigable with the arrow keys, confirmable with enter/tab (or
ctrl+n/ctrl+p to insert without closing). Picking a skill leaves
`$skill-name` in the prompt text and attaches the skill to the message as a
visible, removable chip. The user can attach several skills to one message.
On send, each attached skill is delivered to the agent as an explicit
"loaded skill" instruction block that the model is told to follow, and the
attachment persists with the message so the skill stays in force across
session reloads and subsequent turns.

A skill is selectable by default. The only way to hide a skill from user
invocation is to opt out explicitly in its frontmatter; the same opt-out
logic is applied consistently in the `$` selector and in the command
palette's user-commands list.

## User Stories

1. As a Crush user, I want typing `$` at the start of an empty prompt to open a selector of my skills, so that I can invoke a skill without leaving the input or remembering palette navigation.
2. As a Crush user, I want typing `$` after a space mid-sentence to open the same selector, so that I can append a skill to an in-progress prompt.
3. As a Crush user, I want the selector to stay closed when I type `$` in the middle of a word, so that ordinary text containing a dollar sign (prices, env-var names) doesn't trigger a popup.
4. As a Crush user, I want the selector to list every skill that is invocable, so that I don't have to opt each skill in individually just to be able to call it.
5. As a skill author, I want setting the user-invocable flag to false in my skill's frontmatter to hide my skill from the selector and the command palette, so that I can keep internal-helper skills out of the user's face.
6. As a skill author, I want omitting the user-invocable flag to leave my skill invocable, so that publishing a skill works without knowing about the flag.
7. As a Crush user, I want skills with disable-model-invocation set to still appear in the `$` selector, so that I can explicitly invoke a skill that the model is not allowed to pick up on its own.
8. As a Crush user, I want each selector entry to show the skill's name and a truncated description, so that I can tell similar skills apart before picking one.
9. As a Crush user, I want the selector entries to highlight the characters of my query as I type, so that I can verify I'm looking at the skill I think I am.
10. As a Crush user, I want the selector to fuzzy-filter as I type after `$`, so that I can reach a skill by typing a few distinctive characters rather than its exact name.
11. As a Crush user, I want the selector to rank exact and prefix name matches above deeper fuzzy matches, so that the skill I'm thinking of is on top.
12. As a Crush user, I want to navigate the selector with the arrow keys, wrapping around at the ends, so that I can browse options without a mouse.
13. As a Crush user, I want to confirm a selection with enter or tab, so that I can pick with whichever hand position feels natural.
14. As a Crush user, I want to insert a skill with ctrl+n or ctrl+p without closing the selector, so that I can attach several skills in a row without re-typing `$`.
15. As a Crush user, I want pressing escape or space to dismiss the selector without attaching anything, so that an accidental `$` or a change of mind is costless.
16. As a Crush user, I want the selector to close when I move the cursor back before the `$`, so that editing earlier in the prompt doesn't leave a stale popup.
17. As a Crush user, I want the selected skill to leave the literal `$skill-name` in my prompt text, so that the message I sent visibly records which skills I invoked.
18. As a Crush user, I want a trailing space inserted after a selected skill name, so that continuing to type doesn't merge my sentence into the skill token.
19. As a Crush user, I want each attached skill to appear as a chip in the attachments row above the editor, using the skill icon, so that I can see at a glance what will be injected.
20. As a Crush user, I want to remove a skill chip before sending (as with file attachments), so that I can drop a skill I attached by mistake.
21. As a Crush user, I want to attach multiple different skills to one message, so that a request can combine workflows (for example, a testing discipline plus a research step).
22. As a Crush user, I want selecting a skill that is already attached to be a no-op, so that duplicates don't bloat the prompt or confuse the agent.
23. As a Crush user, I want the `$` selector and the `@` file selector to never be open at the same time, so that the input never has two competing popups.
24. As a Crush user, I want the selector to render above the `$` position and stay within the terminal width, so that it doesn't obscure the prompt or run off-screen.
25. As a Crush user, I want typing `$` with no invocable skills to show no popup, so that an empty list doesn't flash on screen.
26. As a Crush user, I want the trigger character to be rebindable through the keymap configuration, so that `$` can yield to a character I use more.
27. As an agent receiving a message, I want each attached skill to arrive as an explicit loaded-skill block containing its name, description, location and full instructions, so that I can follow the skill without having to fetch its file myself.
28. As an agent, I want a system note accompanying the loaded-skill blocks stating the user attached these skills to this message and I should follow their instructions, so that I treat them as directives rather than background context.
29. As a Crush user, I want attached skills to remain attached after the session is reloaded, so that a resumed conversation keeps following the same skills.
30. As a Crush user, I want skills attached in earlier turns to still be injected on later turns of the same conversation, so that a skill set at the start of a task stays in force for the whole task.
31. As a Crush user, I want skills visible in my user, project and system skill locations to all appear in the selector, so that I don't have to know which path a skill lives in.
32. As a Crush user in client/backend mode, I want the same `$` behavior and the same persistence as in local mode, so that the feature doesn't depend on how Crush is deployed.
33. As a skill author, I want my skill's name and description in the selector to come from its frontmatter, so that the selector is a true preview of the skill's identity.
34. As a Crush user, I want the `@` file selector's behavior to be completely unchanged, so that adding a second trigger doesn't regress my existing workflow.
35. As a Crush user, I want the command palette's user-commands list to show skills under the same default-on rule as the `$` selector, so that one mental model covers both surfaces.
36. As a Crush user, I want skill chips and file chips to coexist in the attachments row of the same message, so that I can combine "here's the relevant file, and here's the workflow" in one request.

## Implementation Decisions

**Skill invocability becomes tri-state.** The skill frontmatter flag
"user-invocable" changes from a boolean defaulting to hidden to a tri-state
where omission means invocable and only an explicit false opts out. A skill's
selector set is the active skills that are user-invocable (regardless of
disable-model-invocation). The resolved (boolean) value is what crosses the
workspace, client, and backend API boundaries, so the wire and persistence
formats are unaffected; the tri-state lives only inside skill parsing.

**The command palette consumes the same resolved flag.** No separate filter
is introduced; the palette's user-commands list simply inherits the new
default-on semantics. This is an intentional, visible behavior change: skills
previously absent from the palette because they never set the flag will now
appear there.

**The popup is a parallel component, not a generalization of the file
popup.** A new skill-selector component is built on the same shared
filterable-list primitive and reuses the same key map (navigation,
select, cancel, insert-without-closing) and the same fuzzy matching with
match highlighting as the file selector. It is kept separate because the two
popups have different item types and different data sources, and the file
popup's behavior must remain untouched.

**Skill items load synchronously.** Unlike file completion, which walks the
filesystem asynchronously, the skill list is already in memory via the
workspace's skill catalog; the popup is populated synchronously from the
catalog, filtered to invocable skills.

**A skill attachment is a first-class content part.** The attachment gains an
optional skill descriptor, and the persisted binary content part of a user
message gains the same optional field (absent by default, so previously saved
sessions load unchanged). The descriptor shape, which came directly from the
grilled design and is the decision-rich core of the persistence contract:

    SkillInfo {
        name         string   // skill name from frontmatter
        description  string   // skill description from frontmatter
        location     string   // skill file location, for reference
        instructions string   // skill body, injected verbatim (escaped)
    }

The raw skill file content continues to be carried as the attachment's data
for display; the descriptor is what drives prompt injection, and only the
descriptor needs to survive serialization.

**Injection happens in prompt assembly, per user message, per turn.** When a
user message is converted for the model, each skill attachment is rendered as
a loaded-skill XML block (name, description, location, instructions) and all
of them are grouped under a single system-information note stating the user
attached these skills to this message and the model should follow their
instructions. Skill blocks are grouped separately from attached-file blocks.
Skill attachments are treated as text for this purpose and are excluded from
the binary file-part conversion used for images. Because prompt assembly runs
over the whole history each turn, earlier skill attachments keep being
injected for the life of the conversation; this prompt-size growth was
explicitly accepted in the design (skill bodies are small relative to the
alternatives).

**Buffer semantics mirror the file selector.** On selection, the typed
`$query` word is replaced by `$skill-name` plus a trailing space, the cursor
moves to the end of the input, and the literal `$skill-name` is sent as part
of the message text alongside the injected blocks. No special parsing of `$`
is added to the outgoing prompt; the dollar token is decoration for humans,
the injection is the signal for the model.

**Trigger rules mirror the file selector.** The selector opens only when the
trigger character is typed at the start of the input or immediately after
whitespace; typing a space, pressing escape, or moving the cursor before the
trigger closes it. Opening one selector closes the other. The trigger key is
added to the editor key map as a user-rebindable binding defaulting to `$`.

**Agent template note.** The coder system prompt gains a short note that a
loaded-skill block in a user message means the user explicitly requested that
skill and it should be followed without the model needing to load the skill
file itself.

**No database or migration changes.** The skill field rides inside the
existing JSON-encoded message content parts.

**Documentation.** A glossary file is created (or extended) with the terms
"skill mention", "skill attachment", "user-invocable" and
"model-invocation-disabled". Two architecture decision records are written:
one for flipping user-invocable to default-on (a hard-to-reverse semantics
change affecting skills in the wild), and one for persisting `$` mentions as
skill attachments with per-turn injection (over the considered alternatives
of ephemeral send-time text, a plain file attachment, or literal command
text). The built-in configuration skill's documentation of skill invocation
and the user-facing docs are updated to describe the `$` trigger.

## Testing Decisions

Good tests for this feature observe external behavior only: what a skill
file parses into and which skills a selector would list; what prompt text a
stored user message produces for the model and what survives a
serialize/deserialize round-trip; and what the editor state (buffer text,
popup open/closed, attached chips) is after a sequence of key presses. No
test reaches into the popup's rendering internals, the list's sorting
implementation, or the agent's internals.

Three seams are used — one per layer the feature spans, each the highest
seam available for that layer:

1. **Skills package API**: given skill files (frontmatter variations), assert
   parsed invocability and the resulting invocable-skill list, and the
   command-palette conversion of that list. Prior art: the existing skills
   and commands test suites.
2. **Message layer**: given a user message carrying skill attachments, assert
   the exact injected prompt text (loaded-skill blocks plus the system note,
   grouped apart from file attachments) and that a message survives the
   content-parts JSON round-trip with the skill field intact. Prior art: the
   existing message content and parts-JSON test suites.
3. **UI model**: drive the Bubble Tea model's update cycle with keypress
   sequences through the existing UI test harness, asserting buffer content,
   popup state, attachment chips, the duplicate-selection no-op, and that the
   file selector is unaffected. Prior art: the existing editor/dialog/
   attachment and file-completion component tests.

An end-to-end agent seam (mock provider asserting what the model receives)
was considered and rejected: the message-layer seam already pins exactly what
the model receives, and an agent-level test would add a mock-provider
dependency for no additional behavioral coverage.

## Out of Scope

- Passing arguments to skills through the `$` trigger (the existing
  command-palette argument-prompting flow is untouched and remains the path
  for that).
- A per-skill icon or preview in the chip; the standard skill icon is used.
- Changes to how the model discovers skills on its own in the system prompt
  (the available-skills listing and the model-invocation-disabled exclusion
  keep their current behavior).
- Publishing the spec to any tracker other than this repo's local-markdown
  tracker, and any ticket breakdown (that is a separate step).
- Editing existing skill frontmatter across the repo to set the flag;
  omission now means invocable, so no per-skill changes are required.
- A "mention history" or recently-used-skills ranking; the selector uses the
  same name-priority ranking as the file selector.

## Further Notes

- The design was stress-tested in a grilling session; the notable accepted
  trade-off is per-turn re-injection of every past skill attachment
  (consistent with how attached files already behave).
- Skill names are restricted to alphanumerics and hyphens by validation, so
  word-based `$query` filtering cannot be broken by spaces in names.
- The `$` character is currently inert in the prompt input, so the trigger
  introduces no conflict; environment-variable-style text typed at a word
  boundary will briefly show a filtered popup, which dismisses on space —
  the same behavior class as `@` with filenames.
- Client/backend mode must be verified end-to-end during implementation:
  the new optional field crosses the client/server message boundary, and
  sessions saved before the change must still load.
