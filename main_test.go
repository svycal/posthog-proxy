package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", w.Body.String())
	}
}

func TestProxyDirector(t *testing.T) {
	target, _ := url.Parse("https://us.i.posthog.com")
	proxy := createProxy(target, "us.i.posthog.com")

	tests := []struct {
		name              string
		path              string
		incomingIP        string
		xForwardedFor     string
		acceptEncoding    string
		wantHost          string
		wantXForwardedFor string
	}{
		{
			name:              "sets host header",
			path:              "/capture",
			wantHost:          "us.i.posthog.com",
			wantXForwardedFor: "",
		},
		{
			name:              "forwards existing X-Forwarded-For",
			path:              "/capture",
			xForwardedFor:     "203.0.113.195",
			wantHost:          "us.i.posthog.com",
			wantXForwardedFor: "203.0.113.195",
		},
		{
			name:              "sets X-Forwarded-For from RemoteAddr",
			path:              "/capture",
			incomingIP:        "198.51.100.178:12345",
			wantHost:          "us.i.posthog.com",
			wantXForwardedFor: "198.51.100.178",
		},
		{
			name:           "removes Accept-Encoding",
			path:           "/capture",
			acceptEncoding: "gzip, deflate",
			wantHost:       "us.i.posthog.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.path, nil)
			if tt.incomingIP != "" {
				req.RemoteAddr = tt.incomingIP
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			proxy.Director(req)

			if req.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", req.Host, tt.wantHost)
			}
			if tt.wantXForwardedFor != "" && req.Header.Get("X-Forwarded-For") != tt.wantXForwardedFor {
				t.Errorf("X-Forwarded-For = %q, want %q", req.Header.Get("X-Forwarded-For"), tt.wantXForwardedFor)
			}
			if tt.acceptEncoding != "" && req.Header.Get("Accept-Encoding") != "" {
				t.Errorf("Accept-Encoding should be removed, got %q", req.Header.Get("Accept-Encoding"))
			}
		})
	}
}

func TestRouting(t *testing.T) {
	// Create mock backends
	apiCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	assetsCalled := false
	assetsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assetsCalled = true
		w.WriteHeader(http.StatusOK)
	}))
	defer assetsServer.Close()

	apiURL, _ := url.Parse(apiServer.URL)
	assetsURL, _ := url.Parse(assetsServer.URL)

	apiProxy := createProxy(apiURL, apiURL.Host)
	assetsProxy := createProxy(assetsURL, assetsURL.Host)

	mux := http.NewServeMux()
	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		assetsProxy.ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		apiProxy.ServeHTTP(w, r)
	})

	tests := []struct {
		path            string
		wantAPICalled   bool
		wantAssetCalled bool
	}{
		{"/capture", true, false},
		{"/batch", true, false},
		{"/decide", true, false},
		{"/static/recorder.js", false, true},
		{"/static/array.js", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			apiCalled = false
			assetsCalled = false

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if apiCalled != tt.wantAPICalled {
				t.Errorf("API called = %v, want %v", apiCalled, tt.wantAPICalled)
			}
			if assetsCalled != tt.wantAssetCalled {
				t.Errorf("Assets called = %v, want %v", assetsCalled, tt.wantAssetCalled)
			}
		})
	}
}
