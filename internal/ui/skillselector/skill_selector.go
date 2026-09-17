// Package skillselector implements the skill-mention popup opened by the
// editor's skill trigger key. It is a parallel component to the file
// completions popup: it shares the filterable list and item rendering
// primitives, but has its own state, keys, and messages. The existing file
// popup is not generalized and remains untouched.
package skillselector

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/crush/internal/ui/completions"
	"github.com/charmbracelet/crush/internal/ui/list"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/ordered"
)

const (
	minHeight = 1
	maxHeight = 10
	minWidth  = 10
	maxWidth  = 100

	tierExactName = iota
	tierPrefixName
	tierFallback
)

// Skill describes a selectable skill in the popup.
type Skill struct {
	// ID is the skill catalog ID (skill file path) used to load the skill
	// content on selection.
	ID          string
	Name        string
	Description string
}

// SelectionMsg is sent when a skill is selected.
type SelectionMsg struct {
	Value    Skill
	KeepOpen bool // If true, insert without closing.
}

// ClosedMsg is sent when the popup is closed.
type ClosedMsg struct{}

// ItemsLoadedMsg is sent by the caller once it has loaded the invocable
// skill list from the workspace catalog.
type ItemsLoadedMsg struct {
	Skills []Skill
}

// KeyMap defines the key bindings for the skill selector. The bindings
// mirror the file completions popup.
type KeyMap struct {
	Down,
	Up,
	Select,
	Cancel key.Binding
	DownInsert,
	UpInsert key.Binding
}

// DefaultKeyMap returns the default key bindings for the skill selector.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("down", "move down"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("up", "move up"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter", "tab", "ctrl+y"),
			key.WithHelp("enter", "select"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc", "alt+esc"),
			key.WithHelp("esc", "cancel"),
		),
		DownInsert: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "insert next"),
		),
		UpInsert: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "insert previous"),
		),
	}
}

// SkillSelector represents the skill-mention popup component.
type SkillSelector struct {
	// Popup dimensions
	width  int
	height int

	// State
	open  bool
	query string

	// Key bindings
	keyMap KeyMap

	// List component
	list *list.FilterableList

	// Styling
	normalStyle  lipgloss.Style
	focusedStyle lipgloss.Style
	matchStyle   lipgloss.Style

	allItems []list.FilterableItem
	filtered []list.FilterableItem
}

// New creates a new skill selector component.
func New(normalStyle, focusedStyle, matchStyle lipgloss.Style) *SkillSelector {
	l := list.NewFilterableList()
	l.SetGap(0)
	l.SetReverse(true)

	return &SkillSelector{
		keyMap:       DefaultKeyMap(),
		list:         l,
		normalStyle:  normalStyle,
		focusedStyle: focusedStyle,
		matchStyle:   matchStyle,
	}
}

// IsOpen returns whether the popup is open.
func (s *SkillSelector) IsOpen() bool {
	return s.open
}

// Size returns the visible size of the popup.
func (s *SkillSelector) Size() (width, height int) {
	visible := len(s.filtered)
	return s.width, min(visible, s.height)
}

// HasItems returns whether there are visible items.
func (s *SkillSelector) HasItems() bool {
	return len(s.filtered) > 0
}

// FilteredSkills returns the skills currently visible after filtering, in
// list order.
func (s *SkillSelector) FilteredSkills() []Skill {
	out := make([]Skill, 0, len(s.filtered))
	for _, item := range s.filtered {
		ci, ok := item.(*completions.CompletionItem)
		if !ok {
			continue
		}
		if skill, ok := ci.Value().(Skill); ok {
			out = append(out, skill)
		}
	}
	return out
}

// SetItems sets the skills and opens the popup. The list is supplied
// synchronously by the caller; the component performs no async loading.
func (s *SkillSelector) SetItems(skills []Skill) {
	items := make([]list.FilterableItem, 0, len(skills))
	for _, skill := range skills {
		items = append(items, newItem(skill, s.normalStyle, s.focusedStyle, s.matchStyle))
	}

	s.open = true
	s.query = ""
	s.allItems = items
	s.filtered = append([]list.FilterableItem(nil), items...)
	s.list.SetItems(s.filtered...)
	s.list.SetFilter("")
	s.list.Focus()

	s.width = maxWidth
	s.height = ordered.Clamp(len(items), int(minHeight), int(maxHeight))
	s.list.SetSize(s.width, s.height)
	s.list.SelectFirst()
	s.list.ScrollToSelected()
	s.updateSize()
}

// Close closes the skill selector popup.
func (s *SkillSelector) Close() {
	s.open = false
}

// Filter filters the skills with the given query.
func (s *SkillSelector) Filter(query string) {
	if !s.open {
		return
	}

	if query == s.query {
		return
	}

	s.query = query
	s.applyFilter(query)
	s.updateSize()
}

