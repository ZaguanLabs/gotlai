package gotlai

import "testing"

func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"es_ES", "Spanish (Spain)"},
		{"ja_JP", "Japanese (Japan)"},
		{"en", "English (United States)"}, // short code expansion
		{"unknown", "unknown"},            // fallback
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := GetLanguageName(tt.code)
			if result != tt.expected {
				t.Errorf("GetLanguageName(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestGetDirection(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"ar_SA", "rtl"},
		{"he_IL", "rtl"},
		{"fa_IR", "rtl"},
		{"ur_PK", "rtl"},
		{"ar", "rtl"}, // short code
		{"es_ES", "ltr"},
		{"en_US", "ltr"},
		{"ja_JP", "ltr"},
		{"zh_CN", "ltr"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := GetDirection(tt.code)
			if result != tt.expected {
				t.Errorf("GetDirection(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

func TestIsRTL(t *testing.T) {
	if !IsRTL("ar_SA") {
		t.Error("IsRTL(ar_SA) should be true")
	}
	if IsRTL("en_US") {
		t.Error("IsRTL(en_US) should be false")
	}
}

func TestNormalizeLocale(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"es-ES", "es_ES"},
		{"en-US", "en_US"},
		{"es_ES", "es_ES"}, // already normalized
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLocale(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeLocale(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToHTMLLang(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"es_ES", "es-ES"},
		{"en_US", "en-US"},
		{"es-ES", "es-ES"}, // already HTML format
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToHTMLLang(tt.input)
			if result != tt.expected {
				t.Errorf("ToHTMLLang(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetLocaleClarification(t *testing.T) {
	tests := []struct {
		code     string
		contains string
	}{
		{"nb_NO", "Bokmål"},
		{"nb", "Bokmål"},
		{"zh_CN", "Simplified"},
		{"zh_TW", "Traditional"},
		{"pt_BR", "Brazilian"},
		{"pt_PT", "European"},
		{"en_GB", "British"},
		{"es_ES", "Castilian"},
		{"es_MX", "Mexican"},
		{"unknown", ""}, // no clarification
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := GetLocaleClarification(tt.code)
			if tt.contains == "" {
				if result != "" {
					t.Errorf("GetLocaleClarification(%q) = %q, want empty", tt.code, result)
				}
			} else if result == "" || !contains(result, tt.contains) {
				t.Errorf("GetLocaleClarification(%q) = %q, want to contain %q", tt.code, result, tt.contains)
			}
		})
	}
}

func TestGetStyleDescription(t *testing.T) {
	tests := []struct {
		style    TranslationStyle
		contains string
	}{
		{StyleFormal, "formal"},
		{StyleNeutral, "neutral"},
		{StyleCasual, "casual"},
		{StyleMarketing, "persuasive"},
		{StyleTechnical, "technical"},
		{"", "neutral"}, // default to neutral
	}

	for _, tt := range tests {
		t.Run(string(tt.style), func(t *testing.T) {
			result := GetStyleDescription(tt.style)
			if !contains(result, tt.contains) {
				t.Errorf("GetStyleDescription(%q) = %q, want to contain %q", tt.style, result, tt.contains)
			}
		})
	}
}

func TestShortCodeToLocale_Comprehensive(t *testing.T) {
	// Test that common short codes resolve correctly
	tests := []struct {
		short string
		full  string
	}{
		{"en", "en_US"},
		{"gb", "en_GB"},
		{"de", "de_DE"},
		{"es", "es_ES"},
		{"mx", "es_MX"},
		{"fr", "fr_FR"},
		{"ja", "ja_JP"},
		{"jp", "ja_JP"},
		{"zh", "zh_CN"},
		{"tw", "zh_TW"},
		{"nb", "nb_NO"},
		{"no", "nb_NO"},
		{"pt", "pt_BR"},
		{"br", "pt_BR"},
	}

	for _, tt := range tests {
		t.Run(tt.short, func(t *testing.T) {
			result, ok := ShortCodeToLocale[tt.short]
			if !ok {
				t.Errorf("ShortCodeToLocale[%q] not found", tt.short)
				return
			}
			if result != tt.full {
				t.Errorf("ShortCodeToLocale[%q] = %q, want %q", tt.short, result, tt.full)
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
