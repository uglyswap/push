package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/uglyswap/push/internal/compat/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/uglyswap/push/internal/agent/tools/mcp"
	"github.com/uglyswap/push/internal/app"
	"github.com/uglyswap/push/internal/config"
	"github.com/uglyswap/push/internal/event"
	"github.com/uglyswap/push/internal/permission"
	"github.com/uglyswap/push/internal/pubsub"
	cmpChat "github.com/uglyswap/push/internal/tui/components/chat"
	"github.com/uglyswap/push/internal/tui/components/chat/splash"
	"github.com/uglyswap/push/internal/tui/components/completions"
	"github.com/uglyswap/push/internal/tui/components/core"
	"github.com/uglyswap/push/internal/tui/components/core/layout"
	"github.com/uglyswap/push/internal/tui/components/core/status"
	"github.com/uglyswap/push/internal/tui/components/dialogs"
	"github.com/uglyswap/push/internal/tui/components/dialogs/commands"
	"github.com/uglyswap/push/internal/tui/components/dialogs/filepicker"
	"github.com/uglyswap/push/internal/tui/components/dialogs/models"
	"github.com/uglyswap/push/internal/tui/components/dialogs/permissions"
	"github.com/uglyswap/push/internal/tui/components/dialogs/quit"
	"github.com/uglyswap/push/internal/tui/components/dialogs/sessions"
	"github.com/uglyswap/push/internal/tui/page"
	"github.com/uglyswap/push/internal/tui/page/chat"
	"github.com/uglyswap/push/internal/tui/styles"
	"github.com/uglyswap/push/internal/tui/util"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var lastMouseEvent time.Time

func MouseEventFilter(m tea.Model, msg tea.Msg) tea.Msg {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		// Filter mouse wheel and motion events
		if msg.Type == tea.MouseWheelUp || msg.Type == tea.MouseWheelDown || msg.Type == tea.MouseMotion {
			now := time.Now()
			// trackpad is sending too many requests
			if now.Sub(lastMouseEvent) < 15*time.Millisecond {
				return nil
			}
			lastMouseEvent = now
		}
	}
	return msg
}

// appModel represents the main application model that manages pages, dialogs, and UI state.
type appModel struct {
	wWidth, wHeight int // Window dimensions
	width, height   int
	keyMap          KeyMap

	currentPage  page.PageID
	previousPage page.PageID
	pages        map[page.PageID]util.Model
	loadedPages  map[page.PageID]bool

	// Status
	status          status.StatusCmp
	showingFullHelp bool

	app *app.App

	dialog       dialogs.DialogCmp
	completions  completions.Completions
	isConfigured bool

	// Chat Page Specific
	selectedSessionID string // The ID of the currently selected session

	// sendProgressBar instructs the TUI to send progress bar updates to the
	// terminal.
	sendProgressBar bool

	// QueryVersion instructs the TUI to query for the terminal version when it
	// starts.
	QueryVersion bool
}

// Init initializes the application model and returns initial commands.
func (a appModel) Init() tea.Cmd {
	item, ok := a.pages[a.currentPage]
	if !ok {
		return nil
	}

	var cmds []tea.Cmd
	cmd := item.Init()
	cmds = append(cmds, cmd)
	a.loadedPages[a.currentPage] = true

	cmd = a.status.Init()
	cmds = append(cmds, cmd)
	// In v1, tea.RequestTerminalVersion doesn't exist, skip version querying

	return tea.Batch(cmds...)
}

