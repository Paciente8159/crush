package model

import (
	"context"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/config"
	"github.com/charmbracelet/crush/internal/message"
	"github.com/charmbracelet/crush/internal/skills"
	"github.com/charmbracelet/crush/internal/ui/attachments"
	"github.com/charmbracelet/crush/internal/ui/completions"
	"github.com/charmbracelet/crush/internal/ui/dialog"
	"github.com/charmbracelet/crush/internal/ui/skillselector"
	"github.com/stretchr/testify/require"
)

// skillTestWorkspace stubs the skill catalog for mention-flow tests.
type skillTestWorkspace struct {
	testWorkspace
	skills    []skills.CatalogEntry
	readSkill map[string]skills.SkillReadResult
	readBody  map[string][]byte
}

func (w *skillTestWorkspace) ListSkills(context.Context) ([]skills.CatalogEntry, error) {
	return w.skills, nil
}

func (w *skillTestWorkspace) ReadSkill(_ context.Context, id string) ([]byte, skills.SkillReadResult, error) {
	return w.readBody[id], w.readSkill[id], nil
}

func newSkillMentionTestUI(t *testing.T) *UI {
	t.Helper()

	u := newTestUI()
	u.dialog = dialog.NewOverlay()
	u.keyMap = DefaultKeyMap()
	sty := u.com.Styles.Attachments
	u.attachments = attachments.New(
		attachments.NewRenderer(sty.Normal, sty.Deleting, sty.Image, sty.Text, sty.Skill, sty.Remove),
		attachments.Keymap{},
	)
	u.completions = completions.New(
		u.com.Styles.Completions.Normal,
		u.com.Styles.Completions.Focused,
		u.com.Styles.Completions.Match,
	)
	u.skillsPopup = skillselector.New(
		u.com.Styles.Completions.Normal,
		u.com.Styles.Completions.Focused,
		u.com.Styles.Completions.Match,
	)
	u.com.Workspace = &skillTestWorkspace{
		testWorkspace: testWorkspace{
			cfg: &config.Config{Options: &config.Options{TUI: &config.TUIOptions{}}},
		},
		skills: []skills.CatalogEntry{
			{ID: "/skills/grilling/SKILL.md", Name: "grilling", Description: "Stress-test plans.", UserInvocable: true, ModelInvocable: true},
			{ID: "/skills/tdd/SKILL.md", Name: "tdd", Description: "Test-driven development.", UserInvocable: true, ModelInvocable: true},
			{ID: "/skills/hidden/SKILL.md", Name: "hidden", Description: "Not invocable."},
			{ID: "/skills/modelonly/SKILL.md", Name: "modelonly", Description: "Disabled for model, still usable by user.", UserInvocable: true, ModelInvocable: false},
		},
		readSkill: map[string]skills.SkillReadResult{
			"/skills/grilling/SKILL.md": {Name: "grilling", Description: "Stress-test plans."},
		},
		readBody: map[string][]byte{
			"/skills/grilling/SKILL.md": []byte("---\nname: grilling\ndescription: Stress-test plans.\n---\nAsk hard questions.\n"),
		},
	}
	u.updateLayoutAndSize()
	return u
}

// loadSkillsIntoPopup simulates the async catalog load completing.
func loadSkillsIntoPopup(t *testing.T, u *UI) {
	t.Helper()
	loaded := u.loadSkillItems()()
	require.IsType(t, skillselector.ItemsLoadedMsg{}, loaded)
	_, _ = u.Update(loaded)
}

func pressRune(u *UI, r rune) {
	_, _ = u.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
}

func TestSkillMentionTriggerAtStart(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	pressRune(u, '$')
	require.True(t, u.skillsPopupOpen)
	require.Equal(t, 0, u.skillsPopupStartIndex)
	require.Equal(t, "$", u.textarea.Value())

	loadSkillsIntoPopup(t, u)
	require.True(t, u.skillsPopup.HasItems())
	// The non-invocable skill is filtered out; the model-hidden skill appears.
	filtered := u.skillsPopup.FilteredSkills()
	require.Len(t, filtered, 3)
	require.Equal(t, "grilling", filtered[0].Name)
	require.NotContains(t, u.skillsPopup.Render(), "hidden")
}

