# 04 — $ trigger: skill mention end to end

**What to build:** The whole feature, demoable. Typing `$` at the start of the input or immediately after whitespace opens the skill selector above the cursor, listing the invocable skills from the workspace skill catalog (user, project, and system locations); typing in the middle of a word does not open it. Subsequent typing fuzzy-filters; arrow keys navigate; enter/tab confirms; ctrl+n/ctrl+p insert without closing; escape, space, or moving the cursor before the `$` dismisses. Confirming replaces the `$query` word with `$skill-name` plus a trailing space, moves the cursor to the end, and attaches the skill to the message as a removable chip in the attachments row using the skill icon. Multiple skills can be attached to one message; selecting an already-attached skill is a no-op. The `$` selector and the `@` file selector are never open at the same time (opening one closes the other). The trigger key is a user-rebindable editor binding defaulting to `$`, and the `@` file selector's behavior is completely unchanged. The literal `$skill-name` is sent as part of the message text alongside the injected loaded-skill blocks.

**Blocked by:** 01 (invocable-by-default skill list), 02 (skill attachment and chip), 03 (skill selector popup component).

**Status:** ready-for-agent

- [ ] `$` at input start or after whitespace opens the selector populated with invocable skills; mid-word `$` does not
- [ ] No invocable skills means no popup flashes on screen
- [ ] Typing after `$` filters; space, escape, or cursor-before-`$` dismisses without attaching
- [ ] Enter/tab replaces `$query` with `$skill-name` plus a trailing space and adds the skill chip; cursor moves to the end
- [ ] Ctrl+n/ctrl+p insert the skill without closing the selector, allowing several skills in a row
- [ ] Multiple skill chips coexist with file chips in the attachments row; a skill chip is removable before sending like a file chip
- [ ] Selecting an already-attached skill is a no-op (no duplicate chip, no duplicate injection)
- [ ] Opening the `$` selector closes the `@` file selector and vice versa
- [ ] The trigger is rebindable through the editor key map; default is `$`
- [ ] The `@` file selector behaves exactly as before (existing tests still pass)
- [ ] UI-model tests drive keypress sequences and assert buffer text, popup state, chips, and the dedupe no-op
