package analyzer

import (
	"context"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func parseHTML(t *testing.T, s string) *html.Node {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}
	return doc
}

func TestAnalyze_SimpleHTMLNoLinks(t *testing.T) {
	rawHTML := []byte(`<!DOCTYPE html><html><head><title>Test</title></head><body><h1>Hello</h1></body></html>`)
	resp, err := Analyze(context.Background(), rawHTML, "http://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.HTMLVersion != "HTML5" {
		t.Errorf("HTMLVersion = %q, want %q", resp.HTMLVersion, "HTML5")
	}
	if resp.Title != "Test" {
		t.Errorf("Title = %q, want %q", resp.Title, "Test")
	}
	if resp.Headings["h1"] != 1 {
		t.Errorf("Headings[h1] = %d, want 1", resp.Headings["h1"])
	}
	if resp.InternalLinks != 0 {
		t.Errorf("InternalLinks = %d, want 0", resp.InternalLinks)
	}
	if resp.ExternalLinks != 0 {
		t.Errorf("ExternalLinks = %d, want 0", resp.ExternalLinks)
	}
	if resp.InaccessibleLinks != 0 {
		t.Errorf("InaccessibleLinks = %d, want 0", resp.InaccessibleLinks)
	}
	if resp.HasLoginForm {
		t.Error("HasLoginForm = true, want false")
	}
}

func TestAnalyze_InvalidURL(t *testing.T) {
	rawHTML := []byte(`<!DOCTYPE html><html></html>`)
	_, err := Analyze(context.Background(), rawHTML, "://bad")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "HTML5",
			html: `<!DOCTYPE html><html><head></head><body></body></html>`,
			want: "HTML5",
		},
		{
			name: "HTML 4.01 Strict",
			html: `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Strict//EN" "http://www.w3.org/TR/html4/strict.dtd"><html><head></head><body></body></html>`,
			want: "HTML 4.01 Strict",
		},
		{
			name: "HTML 4.01 Transitional",
			html: `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd"><html><head></head><body></body></html>`,
			want: "HTML 4.01 Transitional",
		},
		{
			name: "HTML 4.01 Frameset",
			html: `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Frameset//EN" "http://www.w3.org/TR/html4/frameset.dtd"><html><head></head><body></body></html>`,
			want: "HTML 4.01 Frameset",
		},
		{
			name: "XHTML 1.0 Strict",
			html: `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"><html><head></head><body></body></html>`,
			want: "XHTML 1.0 Strict",
		},
		{
			name: "XHTML 1.0 Transitional",
			html: `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"><html><head></head><body></body></html>`,
			want: "XHTML 1.0 Transitional",
		},
		{
			name: "XHTML 1.0 Frameset",
			html: `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Frameset//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-frameset.dtd"><html><head></head><body></body></html>`,
			want: "XHTML 1.0 Frameset",
		},
		{
			name: "XHTML 1.1",
			html: `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd"><html><head></head><body></body></html>`,
			want: "XHTML 1.1",
		},
		{
			name: "Unknown - no doctype",
			html: `<html><head></head><body></body></html>`,
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			got := detectHTMLVersion(doc)
			if got != tt.want {
				t.Errorf("detectHTMLVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "page with title",
			html: `<!DOCTYPE html><html><head><title>My Page</title></head><body></body></html>`,
			want: "My Page",
		},
		{
			name: "page without title",
			html: `<!DOCTYPE html><html><head></head><body></body></html>`,
			want: "",
		},
		{
			name: "empty title tag",
			html: `<!DOCTYPE html><html><head><title></title></head><body></body></html>`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			got := extractTitle(doc)
			if got != tt.want {
				t.Errorf("extractTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCountHeadings(t *testing.T) {
	t.Run("mixed headings", func(t *testing.T) {
		doc := parseHTML(t, `<!DOCTYPE html><html><body>
			<h1>Title</h1>
			<h2>Sub 1</h2>
			<h2>Sub 2</h2>
			<h3>Sub-sub</h3>
			<h4>Deep</h4>
			<h5>Deeper</h5>
			<h6>Deepest</h6>
			<h6>Deepest 2</h6>
		</body></html>`)
		counts := countHeadings(doc)
		expected := map[string]int{"h1": 1, "h2": 2, "h3": 1, "h4": 1, "h5": 1, "h6": 2}
		for tag, want := range expected {
			if counts[tag] != want {
				t.Errorf("countHeadings()[%q] = %d, want %d", tag, counts[tag], want)
			}
		}
	})

	t.Run("no headings", func(t *testing.T) {
		doc := parseHTML(t, `<!DOCTYPE html><html><body><p>No headings here</p></body></html>`)
		counts := countHeadings(doc)
		for _, tag := range []string{"h1", "h2", "h3", "h4", "h5", "h6"} {
			if counts[tag] != 0 {
				t.Errorf("countHeadings()[%q] = %d, want 0", tag, counts[tag])
			}
		}
	})
}

func TestHasLoginForm(t *testing.T) {
	tests := []struct {
		name string
		html string
		want bool
	}{
		{
			name: "form with type=password input",
			html: `<html><body><form><input type="text" name="user"><input type="password" name="pass"></form></body></html>`,
			want: true,
		},
		{
			name: "form with name containing password",
			html: `<html><body><form><input type="text" name="user_password_field"></form></body></html>`,
			want: true,
		},
		{
			name: "form with name containing login",
			html: `<html><body><form><input type="text" name="login_field"></form></body></html>`,
			want: true,
		},
		{
			name: "form without login/password inputs",
			html: `<html><body><form><input type="text" name="search"><input type="submit"></form></body></html>`,
			want: false,
		},
		{
			name: "no form at all",
			html: `<html><body><p>No form here</p></body></html>`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := parseHTML(t, tt.html)
			got := hasLoginForm(doc)
			if got != tt.want {
				t.Errorf("hasLoginForm() = %v, want %v", got, tt.want)
			}
		})
	}
}
