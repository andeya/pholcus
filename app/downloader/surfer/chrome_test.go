package surfer

import (
	"io"
	"strings"
	"testing"
)

func TestChromeDownloaderBaiduSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Chrome test in short mode")
	}
	chrome := NewChrome()
	req := &DefaultRequest{
		URL:          "https://www.baidu.com/s?wd=golang&pn=0",
		DownloaderID: ChromeID,
	}

	r := chrome.Download(req)
	if r.IsErr() {
		t.Fatalf("download error: %v", r.UnwrapErr())
	}

	resp := r.Unwrap()
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status: %d %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body error: %v", err)
	}
	body := string(bodyBytes)
	t.Logf("body length: %d", len(body))

	hasResults := strings.Contains(body, `class="result`) || strings.Contains(body, `class="c-container`)
	if !hasResults {
		t.Fatal("No search result elements found in page")
	}
	t.Log("SUCCESS: first request returned Baidu search results")
}

// Verify that multiple requests to the same domain all succeed,
// not just the first one.
func TestChromeDownloaderReuseSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Chrome test in short mode")
	}
	chrome := NewChrome()

	keywords := []string{"golang", "pholcus"}
	for i, kw := range keywords {
		req := &DefaultRequest{
			URL:          "https://www.baidu.com/s?wd=" + kw + "&pn=0",
			DownloaderID: ChromeID,
		}
		r := chrome.Download(req)
		if r.IsErr() {
			t.Fatalf("request %d (%s) download error: %v", i+1, kw, r.UnwrapErr())
		}

		resp := r.Unwrap()
		bodyBytes, _ := io.ReadAll(resp.Body)
		body := string(bodyBytes)
		t.Logf("request %d (%s): status=%d body_length=%d", i+1, kw, resp.StatusCode, len(body))

		hasResults := strings.Contains(body, `class="result`) || strings.Contains(body, `class="c-container`)
		if !hasResults {
			t.Fatalf("request %d (%s): no search results found", i+1, kw)
		}
		t.Logf("request %d (%s): OK", i+1, kw)
	}
}

func TestExtractHomepage(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://www.baidu.com/s?wd=go", "https://www.baidu.com"},
		{"https://www.google.com/search?q=go", "https://www.google.com"},
		{"http://example.com", "http://example.com"},
		{"http://example.com/path/page", "http://example.com"},
		{"invalid-url", ""},
	}
	for _, tt := range tests {
		got := ExtractHomepage(tt.input)
		if got != tt.want {
			t.Errorf("extractHomepage(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