// Update handles incoming messages and updates the application state.
func (a *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	a.isConfigured = config.HasInitialDataConfig()

	switch msg := msg.(type) {
	// In v1, tea.EnvMsg, tea.TerminalVersionMsg, tea.KeyboardEnhancementsMsg don't exist
	// Progress bar and keyboard enhancements features are v2-only
	case tea.WindowSizeMsg:
		a.wWidth, a.wHeight = msg.Width, msg.Height
		a.completions.Update(msg)
		return a, a.handleWindowResize(msg.Width, msg.Height)

	case pubsub.Event[mcp.Event]:
		switch msg.Payload.Type {
		case mcp.EventStateChanged:
			return a, a.handleStateChanged(context.Background())
		case mcp.EventPromptsListChanged:
			return a, handleMCPPromptsEvent(context.Background(), msg.Payload.Name)
		case mcp.EventToolsListChanged:
			return a, handleMCPToolsEvent(context.Background(), msg.Payload.Name)
		}

	// Completions messages
	case completions.OpenCompletionsMsg, completions.FilterCompletionsMsg,
		completions.CloseCompletionsMsg, completions.RepositionCompletionsMsg:
		u, completionCmd := a.completions.Update(msg)
		if model, ok := u.(completions.Completions); ok {
			a.completions = model
		}

		return a, completionCmd

	// Dialog messages
	case dialogs.OpenDialogMsg, dialogs.CloseDialogMsg:
		u, completionCmd := a.completions.Update(completions.CloseCompletionsMsg{})
		a.completions = u.(completions.Completions)
		u, dialogCmd := a.dialog.Update(msg)
		a.dialog = u.(dialogs.DialogCmp)
		return a, tea.Batch(completionCmd, dialogCmd)
	case commands.ShowArgumentsDialogMsg:
		var args []commands.Argument
		for _, arg := range msg.ArgNames {
			args = append(args, commands.Argument{
				Name:     arg,
				Title:    cases.Title(language.English).String(arg),
				Required: true,
			})
		}
		return a, util.CmdHandler(
			dialogs.OpenDialogMsg{
				Model: commands.NewCommandArgumentsDialog(
					msg.CommandID,
					msg.CommandID,
					msg.CommandID,
					msg.Description,
					args,
					msg.OnSubmit,
				),
			},
		)
	case commands.ShowMCPPromptArgumentsDialogMsg:
		args := make([]commands.Argument, 0, len(msg.Prompt.Arguments))
		for _, arg := range msg.Prompt.Arguments {
			args = append(args, commands.Argument(*arg))
		}
		dialog := commands.NewCommandArgumentsDialog(
			msg.Prompt.Name,
			msg.Prompt.Title,
			msg.Prompt.Name,
			msg.Prompt.Description,
			args,
			msg.OnSubmit,
		)
		return a, util.CmdHandler(
			dialogs.OpenDialogMsg{
				Model: dialog,
			},
		)
	// Page change messages
	case page.PageChangeMsg:
		return a, a.moveToPage(msg.ID)

	// Status Messages
	case util.InfoMsg, util.ClearStatusMsg:
		s, statusCmd := a.status.Update(msg)
		a.status = s.(status.StatusCmp)
		cmds = append(cmds, statusCmd)
		return a, tea.Batch(cmds...)

	// Session
	case cmpChat.SessionSelectedMsg:
		a.selectedSessionID = msg.ID
	case cmpChat.SessionClearedMsg:
		a.selectedSessionID = ""
	// Commands
	case commands.SwitchSessionsMsg:
		return a, func() tea.Msg {
			allSessions, _ := a.app.Sessions.List(context.Background())
			return dialogs.OpenDialogMsg{
				Model: sessions.NewSessionDialogCmp(allSessions, a.selectedSessionID),
			}
		}

	case commands.SwitchModelMsg:
		return a, util.CmdHandler(
			dialogs.OpenDialogMsg{
				Model: models.NewModelDialogCmp(),
			},
		)
	// Compact
	case commands.CompactMsg:
		return a, func() tea.Msg {
			err := a.app.AgentCoordinator.Summarize(context.Background(), msg.SessionID)
			if err != nil {
				return util.ReportError(err)()
			}
			return nil
		}
	case commands.QuitMsg:
		return a, util.CmdHandler(dialogs.OpenDialogMsg{
			Model: quit.NewQuitDialog(),
		})
	case commands.ToggleYoloModeMsg:
		a.app.Permissions.SetSkipRequests(!a.app.Permissions.SkipRequests())
	case commands.ToggleHelpMsg:
		a.status.ToggleFullHelp()
		a.showingFullHelp = !a.showingFullHelp
		return a, a.handleWindowResize(a.wWidth, a.wHeight)
	// Model Switch
	case models.ModelSelectedMsg:
		if a.app.AgentCoordinator.IsBusy() {
			return a, util.ReportWarn("Agent is busy, please wait...")
		}

		cfg := config.Get()
		if err := cfg.UpdatePreferredModel(msg.ModelType, msg.Model); err != nil {
			return a, util.ReportError(err)
		}

		go a.app.UpdateAgentModel(context.TODO())

		modelTypeName := "large"
		if msg.ModelType == config.SelectedModelTypeSmall {
			modelTypeName = "small"
		}
		return a, util.ReportInfo(fmt.Sprintf("%s model changed to %s", modelTypeName, msg.Model.Model))

	// File Picker
	case commands.OpenFilePickerMsg:
		event.FilePickerOpened()

		if a.dialog.ActiveDialogID() == filepicker.FilePickerID {
			// If the commands dialog is already open, close it
			return a, util.CmdHandler(dialogs.CloseDialogMsg{})
		}
		return a, util.CmdHandler(dialogs.OpenDialogMsg{
			Model: filepicker.NewFilePickerCmp(a.app.Config().WorkingDir()),
		})
	// Permissions
	case pubsub.Event[permission.PermissionNotification]:
		item, ok := a.pages[a.currentPage]
		if !ok {
			return a, nil
		}

		// Forward to view.
		updated, itemCmd := item.Update(msg)
		a.pages[a.currentPage] = updated

		return a, itemCmd
	case pubsub.Event[permission.PermissionRequest]:
		return a, util.CmdHandler(dialogs.OpenDialogMsg{
			Model: permissions.NewPermissionDialogCmp(msg.Payload, &permissions.Options{
				DiffMode: config.Get().Options.TUI.DiffMode,
			}),
		})
	case permissions.PermissionResponseMsg:
		switch msg.Action {
		case permissions.PermissionAllow:
			a.app.Permissions.Grant(msg.Permission)
		case permissions.PermissionAllowForSession:
			a.app.Permissions.GrantPersistent(msg.Permission)
		case permissions.PermissionDeny:
			a.app.Permissions.Deny(msg.Permission)
		}
		return a, nil
	case splash.OnboardingCompleteMsg:
		item, ok := a.pages[a.currentPage]
		if !ok {
			return a, nil
		}

		a.isConfigured = config.HasInitialDataConfig()
		updated, pageCmd := item.Update(msg)
		a.pages[a.currentPage] = updated

		cmds = append(cmds, pageCmd)
		return a, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		return a, a.handleKeyPressMsg(msg)

	case tea.MouseWheelMsg:
		if a.dialog.HasDialogs() {
			u, dialogCmd := a.dialog.Update(msg)
			a.dialog = u.(dialogs.DialogCmp)
			cmds = append(cmds, dialogCmd)
		} else {
			item, ok := a.pages[a.currentPage]
			if !ok {
				return a, nil
			}

			updated, pageCmd := item.Update(msg)
			a.pages[a.currentPage] = updated

			cmds = append(cmds, pageCmd)
		}
		return a, tea.Batch(cmds...)
	case tea.PasteMsg:
		if a.dialog.HasDialogs() {
			u, dialogCmd := a.dialog.Update(msg)
			if model, ok := u.(dialogs.DialogCmp); ok {
				a.dialog = model
			}

			cmds = append(cmds, dialogCmd)
		} else {
			item, ok := a.pages[a.currentPage]
			if !ok {
				return a, nil
			}

			updated, pageCmd := item.Update(msg)
			a.pages[a.currentPage] = updated

			cmds = append(cmds, pageCmd)
		}
		return a, tea.Batch(cmds...)
	// Update Available
	case pubsub.UpdateAvailableMsg:
		// Update notifications disabled for push fork
		return a, nil
	}
	s, _ := a.status.Update(msg)
	a.status = s.(status.StatusCmp)

	item, ok := a.pages[a.currentPage]
	if !ok {
		return a, nil
	}

	updated, cmd := item.Update(msg)
	a.pages[a.currentPage] = updated

	if a.dialog.HasDialogs() {
		u, dialogCmd := a.dialog.Update(msg)
		if model, ok := u.(dialogs.DialogCmp); ok {
			a.dialog = model
		}

		cmds = append(cmds, dialogCmd)
	}
	cmds = append(cmds, cmd)
	return a, tea.Batch(cmds...)
}

