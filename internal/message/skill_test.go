package message

import (
	"strings"
	"testing"

	"charm.land/fantasy"
	"github.com/stretchr/testify/require"
)

func testSkillInfo() *SkillInfo {
	return &SkillInfo{
		Name:         "grilling",
		Description:  "Stress-test <ideas> & plans.",
		Location:     "/skills/grilling/SKILL.md",
		Instructions: "Ask hard \"questions\".",
	}
}

func TestFormatLoadedSkill(t *testing.T) {
	t.Parallel()

	got := testSkillInfo().FormatLoadedSkill()

	want := `<loaded_skill>
  <name>grilling</name>
  <description>Stress-test &lt;ideas&gt; &amp; plans.</description>
  <location>/skills/grilling/SKILL.md</location>
  <instructions>
Ask hard &quot;questions&quot;.
  </instructions>
</loaded_skill>`
	require.Equal(t, want, got)
}

func TestPromptWithTextAttachments_SkillsGroupedSeparately(t *testing.T) {
	t.Parallel()

	file := Attachment{FilePath: "/a.txt", MimeType: "text/plain", Content: []byte("file body")}
	out := PromptWithTextAttachments("do it", []Attachment{
		{Skill: &SkillInfo{Name: "one", Description: "First.", Instructions: "i1"}},
		{Skill: &SkillInfo{Name: "two", Description: "Second.", Instructions: "i2"}},
		file,
	})

	require.Equal(t, 2, strings.Count(out, "<system_info>"))
	require.Equal(t, 1, strings.Count(out, "The skills below were attached by the user to this message"))
	require.Contains(t, out, "The files below have been attached by the user")
	require.Contains(t, out, "<file path='/a.txt'>")

	// Both skill blocks share one skills note and are adjacent, before the
	// file block.
	require.Contains(t, out, "<name>one</name>")
	require.Contains(t, out, "<name>two</name>")
	require.Contains(t, out, "i1\n  </instructions>\n</loaded_skill>\n\n<loaded_skill>")
	require.Less(t, strings.Index(out, "<name>two</name>"), strings.Index(out, "<file path="))
	require.Equal(t, 2, strings.Count(out, "<loaded_skill>"))
}

func TestPromptWithTextAttachments_NoSkillsKeepsFileOnlyNote(t *testing.T) {
	t.Parallel()

	out := PromptWithTextAttachments("do it", []Attachment{
		{FilePath: "/a.txt", MimeType: "text/plain", Content: []byte("x")},
	})

	require.Contains(t, out, "The files below have been attached by the user")
	require.NotContains(t, out, "<loaded_skill>")
	require.Equal(t, 1, strings.Count(out, "<system_info>"))
}

func TestToAIMessage_SkillInjection(t *testing.T) {
	t.Parallel()

	msg := &Message{
		Role: User,
		Parts: []ContentPart{
			TextContent{Text: "do it"},
			BinaryContent{
				Path:     "grilling",
				MIMEType: "text/markdown",
				Data:     []byte("Ask hard questions."),
				Skill:    testSkillInfo(),
			},
		},
	}

	messages := msg.ToAIMessage()
	require.Len(t, messages, 1)

	var textPart *fantasy.TextPart
	for _, p := range messages[0].Content {
		if tp, ok := p.(fantasy.TextPart); ok {
			textPart = &tp
		}
		if _, ok := p.(fantasy.FilePart); ok {
			t.Fatalf("skill attachment must not become a FilePart")
		}
	}
	require.NotNil(t, textPart, "skill injection must live in the text part")
	require.Contains(t, textPart.Text, "do it")
	require.Contains(t, textPart.Text, "<loaded_skill>")
	require.Contains(t, textPart.Text, "<name>grilling</name>")
	require.Contains(t, textPart.Text, "Follow their instructions")
	require.Contains(t, textPart.Text, "Ask hard &quot;questions&quot;.")
}

func TestBinaryContentSkillJSONRoundTrip(t *testing.T) {
	t.Parallel()

	withSkill := BinaryContent{
		Path:     "grilling",
		MIMEType: "text/markdown",
		Data:     []byte("body"),
		Skill:    testSkillInfo(),
	}
	withoutSkill := BinaryContent{
		Path:     "legacy.txt",
		MIMEType: "text/plain",
		Data:     []byte("old session data"),
	}

	data, err := marshalParts([]ContentPart{withSkill, withoutSkill})
	require.NoError(t, err)

	parts, err := unmarshalParts(data)
	require.NoError(t, err)
	require.Len(t, parts, 2)

	got1, ok := parts[0].(BinaryContent)
	require.True(t, ok)
	require.NotNil(t, got1.Skill)
	require.Equal(t, withSkill.Skill, got1.Skill)

	got2, ok := parts[1].(BinaryContent)
	require.True(t, ok)
	require.Nil(t, got2.Skill, "legacy parts without a skill stay plain files")
	require.Equal(t, withoutSkill.Path, got2.Path)
}
