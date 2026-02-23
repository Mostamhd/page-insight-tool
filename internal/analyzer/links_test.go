package analyzer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", rawURL, err)
	}
	return u
}

func TestClassifyLinks(t *testing.T) {
	base := mustParseURL(t, "http://example.com/page")

	t.Run("internal and external links", func(t *testing.T) {
		doc := parseHTML(t, `<html><body>
			<a href="http://example.com/about">About</a>
			<a href="http://other.com/page">Other</a>
		</body></html>`)
		links := ClassifyLinks(doc, base)
		if len(links) != 2 {
			t.Fatalf("got %d links, want 2", len(links))
		}
		if !links[0].IsInternal {
			t.Error("first link should be internal")
		}
		if links[1].IsInternal {
			t.Error("second link should be external")
		}
	})

	t.Run("relative links resolved against base", func(t *testing.T) {
		doc := parseHTML(t, `<html><body>
			<a href="/contact">Contact</a>
			<a href="sub">Sub</a>
		</body></html>`)
		links := ClassifyLinks(doc, base)
		if len(links) != 2 {
			t.Fatalf("got %d links, want 2", len(links))
		}
		if links[0].URL != "http://example.com/contact" {
			t.Errorf("link[0].URL = %q, want %q", links[0].URL, "http://example.com/contact")
		}
		if !links[0].IsInternal {
			t.Error("relative link /contact should be internal")
		}
		if links[1].URL != "http://example.com/sub" {
			t.Errorf("link[1].URL = %q, want %q", links[1].URL, "http://example.com/sub")
		}
	})

	t.Run("skip mailto javascript tel and fragment-only", func(t *testing.T) {
		doc := parseHTML(t, `<html><body>
			<a href="mailto:test@example.com">Mail</a>
			<a href="javascript:void(0)">JS</a>
			<a href="tel:+1234567890">Phone</a>
			<a href="#section">Section</a>
			<a href="">Empty</a>
			<a href="http://example.com/valid">Valid</a>
		</body></html>`)
		links := ClassifyLinks(doc, base)
		if len(links) != 1 {
			t.Fatalf("got %d links, want 1 (only the valid one)", len(links))
		}
		if links[0].URL != "http://example.com/valid" {
			t.Errorf("link.URL = %q, want %q", links[0].URL, "http://example.com/valid")
		}
	})
}

func TestCountInaccessibleLinks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	links := []Link{
		{URL: ts.URL + "/ok", IsInternal: true},
		{URL: ts.URL + "/notfound", IsInternal: true},
		{URL: ts.URL + "/error", IsInternal: false},
	}

	count := CountInaccessibleLinks(context.Background(), links, 2)
	if count != 2 {
		t.Errorf("CountInaccessibleLinks = %d, want 2", count)
	}
}

func TestCountInaccessibleLinks_Empty(t *testing.T) {
	count := CountInaccessibleLinks(context.Background(), nil, 2)
	if count != 0 {
		t.Errorf("CountInaccessibleLinks(nil) = %d, want 0", count)
	}
}
