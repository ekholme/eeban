package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/ekholme/eeban/internal/domain"
)

// editField identifies which field of the edit form has focus.
type editField int

const (
	fieldTitle editField = iota
	fieldBody
	fieldPriority
	fieldDueDate
	fieldCount
)

// editForm is the nested model for editing a card's title, body, priority,
// and due date.
type editForm struct {
	cardID   int64
	title    textinput.Model
	body     textarea.Model
	priority int
	dueDate  textinput.Model
	focus    editField
}

// newEditForm builds a form pre-populated from card, focused on the title.
func newEditForm(card domain.Card) editForm {
	title := textinput.New()
	title.Prompt = ""
	title.Placeholder = "Title"
	title.CharLimit = 200
	title.SetValue(card.Title)

	body := textarea.New()
	body.Placeholder = "Description"
	body.ShowLineNumbers = false
	body.SetValue(card.Body)

	due := textinput.New()
	due.Prompt = ""
	due.Placeholder = "YYYY-MM-DD"
	due.CharLimit = 10
	if card.DueDate != nil {
		due.SetValue(*card.DueDate)
	}

	f := editForm{
		cardID:   card.ID,
		title:    title,
		body:     body,
		priority: card.Priority,
		dueDate:  due,
		focus:    fieldTitle,
	}
	f.focusCurrent()
	return f
}

// focusCurrent focuses whichever field f.focus names and blurs the rest.
func (f *editForm) focusCurrent() {
	f.title.Blur()
	f.body.Blur()
	f.dueDate.Blur()
	switch f.focus {
	case fieldTitle:
		f.title.Focus()
	case fieldBody:
		f.body.Focus()
	case fieldDueDate:
		f.dueDate.Focus()
	}
}

// setWidth sizes the text fields to fit within a pane of width w.
func (f *editForm) setWidth(w int) {
	f.title.Width = w
	f.body.SetWidth(w)
	f.dueDate.Width = w
}

// formAction reports what a form keypress should do to the parent model.
type formAction int

const (
	formNone formAction = iota
	formSubmit
	formCancel
)

// update routes a keypress to the focused field, or handles form-level
// navigation and submission.
func (f editForm) update(msg tea.KeyMsg) (editForm, tea.Cmd, formAction) {
	switch msg.String() {
	case "esc":
		return f, nil, formCancel
	case "ctrl+s":
		return f, nil, formSubmit
	case "tab":
		f.focus = (f.focus + 1) % fieldCount
		f.focusCurrent()
		return f, nil, formNone
	case "shift+tab":
		f.focus = (f.focus - 1 + fieldCount) % fieldCount
		f.focusCurrent()
		return f, nil, formNone
	}

	if f.focus == fieldPriority {
		switch msg.String() {
		case "left", "h":
			f.priority = clamp(f.priority-1, domain.PriorityNone, domain.PriorityHigh)
		case "right", "l":
			f.priority = clamp(f.priority+1, domain.PriorityNone, domain.PriorityHigh)
		}
		return f, nil, formNone
	}

	var cmd tea.Cmd
	switch f.focus {
	case fieldTitle:
		f.title, cmd = f.title.Update(msg)
	case fieldBody:
		f.body, cmd = f.body.Update(msg)
	case fieldDueDate:
		f.dueDate, cmd = f.dueDate.Update(msg)
	}
	return f, cmd, formNone
}

// dueDatePtr returns the trimmed due-date value, or nil when it's blank.
func (f editForm) dueDatePtr() *string {
	v := strings.TrimSpace(f.dueDate.Value())
	if v == "" {
		return nil
	}
	return &v
}
