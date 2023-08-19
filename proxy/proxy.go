package proxy

import (
	"bytes"
	"io"
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
	p.upstream.ServeHTTP(w, r)
}

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	b = bytes.Replace(b, []byte(`</head>`), []byte(`<script type="text/javascript" src="/dev-server/live-reload.js"></script></head>`), 1)

	body := io.NopCloser(bytes.NewReader(b))
	resp.Body = body
	resp.ContentLength = int64(len(b))

	resp.Header.Set(`Content-Length`, strconv.Itoa(len(b)))
	resp.Header.Set(`Cache-Control`, `no-cache, no-store, must-revalidate`)
	resp.Header.Set(`Pragma`, `no-cache`)
	resp.Header.Set(`Expires`, `0`)

	return resp, nil
}
