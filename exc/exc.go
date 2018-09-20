package exc

import (
  "fmt"
  "errors"
)

type httpError struct {
  err error
  status_code int16
}

func (e *httpError) Error() string {
  return fmt.Sprintf("HttpError(%d): %s", e.status_code, e.err.Error())
}

var (
  ErrPageNotFound = httpError{err: errors.New("Page not found"), status_code: 404}
  ErrUnknown = httpError{err: errors.New("Unknown error"), status_code: 500}
)

func RaiseHttpExc(status_code int16, msg string) *httpError {
  return &httpError{status_code: status_code, err: errors.New(msg)}
}
