package skills

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalogInvocabilityFlags(t *testing.T) {
	t.Parallel()

	falsity := false

	skillDefaults := &Skill{Name: "defaults", Description: "Flag omitted.", SkillFilePath: "/skills/defaults/SKILL.md"}
	skillUserOff := &Skill{Name: "user-off", Description: "Hidden from user.", SkillFilePath: "/skills/user-off/SKILL.md", UserInvocable: &falsity}
	skillModelOff := &Skill{Name: "model-off", Description: "Hidden from model.", SkillFilePath: "/skills/model-off/SKILL.md", DisableModelInvocation: true}

	entries := Catalog([]*Skill{skillDefaults, skillUserOff, skillModelOff}, nil, "")
	require.Len(t, entries, 3)

	require.Equal(t, true, entries[0].UserInvocable)
	require.Equal(t, true, entries[0].ModelInvocable)
	require.Equal(t, false, entries[1].UserInvocable)
	require.Equal(t, true, entries[1].ModelInvocable)
	require.Equal(t, true, entries[2].UserInvocable)
	require.Equal(t, false, entries[2].ModelInvocable)
}
