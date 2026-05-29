package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/utils"
)

type VerifyHandler struct {
	svc *service.VerifyService
	vd  *validator.Validate
}

type AuthReq struct {
	Sign string `json:"sign" validate:"required"`
}

func NewVerifyHandler(svc *service.VerifyService, vd *validator.Validate) *VerifyHandler {
	return &VerifyHandler{svc: svc, vd: vd}
}

func (au *VerifyHandler) Auth(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	var ssd AuthReq
	if err := json.Unmarshal(body, &ssd); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1002,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	if err := au.vd.Struct(ssd); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, e := range validationErrors {
				utils.ResponseJSON(w, StockResponse{
					Code: 1003,
					Msg:  fmt.Sprintf("required parameter '%s' is missing or empty.", e.Field()),
					Data: "",
				})
				return
			}
		}

		utils.ResponseJSON(w, StockResponse{
			Code: 1003,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	if err := au.svc.GetAuthData(ssd.Sign); err != nil {
		utils.ResponseJSON(w, StockResponse{
			Code: 1004,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "verification successful",
		Data: "",
	})
}
