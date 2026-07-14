package main

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateList state = iota
	stateActions
	stateAdd
	stateEdit
	stateEditField
	stateConfirm
	stateDetails
)

const (
	fieldName = iota
	fieldUser
	fieldHost
	fieldPort
	fieldKeyPath
)

var (
	red        = lipgloss.Color("196")
	textGray   = lipgloss.Color("251")
	borderGray = lipgloss.Color("239")

	bannerStyle  = lipgloss.NewStyle().Foreground(red)
	headingStyle = lipgloss.NewStyle().Bold(true).Foreground(red)
	dimStyle     = lipgloss.NewStyle().Foreground(textGray)
	errStyle     = lipgloss.NewStyle().Foreground(red)
	msgStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	helpStyle    = lipgloss.NewStyle().Foreground(textGray).MarginTop(1)

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(red).
			Padding(1, 2)

	fieldLabel = lipgloss.NewStyle().
			Foreground(textGray).
			Bold(true)

	profileBase = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 2)

	actionBase = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(red).
			Padding(0, 2)

	nameBase  = lipgloss.NewStyle().Bold(true)
	nameRed   = nameBase.Foreground(red)
	nameWhite = nameBase.Foreground(lipgloss.Color("255"))

	actionSel = lipgloss.NewStyle().Bold(true).Foreground(red)

	bannerArt = `   __________ __  __     __    
  / ___/ ___// / / /__  / /___ 
  \__ \\__ \/ /_/ / _ \/ / __ \
 ___/ /__/ / __  /  __/ / /_/ /
/____/____/_/ /_/\___/_/ .___/ 
                      /_/      `

	actions = []string{"[C] Connect", "[E] Edit", "[D] Delete", "[V] Details"}

	stepLabels = []string{
		"Profile Name",
		"Username",
		"Host",
		"Port",
		"Key Path (optional)",
	}
)

type model struct {
	state              state
	profiles           []Profile
	cursor             int
	width, height      int
	err, msg           string
	actionCursor       int

	addStep            int
	addInput           textinput.Model
	addProfile         Profile

	editField          int
	editInput          textinput.Model
	editProfile        Profile
	editIndex          int

	confirmMsg         string
	confirmDelete      bool

	detailProfile      Profile
}

func initialModel(profiles []Profile) model {
	ti := textinput.New()
	ti.Focus()
	ti.Width = 40
	ti.Prompt = "> "
	ti.Placeholder = "profile name"

	return model{
		state:    stateList,
		profiles: profiles,
		addInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.handleListKey(msg)
		case stateActions:
			return m.handleActionKey(msg)
		case stateAdd:
			return m.handleAddKey(msg)
		case stateEdit:
			return m.handleEditKey(msg)
		case stateEditField:
			return m.handleEditFieldKey(msg)
		case stateConfirm:
			return m.handleConfirmKey(msg)
		case stateDetails:
			return m.handleDetailsKey(msg)
		}

	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			return m.handleClick(msg)
		}
	}

	return m, nil
}

func (m model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.err = ""
		}
	case "down", "j":
		if m.cursor < len(m.profiles)-1 {
			m.cursor++
			m.err = ""
		}
	case "enter", " ":
		if len(m.profiles) > 0 {
			m.state = stateActions
			m.actionCursor = 0
		}
	case "a":
		m.state = stateAdd
		m.addStep = 0
		m.addProfile = Profile{}
		m.addInput = textinput.New()
		m.addInput.Focus()
		m.addInput.Width = 40
		m.addInput.Prompt = "> "
		m.addInput.Placeholder = "profile name"
		m.addInput.CharLimit = 100
		m.err = ""
		return m, textinput.Blink
	}
	return m, nil
}

func (m model) handleActionKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "c":
		p := m.profiles[m.cursor]
		p = p.UpdatedNow()
		m.profiles[m.cursor] = p
		SaveProfiles(m.profiles)
		if err := connectSSH(p); err != nil {
			m.err = "Connection error: " + err.Error()
		} else {
			m.msg = "Connecting to " + p.Name + "..."
		}
		m.state = stateList
	case "e":
		m.state = stateEdit
		m.editField = 0
		m.editProfile = m.profiles[m.cursor]
		m.editIndex = m.cursor
		m.err = ""
	case "d":
		m.state = stateConfirm
		m.confirmMsg = "Delete profile \"" + m.profiles[m.cursor].Name + "\"?"
		m.confirmDelete = true
	case "v":
		m.state = stateDetails
		m.detailProfile = m.profiles[m.cursor]
	case "esc":
		m.state = stateList
		m.err = ""
	case "left", "h":
		m.actionCursor = (m.actionCursor + 3) % 4
	case "right", "l":
		m.actionCursor = (m.actionCursor + 1) % 4
	case "enter":
		switch m.actionCursor {
		case 0:
			return m.handleActionKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
		case 1:
			return m.handleActionKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
		case 2:
			return m.handleActionKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
		case 3:
			return m.handleActionKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
		}
	}
	return m, nil
}

