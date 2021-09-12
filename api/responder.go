package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dhurimkelmendi/vending_machine/config"
	"github.com/dhurimkelmendi/vending_machine/internal/trace"
	"github.com/sirupsen/logrus"
)

// Responder is provides helpers for handling and responding to HTTP requests.
type Responder struct {
	cfg *config.Config
}

var responderDefaultInstance *Responder

// GetResponderDefaultInstance returns the default instance of Responder.
func GetResponderDefaultInstance() *Responder {
	if responderDefaultInstance == nil {
		responderDefaultInstance = &Responder{
			cfg: config.GetDefaultInstance(),
		}
	}
	return responderDefaultInstance
}

type errorResponse struct {
	Message    string  `json:"message"`
	Context    string  `json:"context"`
	RequestID  string  `json:"requestID"`
	Code       string  `json:"code"`
	InnerError *string `json:"innerError"`

	// Status code is not part of the response body.
	Status int `json:"-"`

	// Internally logged error is not part of the response body.
	LogErr error `json:"-"`
}

// Error sends the error message as JSON with the given HTTP status code.
func (r *Responder) Error(res http.ResponseWriter, err error, statuses ...int) {
	body, payloadStatus, err := r.commonError(&res, err, statuses...)
	if err != nil {
		logrus.Errorf("%s: Failed to generate JSON error response: %+v", trace.Getfl(), err)
	}

	res.WriteHeader(payloadStatus)

	_, err = res.Write(body)
	if err != nil {
		logrus.Errorf("%s: Error writing JSON error response: %+v", trace.Getfl(), err)
	}
}

// NoContent returns a response with status code 204
func (r *Responder) NoContent(res http.ResponseWriter) {
	res.WriteHeader(http.StatusNoContent)
}

//commonError returns error response body in json format
func (r *Responder) commonError(res *http.ResponseWriter, err error, statuses ...int) ([]byte, int, error) {
	payload := &errorResponse{LogErr: err}
	logrus.Errorf("(ERROR) %v", err)
	switch e := err.(type) {
	case *ResponseError:
		payload.Code = e.Code
		payload.Message = e.Message
		payload.Context = string(e.Context)
		payload.RequestID = e.ContextID
		payload.Status = e.Status

		if e.InnerError != nil {
			payload.LogErr = e.InnerError

			if r.cfg.RespondWithInnerError {
				innerErrorMessage := fmt.Sprintf("%+v", e.InnerError.Error())
				payload.InnerError = &innerErrorMessage
			}

			switch ie := e.InnerError.(type) {
			case *ResponseError:
				if ie.Component == CmpAuthentication || ie.Component == CmpSerializer || ie.Component == CmpService {
					if len(ie.Code) > 0 {
						payload.Code = ie.Code
					}

					if len(ie.Message) > 0 {
						payload.Message = ie.Message
					}

					if ie.Status > 0 {
						payload.Status = ie.Status
					}
				}
			}
		}

	case *json.SyntaxError:
		payload.Message = err.Error() + " at character " + strconv.Itoa(int(e.Offset))
		payload.Status = http.StatusBadRequest

	default:
		payload.Message = err.Error()
	}

	if payload.Status <= 0 {
		payload.Status = http.StatusInternalServerError
	}

	if len(statuses) > 0 && statuses[0] > 0 {
		payload.Status = statuses[0]
	}

	// Log the error internally
	logrus.Errorf("%s: %s: %+v", trace.Getfl(), payload.Message, payload.LogErr)

	(*res).Header().Set("Content-Type", "application/json")

	body, err := json.MarshalIndent(map[string]interface{}{"error": payload}, "", "  ")
	if err != nil {
		logrus.Errorf("%s: Failed to generate JSON error response: %+v", trace.Getfl(), err)
	}
	return body, payload.Status, err
}

// JSON writes the given response as JSON. It will respond with 200 OK, unless
// an optional status code is provided.
func (r *Responder) JSON(res http.ResponseWriter, req *http.Request, v interface{}, status ...int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		r.Error(res, err, http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	for _, s := range status {
		res.WriteHeader(s)
	}

	if _, err := buf.WriteTo(res); err != nil {
		logrus.Errorf("Error writing JSON response: %+v", err)
	}
}

// Text writes the given response as plain text. It will respond with 200 OK, unless an optional status code is provided.
func (r *Responder) Text(res http.ResponseWriter, req *http.Request, v []byte, status ...int) {
	res.Header().Set("Content-Type", "plain/text")
	for _, s := range status {
		res.WriteHeader(s)
	}

	if _, err := res.Write(v); err != nil {
		logrus.Errorf("Error writing text response: %+v", err)
	}
}

// Redirect sends to user to the provided URL.
func (r *Responder) Redirect(res http.ResponseWriter, req *http.Request, destination string, permanent bool) {
	code := http.StatusFound
	if permanent {
		code = http.StatusMovedPermanently
	}
	http.Redirect(res, req, destination, code)
}
