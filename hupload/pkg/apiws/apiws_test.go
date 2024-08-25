package apiws

import (
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ybizeul/hupload/pkg/apiws/middleware/auth"
)

type APIResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(APIResult{Status: "error", Message: msg})
}

func writeSuccessJSON(w http.ResponseWriter, body any) {
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
	}
}

func makeAPI(staticUI fs.FS, templateData any) *APIWS {
	api, err := New(staticUI, templateData)
	if err != nil {
		slog.Error("Error creating APIWS", slog.String("error", err.Error()))
	}

	if api == nil || api.Mux == nil {
		slog.Error("New() returned nil")
	}

	return api
}
func TestSimpleAPI(t *testing.T) {
	api := makeAPI(nil, nil)

	api.AddRoute("GET /", nil, func(w http.ResponseWriter, r *http.Request) {
		writeSuccessJSON(w, map[string]string{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	api.Mux.ServeHTTP(w, req)

	if w.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}
}

type testAuth struct {
	Username string
	Password string
}

func (a *testAuth) AuthenticateUser(username, password string) (bool, error) {
	return a.Username == username && a.Password == password, nil
}

func TestAuthAPI(t *testing.T) {
	authenticator := auth.BasicAuthMiddleware{
		Authentication: &testAuth{
			Username: "admin",
			Password: "password",
		},
	}

	api := makeAPI(nil, nil)

	api.AddRoute("GET /", []auth.AuthMiddleware{authenticator}, func(w http.ResponseWriter, r *http.Request) {
		writeSuccessJSON(w, map[string]string{"status": "ok"})
	})

	var (
		req *http.Request
		w   *httptest.ResponseRecorder
	)

	// Test with no authentication
	req = httptest.NewRequest("GET", "/", nil)

	w = httptest.NewRecorder()

	api.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Unexpected response code: %d", w.Code)
	}

	// Test with authentication
	req = httptest.NewRequest("GET", "/", nil)
	req.SetBasicAuth("admin", "password")

	w = httptest.NewRecorder()

	api.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected response code: %d", w.Code)
	}

	if w.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}
}

func TestTemplate(t *testing.T) {
	statucUI := os.DirFS("apiws_testdata")
	api := makeAPI(statucUI, struct{ Title string }{Title: "My Wonderful API"})

	var (
		req *http.Request
		w   *httptest.ResponseRecorder
	)

	// Test template on existing page
	req = httptest.NewRequest("GET", "/page.html", nil)

	w = httptest.NewRecorder()

	api.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected response code: %d", w.Code)
	}

	if w.Body.String() != "My Wonderful API" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}

	// Test template on non-existing page
	req = httptest.NewRequest("GET", "/random.html", nil)

	w = httptest.NewRecorder()

	api.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected response code: %d", w.Code)
	}

	if w.Body.String() != "index My Wonderful API" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}

	// Test template on non-html content
	req = httptest.NewRequest("GET", "/nothtml.txt", nil)

	w = httptest.NewRecorder()

	api.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Unexpected response code: %d", w.Code)
	}

	if w.Body.String() != "{{.Title}}" {
		t.Errorf("Unexpected response: %s", w.Body.String())
	}

}