func (m model) handleAddKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addStep <= 4 {
		switch msg.String() {
		case "esc":
			m.state = stateList
			return m, nil
		case "enter":
			val := m.addInput.Value()
			switch m.addStep {
			case 0:
				if val == "" {
					m.err = "Name cannot be empty"
					return m, nil
				}
				m.addProfile.Name = val
				m.addStep = 1
				m.addInput.SetValue("")
				m.addInput.Placeholder = "e.g. root, admin, deploy"
			case 1:
				if val == "" {
					m.err = "User cannot be empty"
					return m, nil
				}
				m.addProfile.User = val
				m.addStep = 2
				m.addInput.SetValue("")
				m.addInput.Placeholder = "e.g. 192.168.1.100 or server.example.com"
			case 2:
				if val == "" {
					m.err = "Host cannot be empty"
					return m, nil
				}
				m.addProfile.Host = val
				m.addStep = 3
				m.addInput.SetValue("22")
				m.addInput.Placeholder = "default: 22"
			case 3:
				if val == "" {
					val = "22"
				}
				port, err := strconv.Atoi(val)
				if err != nil || port < 1 || port > 65535 {
					m.err = "Invalid port (1-65535)"
					return m, nil
				}
				m.addProfile.Port = port
				m.addStep = 4
				m.addInput.SetValue("")
				m.addInput.Placeholder = "leave empty if not needed"
			case 4:
				m.addProfile.KeyPath = val
				m.addStep = 5
				m.addInput.Blur()
				return m, nil
			}
			m.err = ""
			return m, textinput.Blink
		default:
			var cmd tea.Cmd
			m.addInput, cmd = m.addInput.Update(msg)
			return m, cmd
		}
	} else {
		switch msg.String() {
		case "enter":
			for _, p := range m.profiles {
				if p.Name == m.addProfile.Name {
					m.err = "A profile with this name already exists"
					return m, nil
				}
			}
			m.profiles = append(m.profiles, m.addProfile)
			SaveProfiles(m.profiles)
			m.cursor = len(m.profiles) - 1
			m.msg = "Profile \"" + m.addProfile.Name + "\" added"
			m.state = stateList
		case "esc":
			m.state = stateList
		}
		return m, nil
	}
}

func (m model) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateList
	case "up", "k":
		if m.editField > 0 {
			m.editField--
		}
	case "down", "j":
		if m.editField < fieldKeyPath {
			m.editField++
		}
	case "enter":
		m.state = stateEditField
		m.editInput = textinput.New()
		m.editInput.Focus()
		m.editInput.Width = 40
		m.editInput.Prompt = ""
		m.editInput.CharLimit = 200
		switch m.editField {
		case fieldName:
			m.editInput.SetValue(m.editProfile.Name)
		case fieldUser:
			m.editInput.SetValue(m.editProfile.User)
		case fieldHost:
			m.editInput.SetValue(m.editProfile.Host)
		case fieldPort:
			m.editInput.SetValue(strconv.Itoa(m.editProfile.Port))
		case fieldKeyPath:
			m.editInput.SetValue(m.editProfile.KeyPath)
		}
		return m, textinput.Blink
	case "ctrl+s":
		for i, p := range m.profiles {
			if i != m.editIndex && p.Name == m.editProfile.Name {
				m.err = "A profile with this name already exists"
				return m, nil
			}
		}
		m.profiles[m.editIndex] = m.editProfile
		SaveProfiles(m.profiles)
		m.msg = "Profile \"" + m.editProfile.Name + "\" updated"
		m.err = ""
		m.state = stateList
	}
	return m, nil
}

func (m model) handleEditFieldKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateEdit
		m.err = ""
		return m, nil
	case "enter":
		val := m.editInput.Value()
		switch m.editField {
		case fieldName:
			if val == "" {
				m.err = "Name cannot be empty"
				return m, nil
			}
			m.editProfile.Name = val
		case fieldUser:
			if val == "" {
				m.err = "User cannot be empty"
				return m, nil
			}
			m.editProfile.User = val
		case fieldHost:
			if val == "" {
				m.err = "Host cannot be empty"
				return m, nil
			}
			m.editProfile.Host = val
		case fieldPort:
			if val == "" {
				val = "22"
			}
			port, err := strconv.Atoi(val)
			if err != nil || port < 1 || port > 65535 {
				m.err = "Invalid port (1-65535)"
				return m, nil
			}
			m.editProfile.Port = port
		case fieldKeyPath:
			m.editProfile.KeyPath = val
		}
		m.err = ""
		m.state = stateEdit
		return m, nil
	default:
		var cmd tea.Cmd
		m.editInput, cmd = m.editInput.Update(msg)
		return m, cmd
	}
}

