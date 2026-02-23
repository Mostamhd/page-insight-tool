package analyzer

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

type AnalyzeResponse struct {
	HTMLVersion       string         `json:"htmlVersion"`
	Title             string         `json:"title"`
	Headings          map[string]int `json:"headings"`
	InternalLinks     int            `json:"internalLinks"`
	ExternalLinks     int            `json:"externalLinks"`
	InaccessibleLinks int            `json:"inaccessibleLinks"`
	HasLoginForm      bool           `json:"hasLoginForm"`
}

const defaultWorkers = 10

func Analyze(ctx context.Context, rawHTML []byte, pageURL string) (*AnalyzeResponse, error) {
	doc, err := html.Parse(bytes.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	base, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("parsing page URL %s: %w", pageURL, err)
	}

	links := ClassifyLinks(doc, base)

	var internal, external int
	for _, l := range links {
		if l.IsInternal {
			internal++
		} else {
			external++
		}
	}

	inaccessible := CountInaccessibleLinks(ctx, links, defaultWorkers)

	return &AnalyzeResponse{
		HTMLVersion:       detectHTMLVersion(doc),
		Title:             extractTitle(doc),
		Headings:          countHeadings(doc),
		InternalLinks:     internal,
		ExternalLinks:     external,
		InaccessibleLinks: inaccessible,
		HasLoginForm:      hasLoginForm(doc),
	}, nil
}

func detectHTMLVersion(doc *html.Node) string {
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.DoctypeNode {
			return classifyDoctype(c)
		}
	}
	return "Unknown"
}

func classifyDoctype(n *html.Node) string {
	if len(n.Attr) == 0 {
		return "HTML5"
	}

	pub := strings.ToLower(n.Attr[0].Val)
	if pub == "" {
		return "HTML5"
	}

	switch {
	case strings.Contains(pub, "xhtml 1.1"):
		return "XHTML 1.1"
	case strings.Contains(pub, "xhtml 1.0"):
		if strings.Contains(pub, "strict") {
			return "XHTML 1.0 Strict"
		}
		if strings.Contains(pub, "frameset") {
			return "XHTML 1.0 Frameset"
		}
		return "XHTML 1.0 Transitional"
	case strings.Contains(pub, "html 4.01"):
		if strings.Contains(pub, "strict") {
			return "HTML 4.01 Strict"
		}
		if strings.Contains(pub, "frameset") {
			return "HTML 4.01 Frameset"
		}
		return "HTML 4.01 Transitional"
	default:
		return "Unknown"
	}
}

func extractTitle(doc *html.Node) string {
	n := findElement(doc, "title")
	if n == nil {
		return ""
	}
	return textContent(n)
}

func findElement(n *html.Node, tag string) *html.Node {
	if n.Type == html.ElementNode && n.Data == tag {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findElement(c, tag); found != nil {
			return found
		}
	}
	return nil
}

func textContent(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

var headingTags = map[string]bool{
	"h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true,
}

func countHeadings(doc *html.Node) map[string]int {
	counts := map[string]int{
		"h1": 0, "h2": 0, "h3": 0,
		"h4": 0, "h5": 0, "h6": 0,
	}
	countHeadingsRecursive(doc, counts)
	return counts
}

func countHeadingsRecursive(n *html.Node, counts map[string]int) {
	if n.Type == html.ElementNode && headingTags[n.Data] {
		counts[n.Data]++
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		countHeadingsRecursive(c, counts)
	}
}

func hasLoginForm(doc *html.Node) bool {
	return findLoginForm(doc)
}

func findLoginForm(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "form" {
		if formIsLogin(n) {
			return true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if findLoginForm(c) {
			return true
		}
	}
	return false
}

func formIsLogin(form *html.Node) bool {
	return findLoginInput(form)
}

func findLoginInput(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "input" {
		if isLoginInput(n) {
			return true
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if findLoginInput(c) {
			return true
		}
	}
	return false
}

func isLoginInput(n *html.Node) bool {
	for _, a := range n.Attr {
		switch a.Key {
		case "type":
			if strings.EqualFold(a.Val, "password") {
				return true
			}
		case "name", "id":
			lower := strings.ToLower(a.Val)
			if strings.Contains(lower, "password") || strings.Contains(lower, "login") {
				return true
			}
		}
	}
	return false
}
