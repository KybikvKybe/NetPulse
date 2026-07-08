package dns

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"net"
	"golang.org/x/sys/windows"
)

// Changer управляет системными DNS на Windows
type Changer struct {
	backupPath string
}

// NewChanger создаёт новый экземпляр
func NewChanger() *Changer {
	appData, _ := os.UserConfigDir()
	return &Changer{
		backupPath: filepath.Join(appData, "NetPulse", "dns_backup.json"),
	}
}

// GetCurrent получает текущие DNS-серверы активного адаптера
func (c *Changer) GetCurrent() ([]string, error) {
	adapter, err := c.getActiveAdapter()
	if err != nil {
		return nil, err
	}

	// Используем name=
	cmd := exec.Command("netsh", "interface", "ip", "show", "dns", "name="+adapter)
	output, err := cmd.Output()
	
	if err != nil {
		return []string{"DHCP (Автоматически)"}, nil
	}

	return c.parseDNSOutput(string(output)), nil
}

// Set устанавливает DNS от указанного провайдера
func (c *Changer) Set(provider string) error {
	config, ok := Servers[provider]
	if !ok {
		return fmt.Errorf("неизвестный провайдер: %s", provider)
	}

	if err := c.Backup(); err != nil {
	}

	adapter, err := c.getActiveAdapter()
	if err != nil {
		return err
	}

	// Используем name=
	adapterArg := "name=" + adapter

	primary := config.IPv4[0]
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns",
		adapterArg, "static", primary, "primary")
	
	if err := c.runElevated(cmd); err != nil {
		return fmt.Errorf("failed to set primary DNS: %w", err)
	}

	if len(config.IPv4) > 1 {
		secondary := config.IPv4[1]
		cmd = exec.Command("netsh", "interface", "ip", "add", "dns",
			adapterArg, secondary, "index=2")
		c.runElevated(cmd)
	}

	c.flushDNSCache()
	return nil
}
// Backup сохраняет текущие DNS
func (c *Changer) Backup() error {
	current, err := c.GetCurrent()
	if err != nil {
		return err
	}

	backup := struct {
		DNS       []string `json:"dns"`
		Timestamp int64    `json:"timestamp"`
	}{
		DNS:       current,
		Timestamp: time.Now().Unix(),
	}

	// Создаём директорию если нужно
	os.MkdirAll(filepath.Dir(c.backupPath), 0755)

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.backupPath, data, 0644)
}

// Restore восстанавливает DNS из бэкапа
func (c *Changer) Restore() error {
	data, err := os.ReadFile(c.backupPath)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	var backup struct {
		DNS []string `json:"dns"`
	}
	if err := json.Unmarshal(data, &backup); err != nil {
		return err
	}

	if len(backup.DNS) == 0 {
		return fmt.Errorf("backup is empty")
	}

	adapter, err := c.getActiveAdapter()
	if err != nil {
		return err
	}

	// Устанавливаем из бэкапа
	cmd := exec.Command("netsh", "interface", "ip", "set", "dns",
		adapter, "static", backup.DNS[0], "primary")
	if err := c.runElevated(cmd); err != nil {
		return err
	}

	for i := 1; i < len(backup.DNS) && i < 2; i++ {
		cmd = exec.Command("netsh", "interface", "ip", "add", "dns",
			adapter, backup.DNS[i], fmt.Sprintf("index=%d", i+1))
		c.runElevated(cmd) // игнорируем ошибки secondary
	}

	c.flushDNSCache()
	return nil
}

// getActiveAdapter находит активный сетевой адаптер
func (c *Changer) getActiveAdapter() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get interfaces: %v", err)
	}

	virtualKeywords := []string{"vpn", "hamachi", "zerotier", "virtual", "radmin", "vmware", "pseudo", "tunnel"}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		nameLower := strings.ToLower(iface.Name)
		isVirtual := false
		for _, keyword := range virtualKeywords {
			if strings.Contains(nameLower, keyword) {
				isVirtual = true
				break
			}
		}
		if isVirtual {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		hasIPv4 := false
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				hasIPv4 = true
				break
			}
		}

		if !hasIPv4 {
			continue
		}

		// ВОЗВРАЩАЕМ ИМЯ (например, "Ethernet")
		return iface.Name, nil 
	}

	return "", fmt.Errorf("no physical active adapter found")
}

// getFriendlyName конвертирует GUID интерфейса в понятное имя (например, "Ethernet")
func (c *Changer) getFriendlyName(guid string) string {
	// Запрашиваем список интерфейсов через netsh
	cmd := exec.Command("netsh", "interface", "show", "interface")
	output, err := cmd.Output()
	if err != nil {
		return guid // Если не удалось, возвращаем GUID (может сработать)
	}

	// Ищем строку, где упоминается наш GUID или имя
	// netsh обычно выводит список, где можно сопоставить индекс или имя.
	// Для упрощения: если мы нашли активный интерфейс через net, 
	// попробуем найти его имя в выводе netsh.
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Connected") {
			// Берем первое попавшееся «Подключенное» имя из netsh, 
			// так как обычно оно одно основное.
			re := regexp.MustCompile(`\s+Connected\s+(\S+.*\S+)\s*$`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				return strings.TrimSpace(matches[1])
			}
		}
	}
	
	return guid
}

// parseDNSOutput парсит вывод netsh show dns
func (c *Changer) parseDNSOutput(output string) []string {
	// Регулярное выражение для поиска любого IPv4 адреса
	re := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	matches := re.FindAllString(output, -1)

	if len(matches) > 0 {
		return matches
	}

	// Если IP не найдены, проверяем, нет ли там упоминания DHCP
	if strings.Contains(output, "DHCP") || strings.Contains(output, "динамически") {
		return []string{"DHCP (Автоматически)"}
	}

	return []string{"Не определено"}
}

// runElevated запускает команду с правами администратора
func (c *Changer) runElevated(cmd *exec.Cmd) error {
	if !c.isAdmin() {
		return fmt.Errorf("ошибка: программа должна быть запущена от имени АДМИНИСТРАТОРА")
	}

	// Создаем буфер, куда пойдет ВЕСЬ вывод команды
	output := &strings.Builder{}
	cmd.Stdout = output
	cmd.Stderr = output

	err := cmd.Run()
	if err != nil {
		// Печатаем вообще всё, что ответила система
		fmt.Printf("\n🔴 SYSTEM ERROR: %v\n🔴 FULL OUTPUT: %s\n", err, output.String())
		return fmt.Errorf("%v: %s", err, output.String())
	}
	return nil
}

// isAdmin проверяет права администратора
func (c *Changer) isAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	return err == nil && member
}

// flushDNSCache сбрасывает DNS-кэш
func (c *Changer) flushDNSCache() {
	exec.Command("ipconfig", "/flushdns").Run()
}