func (m model) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		if m.confirmDelete {
			m.profiles = append(m.profiles[:m.cursor], m.profiles[m.cursor+1:]...)
			if m.cursor >= len(m.profiles) && m.cursor > 0 {
				m.cursor--
			}
			SaveProfiles(m.profiles)
			m.msg = "Profile deleted"
		}
		m.state = stateList
	case "n", "N", "esc":
		m.state = stateList
	}
	return m, nil
}

func (m model) handleDetailsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.state = stateActions
	return m, nil
}

func (m model) handleClick(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.state != stateList && m.state != stateActions {
		return m, nil
	}
	y := msg.Y
	if y < 3 || len(m.profiles) == 0 {
		return m, nil
	}

	idx := (y - 3) / 5
	if idx >= len(m.profiles) {
		return m, nil
	}

	if idx == m.cursor {
		if m.state == stateList {
			m.state = stateActions
			m.actionCursor = 0
		} else {
			m.state = stateList
		}
	} else {
		m.cursor = idx
		m.state = stateList
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateList, stateActions:
		return m.listView()
	case stateAdd:
		return m.addView()
	case stateEdit, stateEditField:
		return m.editView()
	case stateConfirm:
		return m.confirmView()
	case stateDetails:
		return m.detailsView()
	}
	return ""
}

func (m model) bannerView() string {
	return bannerStyle.Render(bannerArt) + "\n"
}

func (m model) listView() string {
	var b strings.Builder
	b.Grow(1024)

	b.WriteString(m.bannerView())
	b.WriteByte('\n')

	if len(m.profiles) == 0 {
		b.WriteString(dimStyle.Render("  No profiles yet."))
		b.WriteByte('\n')
		b.WriteString(dimStyle.Render("  Press 'a' to add a new profile."))
	} else {
		maxW := m.width - 6
		if maxW < 30 {
			maxW = 30
		}
		if maxW < 50 {
			maxW = 50
		}

		for i, p := range m.profiles {
			sel := i == m.cursor

			borderCol := borderGray
			nameSty := nameWhite
			if sel {
				borderCol = red
				nameSty = nameRed
			}

			card := profileBase.BorderForeground(borderCol).Width(maxW)
			b.WriteString(card.Render(nameSty.Render(p.Name) + "\n  " + p.HostInfo()))
			b.WriteByte('\n')

			if sel && m.state == stateActions {
				rendered := make([]string, 4)
				for ai, a := range actions {
					if ai == m.actionCursor {
						rendered[ai] = actionSel.Render(a)
					} else {
						rendered[ai] = a
					}
				}
				b.WriteString(actionBase.Width(maxW).Render(strings.Join(rendered, "  ")))
				b.WriteByte('\n')
			}
		}
	}

	b.WriteByte('\n')
	if m.err != "" {
		b.WriteString(errStyle.Render("ERROR: " + m.err))
		b.WriteString("\n\n")
	}
	if m.msg != "" {
		b.WriteString(msgStyle.Render(m.msg))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("  [a] Add  [↑↓] Nav  [Enter] Actions  [q] Quit"))
	return b.String()
}

func (m model) addView() string {
	var b strings.Builder
	b.Grow(512)

	b.WriteString(m.bannerView())
	b.WriteString(headingStyle.Render("Add Profile"))
	b.WriteString("\n\n")

	maxW := m.width - 8
	if maxW < 30 {
		maxW = 30
	}
	if maxW < 44 {
		maxW = 44
	}

	if m.addStep <= 4 {
		label := "Step " + strconv.Itoa(m.addStep+1) + "/5: " + stepLabels[m.addStep]
		content := headingStyle.Render(label) + "\n\n"
		content += m.addInput.View() + "\n\n"
		content += dimStyle.Render("[Enter] Next  [Esc] Cancel")
		b.WriteString(dialogStyle.Width(maxW).Render(content))
	} else {
		keyVal := m.addProfile.KeyPath
		if keyVal == "" {
			keyVal = dimStyle.Render("(none)")
		}
		content := headingStyle.Render("Review & Save") + "\n\n"
		content += "  " + fieldLabel.Render("Name:") + "  " + m.addProfile.Name + "\n"
		content += "  " + fieldLabel.Render("User:") + "  " + m.addProfile.User + "\n"
		content += "  " + fieldLabel.Render("Host:") + "  " + m.addProfile.Host + "\n"
		content += "  " + fieldLabel.Render("Port:") + "  " + strconv.Itoa(m.addProfile.Port) + "\n"
		content += "  " + fieldLabel.Render("Key:") + "  " + keyVal + "\n"
		content += "\n" + dimStyle.Render("[Enter] Save  [Esc] Cancel")
		b.WriteString(dialogStyle.Width(maxW).Render(content))
	}

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errStyle.Render("ERROR: " + m.err))
	}

	return b.String()
}

