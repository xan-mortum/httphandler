package httphandler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHttpHandlerServeHTTP(t *testing.T) {
	ctx := context.Background()
	testUrls := []string{
		"https://google.com",
		"https://amazon.com",
		"https://facebook.com",
	}
	urls := strings.Join(testUrls, "\n")
	req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBufferString(urls))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := NewHTTPHandler()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			rr.Code, http.StatusOK)
	}

	lines := strings.Split(rr.Body.String(), "\n")
	if len(lines) != len(testUrls) {
		t.Errorf("response length is %d, %d expected", len(lines), len(testUrls))
	}
}

func TestHttpHandlerServeHTTPManyConnections(t *testing.T) {
	ctx := context.Background()
	testUrls := []string{
		"https://google.com",
	}
	urls := strings.Join(testUrls, "\n")
	handler := NewHTTPHandler()

	for i := 1; i < maxConnection; i++ {
		go func(t *testing.T) {
			req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBufferString(urls))
			if err != nil {
				t.Errorf(err.Error())
			}
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					rr.Code, http.StatusOK)
			}
		}(t)
	}
	go func() {
		req, err := http.NewRequestWithContext(ctx, "POST", "/", bytes.NewBufferString(urls))
		if err != nil {
			t.Errorf(err.Error())
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusServiceUnavailable {
			t.Errorf("handler returned wrong status code: got %v want %v",
				rr.Code, http.StatusServiceUnavailable)
		}
	}()
}