// Update handles key events for the skill selector.
func (s *SkillSelector) Update(msg tea.KeyPressMsg) (tea.Msg, bool) {
	if !s.open {
		return nil, false
	}

	switch {
	case key.Matches(msg, s.keyMap.Up):
		s.selectPrev()
		return nil, true

	case key.Matches(msg, s.keyMap.Down):
		s.selectNext()
		return nil, true

	case key.Matches(msg, s.keyMap.UpInsert):
		s.selectPrev()
		return s.selectCurrent(true), true

	case key.Matches(msg, s.keyMap.DownInsert):
		s.selectNext()
		return s.selectCurrent(true), true

	case key.Matches(msg, s.keyMap.Select):
		return s.selectCurrent(false), true

	case key.Matches(msg, s.keyMap.Cancel):
		s.Close()
		return ClosedMsg{}, true
	}

	return nil, false
}

// Render renders the skill selector popup.
func (s *SkillSelector) Render() string {
	if !s.open {
		return ""
	}

	if len(s.filtered) == 0 {
		return ""
	}

	return s.list.List.Render()
}

func (s *SkillSelector) applyFilter(query string) {
	if query == "" {
		s.filtered = append([]list.FilterableItem(nil), s.allItems...)
		s.list.SetItems(s.filtered...)
		return
	}

	s.list.SetItems(s.allItems...)
	s.list.SetFilter(query)
	raw := s.list.FilteredItems()
	filtered := make([]list.FilterableItem, 0, len(raw))
	for _, item := range raw {
		filterable, ok := item.(list.FilterableItem)
		if !ok {
			continue
		}
		filtered = append(filtered, filterable)
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	slices.SortStableFunc(filtered, func(a, b list.FilterableItem) int {
		return skillTier(nameOf(a), queryLower) - skillTier(nameOf(b), queryLower)
	})
	s.filtered = filtered
	s.list.SetItems(s.filtered...)
}

func nameOf(item list.FilterableItem) string {
	ci, ok := item.(*completions.CompletionItem)
	if !ok {
		return ""
	}
	skill, ok := ci.Value().(Skill)
	if !ok {
		return ""
	}
	return skill.Name
}

func skillTier(nameLower, queryLower string) int {
	if queryLower == "" {
		return tierFallback
	}
	nameLower = strings.ToLower(nameLower)
	switch {
	case nameLower == queryLower:
		return tierExactName
	case strings.HasPrefix(nameLower, queryLower):
		return tierPrefixName
	default:
		return tierFallback
	}
}

func (s *SkillSelector) selectPrev() {
	if len(s.filtered) == 0 {
		return
	}
	if !s.list.SelectPrev() {
		s.list.WrapToEnd()
	}
	s.list.ScrollToSelected()
}

func (s *SkillSelector) selectNext() {
	if len(s.filtered) == 0 {
		return
	}
	if !s.list.SelectNext() {
		s.list.WrapToStart()
	}
	s.list.ScrollToSelected()
}

// selectCurrent returns the message for the currently selected skill.
func (s *SkillSelector) selectCurrent(keepOpen bool) tea.Msg {
	if len(s.filtered) == 0 {
		return nil
	}

	selected := s.list.Selected()
	if selected < 0 || selected >= len(s.filtered) {
		return nil
	}

	ci, ok := s.filtered[selected].(*completions.CompletionItem)
	if !ok {
		return nil
	}

	item, ok := ci.Value().(Skill)
	if !ok {
		return nil
	}

	if !keepOpen {
		s.open = false
	}

	return SelectionMsg{Value: item, KeepOpen: keepOpen}
}

func (s *SkillSelector) updateSize() {
	items := s.filtered
	start, end := s.list.VisibleItemIndices()
	width := 0
	for i := start; i <= end; i++ {
		item := s.list.ItemAt(i)
		if item == nil {
			continue
		}
		s := item.(interface{ Text() string }).Text()
		width = max(width, ansi.StringWidth(s))
	}
	s.width = ordered.Clamp(width+2, int(minWidth), int(maxWidth))
	s.height = ordered.Clamp(len(items), int(minHeight), int(maxHeight))
	s.list.SetSize(s.width, s.height)
	s.list.SelectFirst()
	s.list.ScrollToSelected()
}

// newItem builds a list item for a skill, reusing the shared completion
// item rendering (fuzzy match highlighting, truncation, focus styles). The
// name is rendered in bold (via embedded ANSI) and filtered separately from
// the description, so typing filters by name only.
func newItem(skill Skill, normalStyle, focusedStyle, matchStyle lipgloss.Style) list.FilterableItem {
	text := lipgloss.NewStyle().Bold(true).Render(skill.Name)
	if skill.Description != "" {
		text += "  " + skill.Description
	}
	item := completions.NewCompletionItem(text, skill, normalStyle, focusedStyle, matchStyle)
	item.SetNameOnly(skill.Name)
	return item
}
