package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Route struct {
	Prefix  string
	Target  string
}

type Proxy struct {
	routes []Route
}

func NewProxy(routes []Route) *Proxy {
	return &Proxy{routes: routes}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range p.routes {
		if strings.HasPrefix(r.URL.Path, route.Prefix) {
			target, err := url.Parse(route.Target)
			if err != nil {
				http.Error(w, "bad gateway config", http.StatusInternalServerError)
				return
			}

			// strip the prefix before forwarding
			r.URL.Path = strings.TrimPrefix(r.URL.Path, route.Prefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}

			log.Printf("--> %s %s to %s", r.Method, r.URL.Path, route.Target)

			proxy := httputil.NewSingleHostReverseProxy(target)
			proxy.ServeHTTP(w, r)
			return
		}
	}

	http.Error(w, "no route matched", http.StatusNotFound)
}