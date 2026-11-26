package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValentinaBrowserLaunch(t *testing.T) {
	for _, mode := range []string{"ready", "unauthorized", "redirect", "wrong-service"} {
		t.Run(mode, func(t *testing.T) {
			home := t.TempDir()
			if err := os.WriteFile(filepath.Join(home, "server.token"), []byte("synthetic-browser-token\n"), 0600); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/valentina/licenses" || r.Header.Get("Authorization") != "Bearer synthetic-browser-token" {
					t.Error("wrong readiness request")
				}
				switch mode {
				case "unauthorized":
					w.WriteHeader(401)
				case "redirect":
					w.Header().Set("Location", "http://192.0.2.1/")
					w.WriteHeader(302)
				case "wrong-service":
					fmt.Fprint(w, `{"status":"ok"}`)
				default:
					fmt.Fprint(w, `{"upstream_commit":"f01c1d84c08b1de972d8cc268da25602598adff8"}`)
				}
			}))
			defer server.Close()
			port := server.Listener.Addr().(*net.TCPAddr).Port
			calls := 0
			for i := 0; i < 2; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
				err := valentinaWaitAndOpen(ctx, port, home, func(url string) error {
					calls++
					if url != fmt.Sprintf("http://127.0.0.1:%d", port) {
						t.Error("unsafe URL")
					}
					return nil
				})
				cancel()
				if (err == nil) != (mode == "ready") {
					t.Fatal("unexpected readiness", mode, err)
				}
			}
			if mode == "ready" && calls != 2 || mode != "ready" && calls != 0 {
				t.Fatal("wrong browser open count", calls)
			}
		})
	}
}
