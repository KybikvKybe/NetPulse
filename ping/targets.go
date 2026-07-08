package ping

// GetDefaultTargets возвращает все таргеты для пинга
func GetDefaultTargets() map[string][]string {
	return map[string][]string{
		// Google экосистема
		"google": {
			"google.com",
			"www.google.com",
			"dns.google",
		},
		"youtube": {
			"youtube.com",
			"www.youtube.com",
			"youtu.be",
			"googlevideo.com",
			"i.ytimg.com",
			"s.ytimg.com",
		},
		"gstatic": {
			"gstatic.com",
			"fonts.gstatic.com",
			"ssl.gstatic.com",
			"www.gstatic.com",
			"csi.gstatic.com",
		},
		"gemini": {
			"gemini.google.com",
			"aistudio.google.com",
			"generativelanguage.googleapis.com",
		},
		"google_apis": {
			"apis.google.com",
			"ajax.googleapis.com",
			"oauth2.googleapis.com",
		},

		// VK экосистема
		"vk": {
			"vk.com",
			"vk.ru",
			"m.vk.com",
			"vkvideo.ru",
			"vkpay.io",
			"vk.cc",
		},
		"mailru": {
			"mail.ru",
			"imap.mail.ru",
			"smtp.mail.ru",
			"cloud.mail.ru",
			"my.mail.ru",
		},

		// Yandex экосистема
		"yandex": {
			"yandex.ru",
			"ya.ru",
			"disk.yandex.ru",
			"music.yandex.ru",
			"cloud.yandex.ru",
			"maps.yandex.ru",
			"market.yandex.ru",
			"direct.yandex.ru",
		},

		// GitHub экосистема
		"github": {
			"github.com",
			"www.github.com",
			"raw.githubusercontent.com",
			"api.github.com",
			"gist.github.com",
			"github.io",
			"githubusercontent.com",
		},

		// Telegram
		"telegram": {
			"web.telegram.org",
			"telegram.org",
			"api.telegram.org",
			"my.telegram.org",
		},
		"telegram_mtproto": {
			"149.154.167.99",
			"149.154.175.100",
			"149.154.167.51",
			"149.154.175.53",
		},

		// Kimi / Moonshot AI
		"kimi": {
			"kimi.com",
			"www.kimi.com",
			"kimi.moonshot.cn",
			"api.moonshot.cn",
			"moonshot.cn",
		},

		// Cloudflare / CDN
		"cloudflare": {
			"cloudflare.com",
			"1.1.1.1",
		},

		// Дополнительные полезные сервисы
		"dns_test": {
			"dns.google",
			"one.one.one.one",
			"dns.quad9.net",
		},
	}
}