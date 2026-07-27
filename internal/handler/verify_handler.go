package handler

import (
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
		utils.ResponseJSON(w, Response{
			Code: 1001,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	req, err := bindAndValidate[AuthReq](body, au.vd, func(r *AuthReq) {})
	if err != nil {
		writeReqError(w, err)
		return
	}

	if err := au.svc.GetAuthData(req.Sign); err != nil {
		utils.ResponseJSON(w, Response{
			Code: 1004,
			Msg:  err.Error(),
			Data: "",
		})
		return
	}

	utils.ResponseJSON(w, Response{
		Code: 1000,
		Msg:  "verification successful",
		Data: "",
	})
}
