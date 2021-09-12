package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ExpectStatusCode asserts that the response status code is set to the expectedStatusCode
func ExpectStatusCode(t *testing.T, res *httptest.ResponseRecorder, expectedStatusCode int) {
	if res.Code != expectedStatusCode {
		t.Fatalf("expected http status code %v - %s but was %v - %s -- response body: %+v", expectedStatusCode, http.StatusText(expectedStatusCode), res.Code, http.StatusText(res.Code), res.Body.String())
	}
}

// ExpectJson asserts that the response is json
func ExpectJson(t *testing.T, res *httptest.ResponseRecorder) {
	contentType := res.Header().Get("content-type")

	if contentType == "" || !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected response with json but content type was set to %v -- response body: %+v", contentType, res.Body.String())
	}
}
