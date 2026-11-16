package examples

import (
	"net/http"
	"net/http/httptest"
)

func recorder() *httptest.ResponseRecorder { return httptest.NewRecorder() }

func requestWith(rec *httptest.ResponseRecorder) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range rec.Result().Cookies() {
		req.AddCookie(c)
	}
	return req
}
