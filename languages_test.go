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
