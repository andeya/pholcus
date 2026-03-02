package agent

import (
	"strings"
	"testing"
)

func TestUserAgentsPopulated(t *testing.T) {
	if len(UserAgents["all"]) == 0 {
		t.Fatal("UserAgents[all] should not be empty after init")
	}
	if len(UserAgents["common"]) == 0 {
		t.Fatal("UserAgents[common] should not be empty after init")
	}
	if len(UserAgents["all"]) < len(UserAgents["common"]) {
		t.Error("all should have at least as many agents as common")
	}
}

func TestCreateReal(t *testing.T) {
	ua := CreateReal()
	if ua == "" {
		t.Error("CreateReal() returned empty string")
	}
	if !strings.Contains(ua, "Surfer") {
		t.Errorf("CreateReal() = %q, want to contain 'Surfer'", ua)
	}
}

func TestCreateDefault(t *testing.T) {
	browsers := []string{"Chrome", "Firefox", "Edge", "Safari"}
	for _, b := range browsers {
		t.Run(b, func(t *testing.T) {
			ua := CreateDefault(b)
			if ua == "" {
				t.Errorf("CreateDefault(%q) returned empty string", b)
			}
		})
	}
}

func TestCreateVersion(t *testing.T) {
	ua := CreateVersion("Chrome", "127.0.6533.73")
	if ua == "" {
		t.Error("CreateVersion returned empty string")
	}
	if !strings.Contains(ua, "Chrome/127.0.6533.73") {
		t.Errorf("CreateVersion = %q, want to contain 'Chrome/127.0.6533.73'", ua)
	}
}

func TestTopVersion(t *testing.T) {
	tests := []struct {
		browser string
		want    string
	}{
		{"chrome", "127.0.6533.73"},
		{"firefox", "127.0"},
		{"unknown", "1.0"},
	}
	for _, tt := range tests {
		t.Run(tt.browser, func(t *testing.T) {
			got := TopVersion(tt.browser)
			if got != tt.want {
				t.Errorf("TopVersion(%q) = %q, want %q", tt.browser, got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	f := Format("chrome", "127.0")
	if f == "" {
		t.Error("Format(chrome, 127.0) returned empty string")
	}
	if !strings.Contains(f, "Chrome/{{.Ver}}") {
		t.Errorf("Format = %q, want to contain template var", f)
	}

	f2 := Format("unknown_browser", "1.0")
	if f2 == "" {
		t.Error("Format for unknown browser should return default format")
	}
}

func TestFormatFallbackToTopVersion(t *testing.T) {
	f := Format("chrome", "999.0")
	if f == "" {
		t.Error("Format with unknown version should fall back to top version format")
	}
}

func TestDatabaseContainsExpectedBrowsers(t *testing.T) {
	expected := []string{"chrome", "firefox", "edge", "safari", "googlebot", "bingbot", "yahoobot", "default"}
	for _, b := range expected {
		if _, ok := Database[b]; !ok {
			t.Errorf("Database missing browser %q", b)
		}
	}
}

func TestDefaultOSAttributes(t *testing.T) {
	for _, osID := range []int{Windows, Linux, Macintosh} {
		attr, ok := DefaultOSAttributes[osID]
		if !ok {
			t.Errorf("DefaultOSAttributes missing OS %d", osID)
			continue
		}
		if attr.OSName == "" {
			t.Errorf("OS %d has empty OSName", osID)
		}
	}
}
