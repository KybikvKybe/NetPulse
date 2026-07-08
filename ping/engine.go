package ping

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Result struct {
	Target      string
	IP          string
	PacketsSent int
	PacketsRecv int
	PacketLoss  float64
	AvgRTT      float64
	ResolveTime time.Duration
	Error       string
}

type Engine struct {
	Count   int
	Timeout int
}

func NewEngine(count, timeout int) *Engine {
	return &Engine{Count: count, Timeout: timeout}
}

func (e *Engine) Ping(target string) *Result {
	result := &Result{Target: target, PacketsSent: e.Count}

	// 1. DNS Resolve
	start := time.Now()
	ips, err := net.LookupIP(target)
	result.ResolveTime = time.Since(start)
	if err != nil {
		result.Error = fmt.Sprintf("resolve failed: %v", err)
		result.PacketLoss = 100
		return result
	}
	result.IP = ips[0].String()

	// 2. TCP Ping (на порт 443 - HTTPS)
	var totalRTT time.Duration
	successful := 0
	address := net.JoinHostPort(result.IP, "443")

	for i := 0; i < e.Count; i++ {
		start := time.Now()
		// Пытаемся установить TCP соединение
		conn, err := net.DialTimeout("tcp", address, time.Duration(e.Timeout)*time.Millisecond)
		if err == nil {
			totalRTT += time.Since(start)
			successful++
			conn.Close()
		}
		// Небольшая пауза между запросами, чтобы не забанили
		time.Sleep(20 * time.Millisecond)
	}

	result.PacketsRecv = successful
	result.PacketLoss = float64(e.Count-successful) / float64(e.Count) * 100

	if successful > 0 {
		result.AvgRTT = float64(totalRTT.Milliseconds()) / float64(successful)
	} else {
		// Если 443 закрыт, попробуем порт 80 (HTTP) один раз для проверки
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(result.IP, "80"), time.Duration(e.Timeout)*time.Millisecond)
		if err == nil {
			result.AvgRTT = float64(time.Since(start).Milliseconds())
			result.PacketsRecv = 1
			result.PacketLoss = float64(e.Count-1) / float64(e.Count) * 100
			conn.Close()
		}
	}

	return result
}

func (e *Engine) PingAll(targets map[string][]string) map[string][]*Result {
	results := make(map[string][]*Result)
	var mu sync.Mutex
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 20)

	for category, hosts := range targets {
		for _, host := range hosts {
			wg.Add(1)
			go func(cat, h string) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				r := e.Ping(h)
				mu.Lock()
				results[cat] = append(results[cat], r)
				mu.Unlock()
			}(category, host)
		}
	}
	wg.Wait()
	return results
}

func (e *Engine) PrintReport(results map[string][]*Result) {
	fmt.Println("\n" + "================================================================================")
	fmt.Println("📊 ОТЧЁТ ПО ПИНГУ (TCP MODE - Port 443)")
	fmt.Println("================================================================================")

	for category, pings := range results {
		fmt.Printf("\n📁 %s\n", category)
		fmt.Println("------------------------------------------------------------")
		for _, p := range pings {
			status := "✅"
			if p.PacketLoss == 100 { status = "❌" } else if p.PacketLoss > 0 { status = "⚠️" }
			
			if p.Error != "" {
				fmt.Printf("  %s %-35s 🔴 %s\n", status, p.Target, p.Error)
				continue
			}

			latencyIcon := "🟢"
			if p.AvgRTT > 150 { latencyIcon = "🔴" } else if p.AvgRTT > 50 { latencyIcon = "🟡" }

			if p.AvgRTT > 0 {
				fmt.Printf("  %s %-35s %s %6.1fms  (loss: %.0f%%)  IP: %s\n",
					status, p.Target, latencyIcon, p.AvgRTT, p.PacketLoss, p.IP)
			} else {
				fmt.Printf("  %s %-35s 🔴 UNREACHABLE  IP: %s\n", status, p.Target, p.IP)
			}
		}
	}
}