package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ekholme/eeban/internal/domain"
)

// openLabelPicker opens the label overlay for the selected card.
func (m Model) openLabelPicker() (tea.Model, tea.Cmd) {
	if m.svc == nil {
		return m, nil
	}
	if _, ok := m.selectedCard(); !ok {
		return m, nil
	}
	m.labelPicker = true
	m.labelNaming = false
	m.labelCursor = clamp(m.labelCursor, 0, max(len(m.board.Labels)-1, 0))
	return m, nil
}

// updateLabelPicker handles keys while the label overlay is open.
func (m Model) updateLabelPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.labelNaming {
		switch msg.Type {
		case tea.KeyEsc:
			m.labelNaming = false
			return m, nil
		case tea.KeyEnter:
			name := strings.TrimSpace(m.labelNameInput.Value())
			m.labelNaming = false
			if name == "" {
				return m, nil
			}
			return m, m.createLabelCmd(name, nextLabelColor(len(m.board.Labels)))
		}
		var cmd tea.Cmd
		m.labelNameInput, cmd = m.labelNameInput.Update(msg)
		return m, cmd
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Labels), key.Matches(msg, m.keys.Back):
		m.labelPicker = false
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.labelCursor = clamp(m.labelCursor-1, 0, max(len(m.board.Labels)-1, 0))
	case key.Matches(msg, m.keys.Down):
		m.labelCursor = clamp(m.labelCursor+1, 0, max(len(m.board.Labels)-1, 0))
	case key.Matches(msg, m.keys.New):
		ti := textinput.New()
		ti.Placeholder = "Label name"
		ti.CharLimit = 40
		ti.Focus()
		m.labelNameInput = ti
		m.labelNaming = true
		return m, textinput.Blink
	case key.Matches(msg, m.keys.Delete):
		if l, ok := m.highlightedLabel(); ok {
			m.confirm = &confirmState{
				prompt: fmt.Sprintf("Delete label %q from every card?", l.Name),
				action: m.deleteLabelCmd(l.ID),
			}
		}
		return m, nil
	case msg.String() == " " || msg.String() == "enter":
		l, ok := m.highlightedLabel()
		card, cok := m.selectedCard()
		if !ok || !cok {
			return m, nil
		}
		return m, m.setCardLabelCmd(card.ID, l.ID, !card.HasLabel(l.ID))
	case msg.String() == "f":
		if l, ok := m.highlightedLabel(); ok {
			m.filterLabel = l.ID
			m.labelPicker = false
			m.clampToFilter()
		}
		return m, nil
	}
	return m, nil
}

func (m Model) highlightedLabel() (domain.Label, bool) {
	if m.labelCursor < 0 || m.labelCursor >= len(m.board.Labels) {
		return domain.Label{}, false
	}
	return m.board.Labels[m.labelCursor], true
}

// labelColor resolves a label's stored colour code to a lipgloss colour,
// falling back to the accent when unset.
func labelColor(l domain.Label) lipgloss.TerminalColor {
	if l.Color == "" {
		return lipgloss.Color("63")
	}
	return lipgloss.Color(l.Color)
}

// renderLabelDots renders one coloured dot per label, for the compact card view.
func (m Model) renderLabelDots(labels []domain.Label) string {
	if len(labels) == 0 {
		return ""
	}
	var b strings.Builder
	for _, l := range labels {
		b.WriteString(m.styles.LabelChip.Foreground(labelColor(l)).Render("●"))
	}
	return b.String()
}

// renderLabelChips renders "●name" per label, for the detail pane and archive.
func (m Model) renderLabelChips(labels []domain.Label) string {
	if len(labels) == 0 {
		return ""
	}
	parts := make([]string, len(labels))
	for i, l := range labels {
		parts[i] = m.styles.LabelChip.Foreground(labelColor(l)).Render("● " + l.Name)
	}
	return strings.Join(parts, "  ")
}

// labelPickerView renders the label overlay: every board label with a
// checkbox reflecting the selected card's membership.
func (m Model) labelPickerView() string {
	var b strings.Builder
	title := "labels"
	if c, ok := m.selectedCard(); ok {
		title = "labels · " + truncateText(c.Title, 30)
	}
	b.WriteString(m.styles.HelpHeading.Render(title))
	b.WriteByte('\n')

	card, _ := m.selectedCard()
	if len(m.board.Labels) == 0 {
		b.WriteString(m.styles.DetailDim.Render("No labels yet — press n to add one."))
		b.WriteByte('\n')
	}
	for i, l := range m.board.Labels {
		box := "[ ]"
		if card.HasLabel(l.ID) {
			box = "[x]"
		}
		name := m.styles.LabelChip.Foreground(labelColor(l)).Render("● " + l.Name)
		row := fmt.Sprintf("%s %s", box, name)
		if i == m.labelCursor {
			b.WriteString(m.styles.OverlayRowActive.Render("> " + row))
		} else {
			b.WriteString(m.styles.OverlayRow.Render(row))
		}
		b.WriteByte('\n')
	}

	if m.labelNaming {
		b.WriteString("\n" + m.styles.Prompt.Render("New label: ") + m.labelNameInput.View() + "\n")
	}

	b.WriteString(m.styles.DetailDim.Render("\nspace toggle · n new · d delete · f filter by label · t/esc close"))

	box := m.styles.HelpOverlay.Render(b.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
