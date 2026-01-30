package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func main() {
	region := os.Getenv("POSTHOG_REGION")
	if region == "" {
		region = "us"
	}

	apiHost := region + ".i.posthog.com"
	assetsHost := region + "-assets.i.posthog.com"

	apiURL, _ := url.Parse("https://" + apiHost)
	assetsURL, _ := url.Parse("https://" + assetsHost)

	apiProxy := createProxy(apiURL, apiHost)
	assetsProxy := createProxy(assetsURL, assetsHost)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		assetsProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		apiProxy.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting PostHog proxy on :%s (region: %s)", port, region)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func createProxy(target *url.URL, host string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)

		// Set Host header to PostHog domain (critical for authentication)
		r.Host = host

		// Forward client IP
		if clientIP := r.Header.Get("X-Forwarded-For"); clientIP != "" {
			r.Header.Set("X-Forwarded-For", clientIP)
		} else if r.RemoteAddr != "" {
			ip := strings.Split(r.RemoteAddr, ":")[0]
			r.Header.Set("X-Forwarded-For", ip)
		}

		// Remove Accept-Encoding to avoid decompression issues
		r.Header.Del("Accept-Encoding")
	}

	return proxy
}
