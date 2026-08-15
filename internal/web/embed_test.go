package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerServesIndexWithoutRedirect(t *testing.T) {
	for _, route := range []string{"/", "/topologia"} {
		request := httptest.NewRequest(http.MethodGet, route, nil)
		response := httptest.NewRecorder()
		Handler().ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Header().Get("Location") != "" {
			t.Fatalf("route %s returned %d location=%q", route, response.Code, response.Header().Get("Location"))
		}
		if !strings.Contains(response.Body.String(), `id="root"`) {
			t.Fatalf("route %s did not serve the application shell", route)
		}
	}
}