func TestSkillMentionMidWordDoesNotOpen(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	for _, r := range "foo" {
		pressRune(u, r)
	}
	pressRune(u, '$')

	require.False(t, u.skillsPopupOpen)
	require.Equal(t, "foo$", u.textarea.Value())
}

func TestSkillMentionFilterAndSelect(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	pressRune(u, '$')
	loadSkillsIntoPopup(t, u)

	for _, r := range "gri" {
		pressRune(u, r)
	}
	require.True(t, u.skillsPopupOpen)
	require.Equal(t, "$gri", u.textarea.Value())
	require.Len(t, u.skillsPopup.FilteredSkills(), 1)

	_, _ = u.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// The buffer keeps the literal mention with a trailing space.
	require.Equal(t, "$grilling ", u.textarea.Value())
	require.False(t, u.skillsPopupOpen)

	// The selection attaches the skill with parsed instructions.
	att := u.attachSkill("/skills/grilling/SKILL.md", "grilling")()
	require.IsType(t, message.Attachment{}, att)
	attachment := att.(message.Attachment)
	require.NotNil(t, attachment.Skill)
	require.Equal(t, "grilling", attachment.Skill.Name)
	require.Equal(t, "Ask hard questions.", attachment.Skill.Instructions)
	require.Equal(t, "Stress-test plans.", attachment.Skill.Description)

	_, _ = u.Update(attachment)
	require.Len(t, u.attachments.List(), 1)
	require.Equal(t, "grilling", u.attachments.List()[0].FileName)
}

func TestSkillMentionSelectingAttachedSkillIsNoOp(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	att := u.attachSkill("/skills/grilling/SKILL.md", "grilling")()
	attachment := att.(message.Attachment)

	_, _ = u.Update(attachment)
	_, _ = u.Update(attachment)

	require.Len(t, u.attachments.List(), 1, "duplicate skill attachments are a no-op")
}

func TestSkillMentionSpaceDismisses(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	pressRune(u, '$')
	loadSkillsIntoPopup(t, u)

	_, _ = u.Update(tea.KeyPressMsg{Code: tea.KeySpace, Text: " "})

	require.False(t, u.skillsPopupOpen)
	require.Equal(t, "$ ", u.textarea.Value())
	require.Empty(t, u.attachments.List())
}

func TestSkillMentionEscapeDismisses(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	pressRune(u, '$')
	loadSkillsIntoPopup(t, u)

	_, _ = u.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	require.False(t, u.skillsPopupOpen)
	require.Equal(t, "$", u.textarea.Value())
	require.Empty(t, u.attachments.List())
}

func TestSkillMentionDeletionBeforeTriggerDismisses(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	pressRune(u, '$')
	loadSkillsIntoPopup(t, u)

	// Backspace removes the $, moving the cursor before the trigger point.
	_, _ = u.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})

	require.False(t, u.skillsPopupOpen)
	require.Equal(t, "", u.textarea.Value())
}

func TestSkillMentionMutualExclusionWithFilePopup(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)

	// Opening the skill popup closes the file popup when the latter is open.
	u.completionsOpen = true
	pressRune(u, '$')
	require.False(t, u.completionsOpen)
	require.True(t, u.skillsPopupOpen)

	// Opening the file popup closes the skill popup when the latter is open.
	// Press '@' from an empty/whitespace-prefixed buffer.
	u.skillsPopupOpen = true
	u.textarea.SetValue("")
	pressRune(u, '@')
	require.False(t, u.skillsPopupOpen)
	require.True(t, u.completionsOpen)
}

func TestSkillMentionTriggerIsRebindable(t *testing.T) {
	t.Parallel()

	u := newSkillMentionTestUI(t)
	u.keyMap.Editor.MentionSkill = key.NewBinding(
		key.WithKeys("%"),
		key.WithHelp("%", "mention skill"),
	)

	pressRune(u, '%')
	require.True(t, u.skillsPopupOpen)
	require.Equal(t, "%", u.textarea.Value())

	loadSkillsIntoPopup(t, u)
	for _, r := range "gri" {
		pressRune(u, r)
	}
	require.Len(t, u.skillsPopup.FilteredSkills(), 1)

	_, _ = u.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.Equal(t, "%grilling ", u.textarea.Value())
}
