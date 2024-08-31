package proxy

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type Option = func(*proxy)

type Proxy interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// New Proxy
func New(options ...func(*proxy)) Proxy {
	p := &proxy{}

	for _, fn := range options {
		fn(p)
	}

	p.upstream = httputil.NewSingleHostReverseProxy(p.target)
	p.upstream.Transport = &transport{http.DefaultTransport}

	return p
}

func Target(target string) func(*proxy) {
	return func(p *proxy) {
		u, err := url.Parse(target)
		if err != nil {
			panic(err)
		}

		p.target = u
	}
}

type proxy struct {
	target   *url.URL
	upstream *httputil.ReverseProxy
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	slog.Debug("serve-http", "method", r.Method, "url", r.URL.String())
	p.upstream.ServeHTTP(w, r)
}

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	slog.Debug("roundtrip-start", "method", req.Method, "url", req.URL.String())

	// Call the original RoundTripper
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		slog.Debug("roundtrip-error", "error", err)
		return nil, err
	}

	// Log the response
	slog.Debug("roundtrip-complete", "status", resp.Status)

	// Process and modify the response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Debug("roundtrip-read-error", "error", err)
		resp.Body.Close()
		return nil, err
	}
	resp.Body.Close()

	// Modify body
	bodyBytes = bytes.Replace(bodyBytes, []byte("</head>"), []byte(`<script type="text/javascript" src="/__dev-server/ws-live-reload.js"></script></head>`), 1)
	body := io.NopCloser(bytes.NewReader(bodyBytes))
	resp.Body = body
	resp.ContentLength = int64(len(bodyBytes))

	resp.Header.Set("Content-Length", strconv.Itoa(len(bodyBytes)))
	resp.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	resp.Header.Set("Pragma", "no-cache")
	resp.Header.Set("Expires", "0")

	return resp, nil
}