// handleWindowResize processes window resize events and updates all components.
func (a *appModel) handleWindowResize(width, height int) tea.Cmd {
	var cmds []tea.Cmd

	// TODO: clean up these magic numbers.
	if a.showingFullHelp {
		height -= 5
	} else {
		height -= 2
	}

	a.width, a.height = width, height
	// Update status bar
	s, cmd := a.status.Update(tea.WindowSizeMsg{Width: width, Height: height})
	if model, ok := s.(status.StatusCmp); ok {
		a.status = model
	}
	cmds = append(cmds, cmd)

	// Update the current view.
	for p, page := range a.pages {
		updated, pageCmd := page.Update(tea.WindowSizeMsg{Width: width, Height: height})
		a.pages[p] = updated

		cmds = append(cmds, pageCmd)
	}

	// Update the dialogs
	dialog, cmd := a.dialog.Update(tea.WindowSizeMsg{Width: width, Height: height})
	if model, ok := dialog.(dialogs.DialogCmp); ok {
		a.dialog = model
	}

	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

// handleKeyPressMsg processes keyboard input and routes to appropriate handlers.
func (a *appModel) handleKeyPressMsg(msg tea.KeyPressMsg) tea.Cmd {
	// Check this first as the user should be able to quit no matter what.
	if key.Matches(msg, a.keyMap.Quit) {
		if a.dialog.ActiveDialogID() == quit.QuitDialogID {
			return tea.Quit
		}
		return util.CmdHandler(dialogs.OpenDialogMsg{
			Model: quit.NewQuitDialog(),
		})
	}

	if a.completions.Open() {
		// completions
		keyMap := a.completions.KeyMap()
		switch {
		case key.Matches(msg, keyMap.Up), key.Matches(msg, keyMap.Down),
			key.Matches(msg, keyMap.Select), key.Matches(msg, keyMap.Cancel),
			key.Matches(msg, keyMap.UpInsert), key.Matches(msg, keyMap.DownInsert):
			u, cmd := a.completions.Update(msg)
			a.completions = u.(completions.Completions)
			return cmd
		}
	}
	if a.dialog.HasDialogs() {
		u, dialogCmd := a.dialog.Update(msg)
		a.dialog = u.(dialogs.DialogCmp)
		return dialogCmd
	}
	switch {
	// help
	case key.Matches(msg, a.keyMap.Help):
		a.status.ToggleFullHelp()
		a.showingFullHelp = !a.showingFullHelp
		return a.handleWindowResize(a.wWidth, a.wHeight)
	// dialogs
	case key.Matches(msg, a.keyMap.Commands):
		// if the app is not configured show no commands
		if !a.isConfigured {
			return nil
		}
		if a.dialog.ActiveDialogID() == commands.CommandsDialogID {
			return util.CmdHandler(dialogs.CloseDialogMsg{})
		}
		if a.dialog.HasDialogs() {
			return nil
		}
		return util.CmdHandler(dialogs.OpenDialogMsg{
			Model: commands.NewCommandDialog(a.selectedSessionID),
		})
	case key.Matches(msg, a.keyMap.Models):
		// if the app is not configured show no models
		if !a.isConfigured {
			return nil
		}
		if a.dialog.ActiveDialogID() == models.ModelsDialogID {
			return util.CmdHandler(dialogs.CloseDialogMsg{})
		}
		if a.dialog.HasDialogs() {
			return nil
		}
		return util.CmdHandler(dialogs.OpenDialogMsg{
			Model: models.NewModelDialogCmp(),
		})
	case key.Matches(msg, a.keyMap.Sessions):
		// if the app is not configured show no sessions
		if !a.isConfigured {
			return nil
		}
		if a.dialog.ActiveDialogID() == sessions.SessionsDialogID {
			return util.CmdHandler(dialogs.CloseDialogMsg{})
		}
		if a.dialog.HasDialogs() && a.dialog.ActiveDialogID() != commands.CommandsDialogID {
			return nil
		}
		var cmds []tea.Cmd
		cmds = append(cmds,
			func() tea.Msg {
				allSessions, _ := a.app.Sessions.List(context.Background())
				return dialogs.OpenDialogMsg{
					Model: sessions.NewSessionDialogCmp(allSessions, a.selectedSessionID),
				}
			},
		)
		return tea.Sequence(cmds...)
	case key.Matches(msg, a.keyMap.Suspend):
		if a.app.AgentCoordinator != nil && a.app.AgentCoordinator.IsBusy() {
			return util.ReportWarn("Agent is busy, please wait...")
		}
		// In v1, tea.Suspend doesn't exist - just return nil
		return nil
	default:
		item, ok := a.pages[a.currentPage]
		if !ok {
			return nil
		}

		updated, cmd := item.Update(msg)
		a.pages[a.currentPage] = updated
		return cmd
	}
}

// moveToPage handles navigation between different pages in the application.
func (a *appModel) moveToPage(pageID page.PageID) tea.Cmd {
	if a.app.AgentCoordinator.IsBusy() {
		// TODO: maybe remove this :  For now we don't move to any page if the agent is busy
		return util.ReportWarn("Agent is busy, please wait...")
	}

	var cmds []tea.Cmd
	if _, ok := a.loadedPages[pageID]; !ok {
		cmd := a.pages[pageID].Init()
		cmds = append(cmds, cmd)
		a.loadedPages[pageID] = true
	}
	a.previousPage = a.currentPage
	a.currentPage = pageID
	if sizable, ok := a.pages[a.currentPage].(layout.Sizeable); ok {
		cmd := sizable.SetSize(a.width, a.height)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}

// View renders the complete application interface including pages, dialogs, and overlays.
func (a *appModel) View() string {
	t := styles.CurrentTheme()

	if a.wWidth < 25 || a.wHeight < 15 {
		return t.S().Base.Width(a.wWidth).Height(a.wHeight).
			Align(lipgloss.Center, lipgloss.Center).
			Render(t.S().Base.
				Padding(1, 4).
				Foreground(styles.TC(t.White)).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(styles.TC(t.Primary)).
				Render("Window too small!"),
			)
	}

	page := a.pages[a.currentPage]
	if withHelp, ok := page.(core.KeyMapHelp); ok {
		a.status.SetKeyMap(withHelp.Help())
	}
	pageView := page.View()
	components := []string{
		pageView,
	}
	components = append(components, a.status.View())

	appView := lipgloss.JoinVertical(lipgloss.Top, components...)

	// In v1 lipgloss, there's no Layer/Compositor system
	// We use lipgloss.Place to overlay dialogs and completions
	result := appView

	if a.dialog.HasDialogs() {
		// Overlay the active dialog on top of the app view
		activeDialog := a.dialog.ActiveModel()
		if activeDialog != nil {
			dialogView := activeDialog.View()
			row, col := activeDialog.(dialogs.DialogModel).Position()
			// Place the dialog at its specified position
			result = placeOverlay(result, dialogView, col, row, a.wWidth, a.wHeight)
		}
	}

	if a.completions.Open() {
		cmpView := a.completions.View()
		x, y := a.completions.Position()
		// Place completions at specific position by padding
		if x > 0 || y > 0 {
			paddedCmp := lipgloss.NewStyle().MarginLeft(x).MarginTop(y).Render(cmpView)
			// Overlay on top of current result
			result = lipgloss.Place(a.wWidth, a.wHeight, lipgloss.Left, lipgloss.Top,
				paddedCmp,
				lipgloss.WithWhitespaceChars(" "),
			)
		}
	}

	return result
}

// placeOverlay places an overlay on top of a background at the specified position.
func placeOverlay(background, overlay string, x, y, width, height int) string {
	// Handle negative positions by clamping to 0
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	bgLines := strings.Split(background, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// Ensure background has enough lines
	for len(bgLines) < height {
		bgLines = append(bgLines, strings.Repeat(" ", width))
	}

	// Overlay each line
	for i, overlayLine := range overlayLines {
		bgY := y + i
		if bgY >= 0 && bgY < len(bgLines) {
			bgLine := bgLines[bgY]
			overlayWidth := lipgloss.Width(overlayLine)

			// Build the new line: prefix + overlay + suffix
			prefix := ""
			if x > 0 {
				prefixWidth := min(x, lipgloss.Width(bgLine))
				// Get visible prefix from background
				prefix = truncateWithWidth(bgLine, prefixWidth)
				// Pad if needed
				for lipgloss.Width(prefix) < x {
					prefix += " "
				}
			}

			suffix := ""
			suffixStart := x + overlayWidth
			if suffixStart < lipgloss.Width(bgLine) {
				suffix = cutLeft(bgLine, suffixStart)
			}

			bgLines[bgY] = prefix + overlayLine + suffix
		}
	}

	return strings.Join(bgLines, "\n")
}

// truncateWithWidth truncates a string to the specified visible width.
func truncateWithWidth(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Simple truncation - for ANSI strings this is approximate
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i])
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return ""
}

// cutLeft removes the first n visible characters from a string.
func cutLeft(s string, n int) string {
	if n <= 0 {
		return s
	}
	if n >= lipgloss.Width(s) {
		return ""
	}
	// Skip n visible characters
	runes := []rune(s)
	width := 0
	for i, r := range runes {
		width += lipgloss.Width(string(r))
		if width >= n {
			return string(runes[i+1:])
		}
	}
	return ""
}

func (a *appModel) handleStateChanged(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		a.app.UpdateAgentModel(ctx)
		return nil
	}
}

func handleMCPPromptsEvent(ctx context.Context, name string) tea.Cmd {
	return func() tea.Msg {
		mcp.RefreshPrompts(ctx, name)
		return nil
	}
}

func handleMCPToolsEvent(ctx context.Context, name string) tea.Cmd {
	return func() tea.Msg {
		mcp.RefreshTools(ctx, name)
		return nil
	}
}

// New creates and initializes a new TUI application model.
func New(app *app.App) *appModel {
	chatPage := chat.New(app)
	keyMap := DefaultKeyMap()
	keyMap.pageBindings = chatPage.Bindings()

	model := &appModel{
		currentPage: chat.ChatPageID,
		app:         app,
		status:      status.NewStatusCmp(),
		loadedPages: make(map[page.PageID]bool),
		keyMap:      keyMap,

		pages: map[page.PageID]util.Model{
			chat.ChatPageID: chatPage,
		},

		dialog:      dialogs.NewDialogCmp(),
		completions: completions.New(),
	}

	return model
}
