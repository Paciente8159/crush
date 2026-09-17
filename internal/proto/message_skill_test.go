package proto

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/charmbracelet/crush/internal/message"
)

func TestMessageSkillJSONRoundTrip(t *testing.T) {
	t.Parallel()

	msg := Message{
		ID:   "msg-1",
		Role: User,
		Parts: []ContentPart{
			BinaryContent{
				Path:     "grilling",
				MIMEType: "text/markdown",
				Data:     []byte("body"),
				Skill: &message.SkillInfo{
					Name:         "grilling",
					Description:  "Stress-test ideas.",
					Location:     "/skills/grilling/SKILL.md",
					Instructions: "Ask hard questions.",
				},
			},
		},
	}

	data, err := json.Marshal(&msg)
	require.NoError(t, err)

	var out Message
	require.NoError(t, json.Unmarshal(data, &out))

	parts := out.BinaryContent()
	require.Len(t, parts, 1)
	require.NotNil(t, parts[0].Skill)
	require.Equal(t, "grilling", parts[0].Skill.Name)
	require.Equal(t, "Stress-test ideas.", parts[0].Skill.Description)
	require.Equal(t, "/skills/grilling/SKILL.md", parts[0].Skill.Location)
	require.Equal(t, "Ask hard questions.", parts[0].Skill.Instructions)
}

func TestMessageSkillWireOmitsFieldWhenAbsent(t *testing.T) {
	t.Parallel()

	// Existing sessions must stay byte-compatible on the wire: a
	// BinaryContent without a skill must not emit a "skill" key.
	msg := Message{
		ID:   "msg-1",
		Role: User,
		Parts: []ContentPart{
			BinaryContent{Path: "legacy.txt", MIMEType: "text/plain", Data: []byte("old")},
		},
	}

	data, err := json.Marshal(&msg)
	require.NoError(t, err)
	require.NotContains(t, string(data), `"skill"`)
}

func TestSkillInfoJSONFieldNames(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(&message.SkillInfo{
		Name:         "grilling",
		Description:  "desc",
		Location:     "/skills/grilling/SKILL.md",
		Instructions: "body",
	})
	require.NoError(t, err)

	require.Equal(t, `{"name":"grilling","description":"desc","location":"/skills/grilling/SKILL.md","instructions":"body"}`, string(data))
}
