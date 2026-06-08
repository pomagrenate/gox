package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

// DynamicProxy routes HTTP traffic to a dynamically hot-swappable backend.
type DynamicProxy struct {
	mu     sync.RWMutex
	target *url.URL
	proxy  *httputil.ReverseProxy
}

func NewDynamicProxy() *DynamicProxy {
	return &DynamicProxy{}
}

// SetTarget hot-swaps the underlying ReverseProxy target without dropping connections.
func (dp *DynamicProxy) SetTarget(targetURL string) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		return err
	}
	newProxy := httputil.NewSingleHostReverseProxy(u)

	dp.mu.Lock()
	dp.target = u
	dp.proxy = newProxy
	dp.mu.Unlock()
	return nil
}

func (dp *DynamicProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dp.mu.RLock()
	proxy := dp.proxy
	dp.mu.RUnlock()

	if proxy == nil {
		http.Error(w, "GoX is building the first version... Please wait.", http.StatusServiceUnavailable)
		return
	}
	proxy.ServeHTTP(w, r)
}
