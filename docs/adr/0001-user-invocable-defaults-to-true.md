# ADR 0001: User-invocable defaults to true

## Status

Accepted.

## Context

The Agent Skills specification defines an optional `user-invocable` field. The
original implementation defaulted to `false` when omitted, meaning a skill was
_not_ invocable by the user unless the author explicitly opted in. This is
unfriendly to skill authors: they must remember to set it, and the guarantee
"skills work out of the box" is broken.

After the `$` skill-mention feature, every non-broken skill in the catalog
should be reachable by the user unless the author deliberately opts out.

## Decision

`user-invocable` is **tri-state**:

- `nil` (omitted) → `true` (invocable).
- `true` → `true` (invocable).
- `false` → `false` (not invocable).

The `Skill` struct stores `UserInvocable *bool`. The public method
`IsUserInvocable()` resolves the default:

```go
func (s *Skill) IsUserInvocable() bool {
    return s.UserInvocable == nil || *s.UserInvocable
}
```

Downstream callers (`CatalogEntry`, `FromSkillCatalog`, the `$` popup, the `/`
palette) use the resolved boolean; the wire format carries `user_invocable`
and `model_invocable` as plain `bool` values.

## Consequences

- Existing skills without `user-invocable` become invocable automatically.
- Authors who want a skill hidden from the user set `user-invocable: false`.
- The `disable-model-invocation` flag is independent (controls only model
  visibility).
- The `user-invocable` key is not emitted in JSON when absent
  (`omitempty` + pointer) so the wire format for old sessions is unchanged.