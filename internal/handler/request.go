package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/utils"
)

type RequestError struct {
	Code int
	Msg  string
}

func (e *RequestError) Error() string { return e.Msg }

func bindAndValidate[T any](body []byte, vd *validator.Validate, normalize func(*T)) (T, error) {
	var req T

	if err := json.Unmarshal(body, &req); err != nil {
		return req, &RequestError{Code: 1002, Msg: err.Error()}
	}

	if normalize != nil {
		normalize(&req)
	}

	if err := vd.Struct(req); err != nil {
		var verr validator.ValidationErrors
		if errors.As(err, &verr) && len(verr) > 0 {
			return req, &RequestError{
				Code: 1003,
				Msg:  fmt.Sprintf("required parameter '%s' is missing or empty.", verr[0].Field()),
			}
		}
		return req, &RequestError{Code: 1003, Msg: err.Error()}
	}

	return req, nil
}

func writeReqError(w http.ResponseWriter, err error) {
	var re *RequestError
	ok := errors.As(err, &re)
	if !ok {
		re = &RequestError{Code: 1003, Msg: err.Error()}
	}

	utils.ResponseJSON(w, StockResponse{
		Code: re.Code,
		Msg:  re.Msg,
		Data: "",
	})
}
