package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/viewport"
	"netpulse/dns"
	"netpulse/ping"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1)

	statusOK    = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	statusWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00"))
	statusError = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

// === Сообщения для Bubble Tea ===

type pingCompleteMsg struct {
	results map[string][]*ping.Result
}

type dnsSetCompleteMsg struct {
	provider string
	err      error
}

// App TUI приложение
type App struct {
	mode       string
	spinner    spinner.Model
	results    map[string][]*ping.Result
	dnsChanger *dns.Changer
	pingEngine *ping.Engine
	width      int
	height     int
	dnsNames   []string  // отсортированные имена DNS для стабильного списка
	dnsCursor  int       // позиция курсора в списке DNS
	lastError  string
    viewPort   viewport.Model 	
}

// NewApp создаёт TUI
func NewApp() *App {
	s := spinner.New()
	s.Spinner = spinner.Dot

	// Собираем и сортируем имена DNS один раз
	var names []string
	for name := range dns.Servers {
		names = append(names, name)
	}
	sort.Strings(names)

	return &App{
		mode:       "menu",
		spinner:    s,
		dnsChanger: dns.NewChanger(),
		pingEngine: ping.NewEngine(4, 5000),
		dnsNames:   names,
		dnsCursor:  0,
		viewPort:   viewport.New(100, 100), // ДОБАВЬТЕ ЭТУ СТРОКУ
	}
}

// Run запускает TUI
func (a *App) Run() error {
	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) Init() tea.Cmd {
	return nil
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		// ДОБАВЬТЕ ЭТИ СТРОКИ:
		a.viewPort.Width = msg.Width
		a.viewPort.Height = msg.Height - 5 // -5 чтобы оставить место под нижнее меню
		return a, nil

	case tea.KeyMsg:
		switch a.mode {
		case "menu":
			return a.handleMenuKey(msg)
		case "dns":
			return a.handleDNSKey(msg)
		case "results", "bench_results":
			return a.handleResultsKey(msg)
		}

	case pingCompleteMsg:
		a.results = msg.results
		a.mode = "results"
		return a, nil

	case dnsSetCompleteMsg:
		if msg.err != nil {
			// ОШИБКА ДОЛЖНА ПОПАСТЬ СЮДА
			a.lastError = msg.err.Error() 
		} else {
			a.lastError = fmt.Sprintf("DNS изменён на %s", msg.provider)
		}
		return a, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd
	}

	return a, nil
}

func (a *App) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "1":
		a.mode = "dns"
		a.dnsCursor = 0
		a.lastError = ""
		return a, nil
	case "2":
		return a.runPingCmd()
	case "3":
		return a.runBenchmarkCmd()
	}
	return a, nil
}

func (a *App) handleDNSKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "b", "esc":
		a.mode = "menu"
		return a, nil
	case "up", "k":
		if a.dnsCursor > 0 {
			a.dnsCursor--
		}
		return a, nil
	case "down", "j":
		if a.dnsCursor < len(a.dnsNames)-1 {
			a.dnsCursor++
		}
		return a, nil
	case "enter":
		provider := a.dnsNames[a.dnsCursor]
		return a, func() tea.Msg {
			err := a.dnsChanger.Set(provider) // Вот здесь происходит магия
			return dnsSetCompleteMsg{provider: provider, err: err}
		}
	}
	return a, nil
}

func (a *App) handleResultsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "b", "esc":
		a.mode = "menu"
		a.results = nil
		return a, nil
	}
	return a, nil
}

func (a *App) runPingCmd() (tea.Model, tea.Cmd) {
	a.mode = "ping"
	return a, tea.Batch(
		a.spinner.Tick,
		func() tea.Msg {
			targets := ping.GetDefaultTargets()
			results := a.pingEngine.PingAll(targets)
			return pingCompleteMsg{results: results}
		},
	)
}

func (a *App) runBenchmarkCmd() (tea.Model, tea.Cmd) {
	a.mode = "bench"
	return a, tea.Batch(
		a.spinner.Tick,
		func() tea.Msg {
			// TODO: реализовать benchmark
			return pingCompleteMsg{results: nil}
		},
	)
}

