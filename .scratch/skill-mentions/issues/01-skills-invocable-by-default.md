# 01 — Skills invocable by default

**What to build:** A skill is invocable by default. Setting the user-invocable flag to false in a skill's frontmatter is the only way to hide it from user invocation; omitting the flag leaves it invocable. Skills hidden from the model stay hidden from user invocation as well. The command palette's user-commands list applies this same default-on rule, so skills that never set the flag now appear there. From the user's perspective: the invocable-skill list (the set any skill selector would show) contains every skill they expect, and the palette stops hiding skills that never opted in.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] A skill file with no user-invocable flag parses as invocable
- [ ] A skill file with an explicit false user-invocable flag parses as not invocable
- [ ] The resolved (boolean) invocability is what crosses the workspace, client, and backend API boundaries; the tri-state lives only inside skill parsing
- [ ] The invocable-skill list (active skills that are user-invocable and not hidden from the model) is exposed for consumers
- [ ] The command palette's user-commands list shows skills under the default-on rule and excludes explicit-false and model-hidden skills
- [ ] Tests at the skills-package seam cover omitted, explicit true, and explicit false frontmatter, plus the invocable list and the palette conversion
- [ ] Existing sessions, wire formats, and persisted data are unaffected by the change
