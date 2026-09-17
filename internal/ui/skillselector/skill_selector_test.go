package skillselector

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/require"
)

func testSkills() []Skill {
	return []Skill{
		{ID: "/skills/grilling/SKILL.md", Name: "grilling", Description: "Stress-test plans."},
		{ID: "/skills/grill-me/SKILL.md", Name: "grill-me", Description: "Grill the user."},
		{ID: "/skills/tdd/SKILL.md", Name: "tdd", Description: "Test-driven development."},
	}
}

func newTestSelector() *SkillSelector {
	return New(lipgloss.NewStyle(), lipgloss.NewStyle(), lipgloss.NewStyle())
}

func filteredNames(s *SkillSelector) []string {
	names := make([]string, 0, len(s.filtered))
	for _, item := range s.filtered {
		names = append(names, nameOf(item))
	}
	return names
}

func TestSetItemsOpensWithAllSkills(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	require.False(t, s.IsOpen())

	s.SetItems(testSkills())

	require.True(t, s.IsOpen())
	require.True(t, s.HasItems())
	require.Equal(t, []string{"grilling", "grill-me", "tdd"}, filteredNames(s))
}

func TestSetItemsEmpty(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(nil)

	require.True(t, s.IsOpen())
	require.False(t, s.HasItems())
	require.Empty(t, s.Render())
}

func TestFilterNarrows(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	s.Filter("tdd")
	require.Equal(t, []string{"tdd"}, filteredNames(s))

	s.Filter("")
	require.Equal(t, []string{"grilling", "grill-me", "tdd"}, filteredNames(s))
}

func TestFilterRanksExactAndPrefixAboveFuzzy(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems([]Skill{
		{ID: "1", Name: "xgrill", Description: "Fuzzy contains."},
		{ID: "2", Name: "grill-me", Description: "Prefix match."},
		{ID: "3", Name: "grilling", Description: "Exact match."},
	})

	s.Filter("grill")
	require.Equal(t, []string{"grilling", "grill-me", "xgrill"}, filteredNames(s))
}

func TestFilterNoMatch(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	s.Filter("zzz-no-match")
	require.False(t, s.HasItems())
	require.Empty(t, filteredNames(s))
}

func TestSelectClosesAndReturnsValue(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	msg, ok := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, ok)
	require.False(t, s.IsOpen())

	selection, ok := msg.(SelectionMsg)
	require.True(t, ok)
	require.Equal(t, "grilling", selection.Value.Name)
	require.Equal(t, "/skills/grilling/SKILL.md", selection.Value.ID)
	require.False(t, selection.KeepOpen)
}

func TestSelectWrapsAround(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	// The list renders reversed: the first item sits at the bottom next to
	// the input, so "down" moves toward the top of the popup.
	_, ok := s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.True(t, ok)
	msg, ok := s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, ok)
	require.Equal(t, "tdd", msg.(SelectionMsg).Value.Name)

	// Two more downs land on the middle item.
	s.SetItems(testSkills())
	_, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.True(t, ok)
	_, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.True(t, ok)
	msg, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, ok)
	require.Equal(t, "grill-me", msg.(SelectionMsg).Value.Name)

	// Wrapping: down past the top returns to the first item.
	s.SetItems(testSkills())
	_, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.True(t, ok)
	_, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.True(t, ok)
	_, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	require.True(t, ok)
	msg, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, ok)
	require.Equal(t, "grilling", msg.(SelectionMsg).Value.Name)
}

func TestInsertWithoutClosingKeepsOpenAndMoves(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	msg, ok := s.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	require.True(t, ok)
	selection, ok := msg.(SelectionMsg)
	require.True(t, ok)
	require.Equal(t, "tdd", selection.Value.Name)
	require.True(t, selection.KeepOpen)
	require.True(t, s.IsOpen(), "ctrl+n must keep the popup open")

	// Next insert-continue selects the following skill.
	msg, ok = s.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	require.True(t, ok)
	require.Equal(t, "grill-me", msg.(SelectionMsg).Value.Name)

	// Wraps back to the first skill, still open.
	msg, ok = s.Update(tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	require.True(t, ok)
	require.Equal(t, "grilling", msg.(SelectionMsg).Value.Name)
	require.True(t, s.IsOpen())
}

func TestUpInsertMovesTowardTop(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	// From the first item (bottom), ctrl+p moves toward the top.
	msg, ok := s.Update(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	require.True(t, ok)
	selection, ok := msg.(SelectionMsg)
	require.True(t, ok)
	require.Equal(t, "grill-me", selection.Value.Name)
	require.True(t, selection.KeepOpen)

	// And ctrl+p from the top wraps back to the first item.
	s.SetItems(testSkills())
	_, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.True(t, ok)
	_, ok = s.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	require.True(t, ok)
	msg, ok = s.Update(tea.KeyPressMsg{Code: 'p', Mod: tea.ModCtrl})
	require.True(t, ok)
	require.Equal(t, "grilling", msg.(SelectionMsg).Value.Name)
}

func TestCancelCloses(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	msg, ok := s.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, ok)
	require.IsType(t, ClosedMsg{}, msg)
	require.False(t, s.IsOpen())
}

func TestIgnoresKeysWhenClosed(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())
	s.Close()

	for _, msg := range []tea.KeyPressMsg{
		{Code: tea.KeyEnter},
		{Code: tea.KeyDown},
		{Code: tea.KeyEscape},
	} {
		_, handled := s.Update(msg)
		require.False(t, handled, "closed popup must not consume %v", msg)
	}
}

func TestIgnoresNonPopupKeysWhenOpen(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())

	for _, msg := range []tea.KeyPressMsg{
		{Code: 'g', Text: "g"},
		{Code: tea.KeySpace, Text: " "},
		{Code: tea.KeyBackspace},
	} {
		_, handled := s.Update(msg)
		require.False(t, handled, "popup must not consume %v", msg)
	}
}

func TestUpdateAfterCloseIsInert(t *testing.T) {
	t.Parallel()

	s := newTestSelector()
	s.SetItems(testSkills())
	s.Close()
	s.SetItems(testSkills())
	require.True(t, s.IsOpen())
}
