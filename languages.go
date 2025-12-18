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
	// Tier 1 Languages
	"en": "en_US",
	"gb": "en_GB", // Great Britain -> British English
	"uk": "en_GB", // UK (country code) -> British English
	"de": "de_DE",
	"es": "es_ES",
	"mx": "es_MX", // Mexico -> Mexican Spanish
	"fr": "fr_FR",
	"it": "it_IT",
	"ja": "ja_JP",
	"jp": "ja_JP", // Japan (country code) -> Japanese
	"pt": "pt_BR",
	"br": "pt_BR", // Brazil (country code) -> Brazilian Portuguese
	"zh": "zh_CN",
	"cn": "zh_CN", // China (country code) -> Simplified Chinese
	"tw": "zh_TW", // Taiwan -> Traditional Chinese
	"hk": "zh_TW", // Hong Kong -> Traditional Chinese

	// Tier 2 Languages
	"ar": "ar_SA",
	"sa": "ar_SA", // Saudi Arabia (country code) -> Arabic
	"bn": "bn_BD",
	"bd": "bn_BD", // Bangladesh (country code) -> Bengali
	"cs": "cs_CZ",
	"cz": "cs_CZ", // Czech Republic (country code) -> Czech
	"da": "da_DK",
	"dk": "da_DK", // Denmark (country code) -> Danish
	"el": "el_GR",
	"gr": "el_GR", // Greece (country code) -> Greek
	"fi": "fi_FI",
	"he": "he_IL",
	"il": "he_IL", // Israel (country code) -> Hebrew
	"iw": "he_IL", // Legacy Hebrew code -> Hebrew
	"hi": "hi_IN",
	"in": "hi_IN", // India (country code) -> Hindi
	"hu": "hu_HU",
	"id": "id_ID",
	"ko": "ko_KR",
	"kr": "ko_KR", // South Korea (country code) -> Korean
	"nl": "nl_NL",
	"nb": "nb_NO",
	"no": "nb_NO", // Norwegian -> Bokmål
	"nn": "nb_NO", // Nynorsk -> Bokmål (closest supported)
	"pl": "pl_PL",
	"ro": "ro_RO",
	"ru": "ru_RU",
	"sv": "sv_SE",
	"se": "sv_SE", // Sweden (country code) -> Swedish
	"th": "th_TH",
	"tr": "tr_TR",
	"ua": "uk_UA", // Ukraine (country code) -> Ukrainian
	"vi": "vi_VN",
	"vn": "vi_VN", // Vietnam (country code) -> Vietnamese

	// Tier 3 Languages
	"bg":  "bg_BG",
	"ca":  "ca_ES",
	"fa":  "fa_IR",
	"ir":  "fa_IR", // Iran (country code) -> Persian
	"hr":  "hr_HR",
	"lt":  "lt_LT",
	"lv":  "lv_LV",
	"ms":  "ms_MY",
	"my":  "ms_MY", // Malaysia (country code) -> Malay
	"sk":  "sk_SK",
	"sl":  "sl_SI",
	"si":  "sl_SI", // Slovenia (country code) -> Slovenian
	"sr":  "sr_RS",
	"rs":  "sr_RS", // Serbia (country code) -> Serbian
	"sw":  "sw_KE",
	"ke":  "sw_KE", // Kenya (country code) -> Swahili
	"tl":  "tl_PH",
	"fil": "tl_PH", // Filipino -> Tagalog
	"ph":  "tl_PH", // Philippines (country code) -> Tagalog
	"ur":  "ur_PK",
	"pk":  "ur_PK", // Pakistan (country code) -> Urdu
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

// LocaleClarifications provides language-specific hints for the AI model.
// Helps the model understand which variant to use.
var LocaleClarifications = map[string]string{
	// Norwegian variants
	"nb_NO": "Use Norwegian Bokmål (nb-NO), not Nynorsk.",
	"nb":    "Use Norwegian Bokmål (nb-NO), not Nynorsk.",
	"no":    "Use Norwegian Bokmål (nb-NO).",
	"nn_NO": "Use Norwegian Nynorsk (nn-NO), not Bokmål.",
	"nn":    "Use Norwegian Nynorsk (nn-NO), not Bokmål.",
	// Chinese variants
	"zh_CN": "Use Simplified Chinese characters.",
	"zh_TW": "Use Traditional Chinese characters.",
	"zh":    "Use Simplified Chinese characters.",
	// Portuguese variants
	"pt_BR": "Use Brazilian Portuguese conventions.",
	"pt_PT": "Use European Portuguese conventions.",
	"pt":    "Use Brazilian Portuguese conventions.",
	// English variants
	"en_GB": "Use British English spelling and conventions.",
	"en_US": "Use American English spelling and conventions.",
	// Spanish variants
	"es_ES": "Use Castilian Spanish (Spain) conventions.",
	"es_MX": "Use Mexican Spanish conventions.",
}

// StyleDescriptions maps TranslationStyle to human-readable descriptions for AI prompts.
var StyleDescriptions = map[TranslationStyle]string{
	StyleFormal:    "Use formal, professional language suitable for official documents or business communication.",
	StyleNeutral:   "Use a neutral, professional tone suitable for general web content and documentation.",
	StyleCasual:    "Use casual, conversational language suitable for blogs, social media, or friendly communication.",
	StyleMarketing: "Use persuasive, engaging language suitable for marketing copy, landing pages, and promotional content.",
	StyleTechnical: "Use precise, technical language suitable for developer documentation, API references, and technical guides.",
}

// GetLocaleClarification returns the locale-specific hint for a language code.
func GetLocaleClarification(langCode string) string {
	if hint, ok := LocaleClarifications[langCode]; ok {
		return hint
	}
	// Try normalized version
	normalized := NormalizeLocale(langCode)
	if hint, ok := LocaleClarifications[normalized]; ok {
		return hint
	}
	return ""
}

// GetStyleDescription returns the description for a translation style.
func GetStyleDescription(style TranslationStyle) string {
	if desc, ok := StyleDescriptions[style]; ok {
		return desc
	}
	return StyleDescriptions[StyleNeutral]
}
