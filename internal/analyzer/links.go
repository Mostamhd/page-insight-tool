package analyzer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

var linkClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

type Link struct {
	URL        string
	IsInternal bool
}

var skipSchemes = map[string]bool{
	"mailto":     true,
	"javascript": true,
	"tel":        true,
}

func ClassifyLinks(doc *html.Node, baseURL *url.URL) []Link {
	var links []Link
	collectLinks(doc, baseURL, &links)
	return links
}

func collectLinks(n *html.Node, baseURL *url.URL, links *[]Link) {
	if n.Type == html.ElementNode && n.Data == "a" {
		if link, ok := extractLink(n, baseURL); ok {
			*links = append(*links, link)
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectLinks(c, baseURL, links)
	}
}

func extractLink(n *html.Node, baseURL *url.URL) (Link, bool) {
	var href string
	for _, a := range n.Attr {
		if a.Key == "href" {
			href = strings.TrimSpace(a.Val)
			break
		}
	}

	if href == "" || strings.HasPrefix(href, "#") {
		return Link{}, false
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return Link{}, false
	}

	if skipSchemes[parsed.Scheme] {
		return Link{}, false
	}

	resolved := baseURL.ResolveReference(parsed)

	return Link{
		URL:        resolved.String(),
		IsInternal: strings.EqualFold(resolved.Host, baseURL.Host),
	}, true
}

func CountInaccessibleLinks(ctx context.Context, links []Link, workers int) int {
	if len(links) == 0 {
		return 0
	}

	var (
		mu    sync.Mutex
		count int
		wg    sync.WaitGroup
	)
	sem := make(chan struct{}, workers)

	for _, l := range links {
		wg.Add(1)
		go func(rawURL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if !isAccessible(ctx, rawURL) {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}(l.URL)
	}

	wg.Wait()
	return count
}

func isAccessible(ctx context.Context, rawURL string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		return false
	}

	resp, err := linkClient.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.StatusCode < 400
}
