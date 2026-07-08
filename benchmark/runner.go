package benchmark

import (
	"fmt"
	"sort"
	"time"

	"netpulse/dns"
	"netpulse/ping"
)

// Result результат бенчмарка одного DNS
type Result struct {
	Provider    string
	AvgLatency  float64
	MaxLatency  float64
	PacketLoss  float64
	ResolveTime time.Duration
	Category    string
}

// Runner запускает бенчмарк
type Runner struct {
	dnsChanger *dns.Changer
	pingEngine *ping.Engine
}

// NewRunner создаёт runner
func NewRunner() *Runner {
	return &Runner{
		dnsChanger: dns.NewChanger(),
		pingEngine: ping.NewEngine(4, 5000),
	}
}

// RunAll тестирует все DNS
func (r *Runner) RunAll() ([]Result, error) {
	targets := ping.GetDefaultTargets()
	
	// Сохраняем текущий DNS
	originalDNS, err := r.dnsChanger.GetCurrent()
	if err != nil {
		originalDNS = []string{}
	}

	var results []Result

	for name, config := range dns.Servers {
		fmt.Printf("\n⏳ Тестируем %s (%s)...\n", config.Name, name)

		// Меняем DNS
		if err := r.dnsChanger.Set(name); err != nil {
			fmt.Printf("   ⚠️  Ошибка смены DNS: %v\n", err)
			continue
		}

		// Ждём применения
		time.Sleep(5 * time.Second)

		// Пингуем
		pingResults := r.pingEngine.PingAll(targets)

		// Считаем метрики
		var allLatencies []float64
		var totalLoss float64
		var count int

		for _, pings := range pingResults {
			for _, p := range pings {
				if p.AvgRTT > 0 {
					allLatencies = append(allLatencies, p.AvgRTT)
				}
				totalLoss += p.PacketLoss
				count++
			}
		}

		avgLatency := 0.0
		maxLatency := 0.0
		if len(allLatencies) > 0 {
			var sum float64
			for _, v := range allLatencies {
				sum += v
				if v > maxLatency {
					maxLatency = v
				}
			}
			avgLatency = sum / float64(len(allLatencies))
		}

		avgLoss := 0.0
		if count > 0 {
			avgLoss = totalLoss / float64(count)
		}

		results = append(results, Result{
			Provider:    name,
			AvgLatency:  avgLatency,
			MaxLatency:  maxLatency,
			PacketLoss:  avgLoss,
			Category:    config.Category,
		})

		fmt.Printf("   ✅ Средняя latency: %.1fms, Loss: %.1f%%\n", avgLatency, avgLoss)
	}

	// Восстанавливаем оригинальный DNS
	fmt.Println("\n🔄 Восстановление оригинального DNS...")
	if len(originalDNS) > 0 {
		// Восстановление через backup
		r.dnsChanger.Restore()
	}

	// Сортируем по latency
	sort.Slice(results, func(i, j int) bool {
		return results[i].AvgLatency < results[j].AvgLatency
	})

	// Выводим рейтинг
	fmt.Println("\n" + "🏆 РЕЙТИНГ DNS ПО СРЕДНЕЙ ЗАДЕРЖКЕ")
	fmt.Println("=" + string(make([]byte, 60)))
	for i, r := range results {
		medal := "  "
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		}
		fmt.Printf("%s %-20s %6.1fms  (max: %6.1fms, loss: %.1f%%) [%s]\n",
			medal, r.Provider, r.AvgLatency, r.MaxLatency, r.PacketLoss, r.Category)
	}

	return results, nil
}