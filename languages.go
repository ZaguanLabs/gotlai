package gotlai

import "strings"

// LanguageNames maps locale codes to human-readable names for AI prompts.
var LanguageNames = map[string]string{
	// Tier 1 (High Quality)
	"en_US": "English (United States)",
	"en_GB": "English (United Kingdom)",
	"de_DE": "German (Germany)",
	"es_ES": "Spanish (Spain)",
	"es_MX": "Spanish (Mexico)",
	"fr_FR": "French (France)",
	"it_IT": "Italian (Italy)",
	"ja_JP": "Japanese (Japan)",
	"pt_BR": "Portuguese (Brazil)",
	"pt_PT": "Portuguese (Portugal)",
	"zh_CN": "Chinese (Simplified)",
	"zh_TW": "Chinese (Traditional)",

	// Tier 2 (Good Quality)
	"ar_SA": "Arabic (Saudi Arabia)",
	"bn_BD": "Bengali (Bangladesh)",
	"cs_CZ": "Czech (Czech Republic)",
	"da_DK": "Danish (Denmark)",
	"el_GR": "Greek (Greece)",
	"fi_FI": "Finnish (Finland)",
	"he_IL": "Hebrew (Israel)",
	"hi_IN": "Hindi (India)",
	"hu_HU": "Hungarian (Hungary)",
	"id_ID": "Indonesian (Indonesia)",
	"ko_KR": "Korean (South Korea)",
	"nl_NL": "Dutch (Netherlands)",
	"nb_NO": "Norwegian Bokmål (Norway)",
	"pl_PL": "Polish (Poland)",
	"ro_RO": "Romanian (Romania)",
	"ru_RU": "Russian (Russia)",
	"sv_SE": "Swedish (Sweden)",
	"th_TH": "Thai (Thailand)",
	"tr_TR": "Turkish (Turkey)",
	"uk_UA": "Ukrainian (Ukraine)",
	"vi_VN": "Vietnamese (Vietnam)",

	// Tier 3 (Functional)
	"bg_BG": "Bulgarian (Bulgaria)",
	"ca_ES": "Catalan (Spain)",
	"fa_IR": "Persian (Iran)",
	"hr_HR": "Croatian (Croatia)",
	"lt_LT": "Lithuanian (Lithuania)",
	"lv_LV": "Latvian (Latvia)",
	"ms_MY": "Malay (Malaysia)",
	"sk_SK": "Slovak (Slovakia)",
	"sl_SI": "Slovenian (Slovenia)",
	"sr_RS": "Serbian (Serbia)",
	"sw_KE": "Swahili (Kenya)",
	"tl_PH": "Tagalog (Philippines)",
	"ur_PK": "Urdu (Pakistan)",
}

// ShortCodeToLocale maps short language codes to full locale codes.
var ShortCodeToLocale = map[string]string{
	"en": "en_US",
	"de": "de_DE",
	"es": "es_ES",
	"fr": "fr_FR",
	"it": "it_IT",
	"ja": "ja_JP",
	"pt": "pt_BR",
	"zh": "zh_CN",
	"ko": "ko_KR",
	"ru": "ru_RU",
	"ar": "ar_SA",
	"he": "he_IL",
	"hi": "hi_IN",
	"nl": "nl_NL",
	"pl": "pl_PL",
	"tr": "tr_TR",
	"vi": "vi_VN",
}

// GetLanguageName returns the human-readable name for a language code.
// Falls back to the code itself if not found.
func GetLanguageName(langCode string) string {
	if name, ok := LanguageNames[langCode]; ok {
		return name
	}
	// Try expanding short code
	if locale, ok := ShortCodeToLocale[langCode]; ok {
		if name, ok := LanguageNames[locale]; ok {
			return name
		}
	}
	return langCode
}

// GetDirection returns "rtl" for right-to-left languages, "ltr" otherwise.
func GetDirection(langCode string) string {
	// Extract base language code (e.g., "ar" from "ar_SA")
	base := strings.Split(langCode, "_")[0]
	base = strings.ToLower(base)

	if RTLLanguages[base] {
		return "rtl"
	}
	return "ltr"
}

// IsRTL returns true if the language uses right-to-left text direction.
func IsRTL(langCode string) bool {
	return GetDirection(langCode) == "rtl"
}

// NormalizeLocale converts a language code to the standard format (e.g., "es-ES" → "es_ES").
func NormalizeLocale(langCode string) string {
	return strings.ReplaceAll(langCode, "-", "_")
}

// ToHTMLLang converts a locale code to HTML lang attribute format (e.g., "es_ES" → "es-ES").
func ToHTMLLang(langCode string) string {
	return strings.ReplaceAll(langCode, "_", "-")
}
