package tfc

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type CloudAPIErrorError struct {
	Status string                 `json:"status"`
	Title  string                 `json:"title"`
	Detail string                 `json:"detail"`
	Source map[string]interface{} `json:"source"`
}

func (e *CloudAPIErrorError) Error() string {
	if e == nil {
		return ""
	}
	b, _ := json.Marshal(e.Source)
	return fmt.Sprintf("status=%s, title=%s, detail=%s, source=%s", e.Status, e.Title, e.Detail, string(b))
}

func UnwrapCloudAPIError(err error) *CloudAPIErrorError {
	for err != nil {
		if e, ok := err.(*CloudAPIErrorError); ok {
			return e
		}
		err = errors.Unwrap(err)
	}
	return nil
}

type APIError struct {
	Errors  []CloudAPIErrorError `json:"errors,omitempty"`
	Success *bool                `json:"success,omitempty"`
}

func (e *APIError) UnmarshalJSON(b []byte) error {
	type tmp struct {
		Success *bool           `json:"success"`
		Errors  json.RawMessage `json:"errors"`
	}

	t := new(tmp)
	if err := json.Unmarshal(b, t); err != nil {
		return err
	}

	// "errors"" is usually an array of objects, but because javascript and other horrible things exist, it can
	// sometimes be an array of strings.

	e.Errors = make([]CloudAPIErrorError, 0)

	// test to see if its a friggin array of friggin strings.
	if strings.HasPrefix(string(t.Errors), "{\"errors\":[\"") {
		strs := make([]string, 0)
		if err := json.Unmarshal(t.Errors, &strs); err != nil {
			return err
		}
		for _, str := range strs {
			e.Errors = append(e.Errors, CloudAPIErrorError{Title: str})
		}
		return nil
	}

	e.Errors = make([]CloudAPIErrorError, 0)
	return json.Unmarshal(t.Errors, &e.Errors)
}

func (e *APIError) Error() string {
	if e == nil || len(e.Errors) == 0 {
		return ""
	}

	out := fmt.Sprintf("%d errors:", len(e.Errors))
	for i, e := range e.Errors {
		if i > 0 {
			out = fmt.Sprintf("%s;", out)
		}
		out = fmt.Sprintf("%s%s", out, e.Error())
	}

	return out
}

func UnwrapAPIError(err error) *APIError {
	for err != nil {
		if e, ok := err.(*APIError); ok {
			return e
		}
		err = errors.Unwrap(err)
	}
	return nil
}

type StatusError struct {
	ExpectedCode int      `json:"expected-code"`
	ActualCode   int      `json:"actual-code"`
	Body         string   `json:"body"`
	CloudError   APIError `json:"tfcloud-error"`
}

func (e *StatusError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("expected response code %v, saw: code=%d; body=\"%s\"", e.ExpectedCode, e.ActualCode, e.Body)
}

func UnwrapStatusError(err error) *StatusError {
	for err != nil {
		if e, ok := err.(*StatusError); ok {
			return e
		}
		err = errors.Unwrap(err)
	}
	return nil
}