func (m model) editView() string {
	var b strings.Builder
	b.Grow(512)

	b.WriteString(m.bannerView())
	b.WriteString(headingStyle.Render("Edit Profile"))
	b.WriteString("\n\n")

	fields := [5][2]string{
		{"Name:", m.editProfile.Name},
		{"User:", m.editProfile.User},
		{"Host:", m.editProfile.Host},
		{"Port:", strconv.Itoa(m.editProfile.Port)},
		{"Key:", m.editProfile.KeyPath},
	}

	maxW := m.width - 8
	if maxW < 30 {
		maxW = 30
	}
	if maxW < 50 {
		maxW = 50
	}

	var content strings.Builder
	content.Grow(256)

	for i, f := range fields {
		if i == m.editField && m.state == stateEdit {
			content.WriteString("> ")
		} else {
			content.WriteString("  ")
		}

		content.WriteString(fieldLabel.Render(f[0]))
		content.WriteByte(' ')
		content.WriteByte(' ')

		if i == m.editField && m.state == stateEditField {
			content.WriteString(m.editInput.View())
		} else if f[1] == "" {
			content.WriteString(dimStyle.Render("(empty)"))
		} else {
			content.WriteString(f[1])
		}
		content.WriteByte('\n')
	}

	content.WriteByte('\n')
	if m.state == stateEdit {
		content.WriteString(dimStyle.Render("[↑↓] Nav  [Enter] Edit Field  [Ctrl+S] Save  [Esc] Back"))
	} else {
		content.WriteString(dimStyle.Render("[Enter] Confirm  [Esc] Cancel"))
	}

	b.WriteString(dialogStyle.Width(maxW).Render(content.String()))

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errStyle.Render("ERROR: " + m.err))
	}

	return b.String()
}

func (m model) confirmView() string {
	var b strings.Builder
	b.Grow(256)

	b.WriteString(m.bannerView())
	b.WriteString(headingStyle.Render("Confirm"))
	b.WriteString("\n\n")

	maxW := m.width - 8
	if maxW < 30 {
		maxW = 30
	}
	if maxW < 44 {
		maxW = 44
	}

	content := m.confirmMsg + "\n\n" + dimStyle.Render("[Y]es  [N]o  [Esc] Cancel")
	b.WriteString(dialogStyle.Width(maxW).Render(content))

	return b.String()
}

func (m model) detailsView() string {
	var b strings.Builder
	b.Grow(512)

	p := m.detailProfile
	b.WriteString(m.bannerView())
	b.WriteString(headingStyle.Render("Profile Details"))
	b.WriteString("\n\n")

	maxW := m.width - 8
	if maxW < 30 {
		maxW = 30
	}
	if maxW < 44 {
		maxW = 44
	}

	last := p.LastAccess
	if last == "" {
		last = dimStyle.Render("never")
	}
	keyVal := p.KeyPath
	if keyVal == "" {
		keyVal = dimStyle.Render("(none)")
	}

	content := headingStyle.Render(p.Name) + "\n\n"
	content += "  " + fieldLabel.Render("User:") + "  " + p.User + "\n"
	content += "  " + fieldLabel.Render("Host:") + "  " + p.Host + "\n"
	content += "  " + fieldLabel.Render("Port:") + "  " + strconv.Itoa(p.Port) + "\n"
	content += "  " + fieldLabel.Render("Key:") + "  " + keyVal + "\n"
	content += "  " + fieldLabel.Render("Connect:") + "  " + p.HostInfo() + "\n"
	content += "  " + fieldLabel.Render("Last access:") + "  " + last + "\n"
	content += "\n" + dimStyle.Render("[Any key] Back")

	b.WriteString(dialogStyle.Width(maxW).Render(content))

	return b.String()
}

func connectSSH(p Profile) error {
	args := []string{"/c", "start", "SSH: " + p.Name, "ssh"}
	if p.Port > 0 && p.Port != 22 {
		args = append(args, "-p", strconv.Itoa(p.Port))
	}
	if p.KeyPath != "" {
		args = append(args, "-i", p.KeyPath)
	}
	args = append(args, p.User+"@"+p.Host)
	return exec.Command("cmd", args...).Start()
}