func (a *App) View() string {
	switch a.mode {
	case "menu":
		return a.menuView()
	case "dns":
		return a.dnsView()
	case "ping":
		return a.pingView()
	case "results":
		return a.resultsView()
	case "bench":
		return a.benchView()
	case "bench_results":
		return a.benchResultsView()
	default:
		return a.menuView()
	}
}

func (a *App) menuView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(" 🌐 NetPulse - DNS Changer + Ping Checker "))
	b.WriteString("\n\n")

	b.WriteString("  [1] 🔄 Сменить DNS\n")
	b.WriteString("  [2] 📡 Запустить пинг-тест\n")
	b.WriteString("  [3] 🏎️  Бенчмарк всех DNS\n")
	b.WriteString("  [q] 🚪 Выход\n\n")

	// === ЗАМЕНЯЕМ ЭТОТ БЛОК ===
	current, err := a.dnsChanger.GetCurrent()
	if err != nil {
		// Теперь, если адаптер не найден, вы увидите это прямо в главном меню
		b.WriteString(fmt.Sprintf("  ⚠️ Ошибка адаптера: %v\n", err))
	} else {
		b.WriteString(fmt.Sprintf("  Текущие DNS: %v\n", current))
	}
	// ==========================

	return b.String()
}

func (a *App) dnsView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" 🔄 Выбор DNS (Enter — применить, стрелки — навигация) "))
	b.WriteString("\n\n")

	for i, name := range a.dnsNames {
		srv := dns.Servers[name]
		cursor := "  "
		if i == a.dnsCursor {
			cursor = "▸ "
		}
		b.WriteString(fmt.Sprintf("%s%-20s %s / %s  (%s)\n",
			cursor, name, srv.IPv4[0], srv.IPv4[1], srv.Description))
	}

	if a.lastError != "" {
		b.WriteString("\n\n") 
		b.WriteString(statusError.Render(" ⚠️  " + a.lastError + "\n"))
	}

	// 1. Устанавливаем контент в Viewport
	a.viewPort.SetContent(b.String())
	
	// 2. Авто-скролл: если курсор опустился слишком низко, скроллим Viewport
	// (Примерный расчет: каждая строка ~1 символ переноса)
	lineHeight := 1
	offset := a.dnsCursor * lineHeight
	if offset > a.viewPort.Height {
		a.viewPort.YOffset = offset - a.viewPort.Height
	} else if offset < 0 {
		a.viewPort.YOffset = 0
	}

	// 3. Возвращаем отрендеренный Viewport + нижнее меню
	return a.viewPort.View() + "\n\n  [b/esc] Назад  [q] Выход\n"
}

func (a *App) pingView() string {
	return fmt.Sprintf("\n\n  %s Запуск пинга... подождите\n\n", a.spinner.View())
}

func (a *App) benchView() string {
	return fmt.Sprintf("\n\n  %s Бенчмарк DNS... подождите\n\n", a.spinner.View())
}

func (a *App) resultsView() string {
	if a.results == nil {
		return "Загрузка..."
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(" 📊 Результаты пинга "))
	b.WriteString("\n\n")

	// --- СОРТИРОВКА КАТЕГОРИЙ ---
	categories := make([]string, 0, len(a.results))
	for cat := range a.results {
		categories = append(categories, cat)
	}
	sort.Strings(categories) // Теперь категории всегда будут в алфавитном порядке
	// ---------------------------

	for _, category := range categories {
		pings := a.results[category]
		b.WriteString(fmt.Sprintf("📁 %s\n", strings.ToUpper(category)))

		for _, p := range pings {
			if p.Error != "" {
				b.WriteString(statusError.Render(fmt.Sprintf("  ❌ %-30s %s\n", p.Target, p.Error)))
				continue
			}

			status := statusOK
			if p.PacketLoss > 0 {
				status = statusWarn
			}
			if p.PacketLoss == 100 {
				status = statusError
			}

			b.WriteString(status.Render(fmt.Sprintf(
				"  %-30s %6.1fms  loss: %.0f%%  %s\n",
				p.Target, p.AvgRTT, p.PacketLoss, p.IP)))
		}
		b.WriteString("\n")
	}

	b.WriteString("  [b/esc] Назад  [q] Выход\n")
	return b.String()
}

func (a *App) benchResultsView() string {
	return titleStyle.Render(" 🏆 Бенчмарк завершён ") + "\n\n  [b/esc] Назад\n"
}