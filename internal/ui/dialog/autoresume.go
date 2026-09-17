package dialog

import (
	"fmt"
	"strconv"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/crush/internal/ui/common"
	uv "github.com/charmbracelet/ultraviolet"
)

const (
	AutoResumeID             = "auto_resume"
	autoResumeDialogMaxWidth = 55
)

// AutoResume is a dialog for configuring auto-resume settings.
type AutoResume struct {
	com       *common.Common
	help      help.Model
	input     textinput.Model
	model     string // "large" or "small"
	threshold int

	keyMap struct {
		Select key.Binding
		Close  key.Binding
		Tab    key.Binding
	}
}

var _ Dialog = (*AutoResume)(nil)

// NewAutoResume creates a new auto-resume config dialog.
func NewAutoResume(com *common.Common) *AutoResume {
	initialModel := "large"
	initialThreshold := "100"

	if cfg := com.Config(); cfg != nil && cfg.Options != nil {
		initialModel = cfg.Options.AutoResumeModel
		if initialModel == "" {
			initialModel = "large"
		}
		initialThreshold = strconv.Itoa(cfg.Options.AutoResumeThreshold)
		if cfg.Options.AutoResumeThreshold == 0 {
			initialThreshold = "100"
		}
	}

	r := &AutoResume{
		com:   com,
		model: initialModel,
	}

	h := help.New()
	h.Styles = com.Styles.DialogHelpStyles()
	r.help = h

	r.input = textinput.New()
	r.input.SetVirtualCursor(false)
	r.input.Placeholder = "100"
	r.input.SetStyles(com.Styles.TextInput)
	r.input.Focus()
	r.input.SetValue(initialThreshold)
	r.input.Validate = func(s string) error {
		if s == "" {
			return nil
		}
		for _, c := range s {
			if c < '0' || c > '9' {
				return fmt.Errorf("only digits allowed")
			}
		}
		return nil
	}

	r.keyMap.Select = key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "save"),
	)
	r.keyMap.Tab = key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "toggle model"),
	)
	r.keyMap.Close = CloseKey

	return r
}

// ID implements Dialog.
func (r *AutoResume) ID() string { return AutoResumeID }

// HandleMsg implements Dialog.
func (r *AutoResume) HandleMsg(msg tea.Msg) Action {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, r.keyMap.Close):
			return ActionClose{}
		case key.Matches(msg, r.keyMap.Tab):
			if r.model == "large" {
				r.model = "small"
			} else {
				r.model = "large"
			}
		case key.Matches(msg, r.keyMap.Select):
			threshold := 100
			if v := r.input.Value(); v != "" {
				if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 100 {
					threshold = n
				}
			}
			return ActionAutoResumeConfig{
				Threshold: threshold,
				Model:     r.model,
			}
		default:
			var cmd tea.Cmd
			r.input, cmd = r.input.Update(msg)
			return ActionCmd{Cmd: cmd}
		}
	}
	return nil
}

// Cursor implements Dialog.
func (r *AutoResume) Cursor() *tea.Cursor {
	cur := r.input.Cursor()
	if cur == nil {
		return nil
	}
	// The threshold input is the second content line in the dialog
	// (after the title and the model line). Offset Y by 1 so the
	// cursor appears on the correct line.
	cur.Y += 1
	// The threshold line includes a label before the input value.
	// Offset X by the label length so the cursor appears inside the
	// input field, not on the label text.
	cur.X += len("Threshold (%): ")
	return InputCursor(r.com.Styles, cur)
}

// Draw implements Dialog.
func (r *AutoResume) Draw(scr uv.Screen, area uv.Rectangle) *tea.Cursor {
	t := r.com.Styles
	width := max(0, min(autoResumeDialogMaxWidth, area.Dx()-t.Dialog.View.GetHorizontalBorderSize()))
	innerWidth := width - t.Dialog.View.GetHorizontalFrameSize()

	r.input.SetWidth(dialogInputTextWidth(t, r.input, innerWidth))

	rc := NewRenderContext(t, width)
	rc.Title = "Auto Resume Session"

	modelLine := "Resume Model:   "
	if r.model == "large" {
		modelLine += t.Dialog.SelectedItem.Render(" Large ") + "   " + t.Dialog.NormalItem.Render(" Small ")
	} else {
		modelLine += t.Dialog.NormalItem.Render(" Large ") + "   " + t.Dialog.SelectedItem.Render(" Small ")
	}
	rc.AddPart(t.Dialog.NormalItem.Render(modelLine))

	thresholdLine := "Threshold (%): " + r.input.View()
	rc.AddPart(t.Dialog.InputPrompt.Render(thresholdLine))

	rc.Help = "tab: toggle model  |  enter: save  |  esc: cancel"

	view := rc.Render()

	cur := r.Cursor()
	DrawCenterCursor(scr, area, view, cur)
	return cur
}