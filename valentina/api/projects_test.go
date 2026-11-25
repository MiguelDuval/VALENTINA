package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProjectHTTPContract(t *testing.T) {
	s := testServer(t)
	mux := http.NewServeMux()
	for route, h := range s.Routes() {
		mux.HandleFunc(route, h)
	}
	call := func(method, path, body string, want int) map[string]any {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, "/api/v1/valentina"+path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		var result map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &result)
		return result
	}
	call("POST", "/projects", `{"title":"Synthetic","status":"active"}`, 400)
	call("POST", "/projects", `{"title":"Synthetic"} {}`, 400)
	p := call("POST", "/projects", `{"title":"Synthetic"}`, 201)
	id := p["project_id"].(string)
	call("GET", "/projects/"+id, "", 200)
	call("POST", "/projects/"+id+"/status", `{"status":"completed","version":1}`, 400)
	call("POST", "/projects/"+id+"/status", `{"status":"active cancelled","version":1}`, 400)
	call("POST", "/projects/"+id+"/status", `{"status":"active","version":1}`, 200)
	call("POST", "/projects/"+id+"/status", `{"status":"on_hold","version":1}`, 409)
	body := `{"base_revision":0,"lines":[{"kind":"labor","name":"Synthetic","quantity_milli":1000,"unit_cents":100,"vat_bps":2100}]}`
	call("POST", "/projects/"+id+"/estimates/preview", body, 200)
	e := call("POST", "/projects/"+id+"/estimates", body, 201)
	call("POST", "/projects/"+id+"/estimates", body, 409)
	call("POST", "/projects/"+id+"/estimates/"+e["estimate_id"].(string)+"/approve", `{}`, 200)
	call("GET", "/projects/"+id+"/estimates", "", 200)
	call("POST", "/projects/"+id+"/documents", `{"type":"note","name":"Synthetic","annotation":"Metadata only"}`, 201)
	call("POST", "/projects/"+id+"/documents", `{"type":"note","name":"C:\\private.txt"}`, 400)
	call("GET", "/projects/"+id+"/documents", "", 200)
	call("GET", "/projects/missing", "", 404)
	call("DELETE", "/projects/"+id, "", 405)
}
