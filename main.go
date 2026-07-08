package main

import (
	"fmt"
	"os"

	"netpulse/dns"
	"netpulse/ping"
	"netpulse/ui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "dns":
			runDNSCommand(os.Args[2:])
		case "ping":
			runPingCommand(os.Args[2:])
		case "bench":
			runBenchmark()
		default:
			printHelp()
		}
		return
	}

	// Интерактивный TUI режим
	app := ui.NewApp()
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка: %v\n", err)
		os.Exit(1)
	}
}

func runDNSCommand(args []string) {
	changer := dns.NewChanger()
	
	if len(args) == 0 {
		current, _ := changer.GetCurrent()
		fmt.Printf("Текущие DNS: %v\n", current)
		return
	}

	switch args[0] {
	case "list":
		for name, srv := range dns.Servers {
			fmt.Printf("%-20s %s / %s\n", name, srv.IPv4[0], srv.IPv4[1])
		}
	case "set":
		if len(args) < 2 {
			fmt.Println("Использование: netpulse dns set <provider>")
			return
		}
		if err := changer.Set(args[1]); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			return
		}
		fmt.Printf("✅ DNS изменён на %s\n", args[1])
	case "backup":
		if err := changer.Backup(); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			return
		}
		fmt.Println("💾 Бэкап создан")
	case "restore":
		if err := changer.Restore(); err != nil {
			fmt.Printf("Ошибка: %v\n", err)
			return
		}
		fmt.Println("🔄 DNS восстановлен")
	}
}

func runPingCommand(args []string) {
	engine := ping.NewEngine(4, 5000)
	
	targets := ping.GetDefaultTargets()
	if len(args) > 0 {
		// Пинг конкретного таргета
		targets = map[string][]string{"custom": args}
	}

	fmt.Println("⏳ Запуск пинга...")
	results := engine.PingAll(targets)
	engine.PrintReport(results)
}

func runBenchmark() {
	fmt.Println("🏎️  Запуск бенчмарка всех DNS...")
	// Реализация в benchmark/runner.go
}

func printHelp() {
	fmt.Println(`NetPulse - DNS Changer + Ping Checker

Использование:
  netpulse              Интерактивный режим (TUI)
  netpulse dns list     Список доступных DNS
  netpulse dns set <n>  Установить DNS (google, cloudflare, yandex...)
  netpulse dns backup   Сохранить текущие DNS
  netpulse dns restore  Восстановить DNS из бэкапа
  netpulse ping         Пинг всех таргетов
  netpulse ping <host>  Пинг конкретного хоста
  netpulse bench        Бенчмарк всех DNS`)
}