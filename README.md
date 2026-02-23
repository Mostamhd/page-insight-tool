# Page Insight Tool

A web application that analyzes a given URL and extracts key insights: HTML version, page title, heading structure, link counts (internal/external/inaccessible), and login form detection.

Built with **Go** (backend) and **React + TypeScript + Vite** (frontend).

## Build & Run

### With Docker

```sh
docker build -t page-insight .
docker run -p 8080:8080 page-insight
```

### Without Docker

Start the Go backend:

```sh
go run ./cmd/server
```

In a separate terminal, start the frontend dev server:

```sh
cd frontend && npm install && npm run dev
```

Then open [http://localhost:5173](http://localhost:5173). The Vite dev server proxies API requests to the Go backend on port 8080.

### Configuration

| Flag      | Default | Description      |
|-----------|---------|------------------|
| `-port`   | `8080`  | HTTP listen port |

## API

### `POST /api/analyze`

**Request:**

```json
{ "url": "https://example.com" }
```

**Response:**

```json
{
  "htmlVersion": "HTML5",
  "title": "Example Domain",
  "headings": { "h1": 1, "h2": 0, "h3": 0, "h4": 0, "h5": 0, "h6": 0 },
  "internalLinks": 5,
  "externalLinks": 12,
  "inaccessibleLinks": 2,
  "hasLoginForm": false
}
```

## Assumptions & Design Decisions

- **React + TypeScript with Vite** for a component-based, type-safe frontend. The frontend runs as a separate dev server and proxies `/api` requests to the Go backend.
- **Login form detection** uses a heuristic: a `<form>` is considered a login form if it contains an `<input>` with `type="password"` or a `name`/`id` containing "password" or "login".
- **Link accessibility** is checked via concurrent HEAD requests.
- **Bounded concurrency** (~10 goroutines) for link checking to avoid overwhelming target servers, coordinated with a mutex and semaphore channel.
- **HTML version detection** inspects the DOCTYPE node's public identifier to classify HTML5, HTML 4.01, XHTML 1.0/1.1, or Unknown.

## Future Improvements

- Caching of analysis results.
- Rate limiting.
- Support for JavaScript-rendered pages (headless browser).
- Persistent storage of past analyses.
- `embed.FS` for single-binary deployment (embed frontend build output in the Go binary).
- Configurable fetch timeout via flag or environment variable.
