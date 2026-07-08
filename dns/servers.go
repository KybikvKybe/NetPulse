package dns

// ServerConfig описывает DNS-сервер
type ServerConfig struct {
	Name        string
	IPv4        []string
	IPv6        []string
	Description string
	Category    string // global, russia, privacy, family
}

// Servers — справочник всех DNS
var Servers = map[string]ServerConfig{
	"google": {
		Name:        "Google DNS",
		IPv4:        []string{"8.8.8.8", "8.8.4.4"},
		IPv6:        []string{"2001:4860:4860::8888", "2001:4860:4860::8844"},
		Description: "Самый быстрый глобальный DNS",
		Category:    "global",
	},
	"cloudflare": {
		Name:        "Cloudflare",
		IPv4:        []string{"1.1.1.1", "1.0.0.1"},
		IPv6:        []string{"2606:4700:4700::1111", "2606:4700:4700::1001"},
		Description: "Privacy-first, DoH/DoT",
		Category:    "privacy",
	},
	"cloudflare_malware": {
		Name:        "Cloudflare Malware",
		IPv4:        []string{"1.1.1.2", "1.0.0.2"},
		IPv6:        []string{"2606:4700:4700::1112", "2606:4700:4700::1002"},
		Description: "Блокировка malware",
		Category:    "security",
	},
	"cloudflare_family": {
		Name:        "Cloudflare Family",
		IPv4:        []string{"1.1.1.3", "1.0.0.3"},
		IPv6:        []string{"2606:4700:4700::1113", "2606:4700:4700::1003"},
		Description: "Блокировка malware + adult",
		Category:    "family",
	},
	"yandex": {
		Name:        "Yandex DNS",
		IPv4:        []string{"77.88.8.8", "77.88.8.1"},
		IPv6:        []string{"2a02:6b8::feed:0ff", "2a02:6b8:0:1::feed:0ff"},
		Description: "Лучший для РФ",
		Category:    "russia",
	},
	"yandex_safe": {
		Name:        "Yandex Safe",
		IPv4:        []string{"77.88.8.88", "77.88.8.2"},
		Description: "Антивирус + антифишинг",
		Category:    "security",
	},
	"yandex_family": {
		Name:        "Yandex Family",
		IPv4:        []string{"77.88.8.7", "77.88.8.3"},
		Description: "Защита детей",
		Category:    "family",
	},
	"quad9": {
		Name:        "Quad9",
		IPv4:        []string{"9.9.9.9", "149.112.112.112"},
		IPv6:        []string{"2620:fe::fe", "2620:fe::9"},
		Description: "Блокировка malware, GDPR",
		Category:    "security",
	},
	"quad9_nosec": {
		Name:        "Quad9 No Security",
		IPv4:        []string{"9.9.9.10", "149.112.112.10"},
		IPv6:        []string{"2620:fe::10", "2620:fe::fe:10"},
		Description: "Без фильтрации",
		Category:    "global",
	},
	"opendns": {
		Name:        "OpenDNS",
		IPv4:        []string{"208.67.222.222", "208.67.220.220"},
		IPv6:        []string{"2620:119:35::35", "2620:119:53::53"},
		Description: "Фильтры, требуется аккаунт",
		Category:    "global",
	},
	"opendns_family": {
		Name:        "OpenDNS Family",
		IPv4:        []string{"208.67.222.123", "208.67.220.123"},
		Description: "Family Shield",
		Category:    "family",
	},
	"adguard": {
		Name:        "AdGuard DNS",
		IPv4:        []string{"94.140.14.14", "94.140.15.15"},
		IPv6:        []string{"2a10:50c0::ad1:ff", "2a10:50c0::ad2:ff"},
		Description: "Блокировка рекламы",
		Category:    "privacy",
	},
	"adguard_family": {
		Name:        "AdGuard Family",
		IPv4:        []string{"94.140.14.15", "94.140.15.16"},
		IPv6:        []string{"2a10:50c0::bad1:ff", "2a10:50c0::bad2:ff"},
		Description: "Блокировка рекламы + adult",
		Category:    "family",
	},
	"gcore": {
		Name:        "G-Core Labs",
		IPv4:        []string{"95.85.95.85", "2.56.220.2"},
		IPv6:        []string{"2a03:90c0:999d::1", "2a03:90c0:9992::1"},
		Description: "Европа/СНГ",
		Category:    "russia",
	},
	"control_d": {
		Name:        "Control D",
		IPv4:        []string{"76.76.2.0", "76.76.10.0"},
		Description: "Кастомные фильтры",
		Category:    "privacy",
	},
	"nextdns": {
		Name:        "NextDNS",
		IPv4:        []string{"45.90.28.0", "45.90.28.255"},
		IPv6:        []string{"2a0d:2406:1801::", "2a0d:2406:1802::"},
		Description: "Персональный DNS",
		Category:    "privacy",
	},
	"dns_sb": {
		Name:        "DNS.SB",
		IPv4:        []string{"185.222.222.222", "45.11.45.11"},
		IPv6:        []string{"2a09::", "2a11::"},
		Description: "Швейцария",
		Category:    "privacy",
	},
	"dns0_eu": {
		Name:        "dns0.eu",
		IPv4:        []string{"193.110.81.0", "185.253.5.0"},
		IPv6:        []string{"2a0f:fc80::", "2a0f:fc81::"},
		Description: "Европейский",
		Category:    "privacy",
	},
	"verisign": {
		Name:        "Verisign",
		IPv4:        []string{"64.6.64.6", "64.6.65.6"},
		IPv6:        []string{"2620:74:1b::1:1", "2620:74:1c::2:2"},
		Description: "Стабильный",
		Category:    "global",
	},
	"dyn": {
		Name:        "Dyn (Oracle)",
		IPv4:        []string{"216.146.35.35", "216.146.36.36"},
		Description: "Старый игрок",
		Category:    "global",
	},
	"msk_ix": {
		Name:        "MSK-IX",
		IPv4:        []string{"62.76.76.62", "62.76.62.76"},
		IPv6:        []string{"2001:6d0:6d0::2001", "2001:6d0:d6::2001"},
		Description: "Россия",
		Category:    "russia",
	},
	"nsdi": {
		Name:        "НСДИ",
		IPv4:        []string{"195.208.4.1", "195.208.5.1"},
		IPv6:        []string{"2a0c:a9c7:8::1", "2a0c:a9c7:9::1"},
		Description: "Россия",
		Category:    "russia",
	},
}