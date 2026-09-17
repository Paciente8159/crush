# 03 — Skill selector popup component

**What to build:** A new skill-selector popup component, parallel to the existing file-completion popup and built on the same shared filterable-list primitive. It presents skill items (name plus a truncated description, with fuzzy match-character highlighting), ranks exact and prefix name matches above deeper fuzzy matches, and reuses the file selector's key map: arrow-key navigation with wraparound, enter/tab to select, ctrl+n/ctrl+p to insert without closing, escape/space to cancel. It accepts its item list synchronously from the caller (no async loading, no filtering policy inside the component) and stays visually within the terminal width. The existing file popup is untouched.

**Blocked by:** None — can start immediately (prefactor for the `$` trigger wiring).

**Status:** ready-for-agent

- [ ] The component renders skill items with name and truncated description
- [ ] Fuzzy filtering as the query changes, with match-character highlighting, behaves like the file selector
- [ ] Exact and prefix name matches rank above path-segment/fallback matches
- [ ] The key map matches the file selector: navigate (wraparound), select, insert-without-closing, cancel
- [ ] Items load synchronously from the caller-supplied list; no filesystem or async work inside the component
- [ ] Component tests cover filtering, ranking, navigation, and all selection outcomes using synthetic skills
- [ ] The existing file-completion component and its tests are unchanged
