package trace

import "net/http"

// A ResponseStaller is used to save the status code of a written response, in
// order to be logged after the response has been written.
type ResponseStaller struct {
	w      http.ResponseWriter
	Status int
}

// NewResponseStaller returns a new response ResponseStaller from an
// http.ResponseWriter.
func NewResponseStaller(w http.ResponseWriter) *ResponseStaller {
	return &ResponseStaller{w, 0}
}

// Write calls the underlying http.ResponseWriter's `Write` method.
func (r *ResponseStaller) Write(b []byte) (int, error) {
	return r.w.Write(b)
}

// WriteHeader wraps the underlying http.ResponseWriter's `Write` method,
// saving the staus code in the `status` field.
func (r *ResponseStaller) WriteHeader(n int) {
	r.Status = n
	r.w.WriteHeader(n)
}

// Header calls the underlying http.ResponseWriter's `Header` method.
func (r *ResponseStaller) Header() http.Header {
	return r.w.Header()
}
