package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/ingoxx/stock-backend/configs"
	"github.com/ingoxx/stock-backend/internal/service"
	"github.com/ingoxx/stock-backend/utils"
)

type GithubCallBackHandler struct {
	svc *service.GithubCallBackService
	vd  *validator.Validate
}

type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

func NewGithubCallBackHandler(svc *service.GithubCallBackService, vd *validator.Validate) *GithubCallBackHandler {
	return &GithubCallBackHandler{svc: svc, vd: vd}
}

func (sh *GithubCallBackHandler) GithubCallBackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	clientID := configs.AppId
	clientSecret := configs.ClientId

	jsonData := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	body, _ := json.Marshal(jsonData)

	req, _ := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to get token", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 3. 解析 Token
	var tokenResp GitHubTokenResponse
	json.NewDecoder(resp.Body).Decode(&tokenResp)

	// 4. 使用 Token 获取用户信息或进行后续操作
	fmt.Fprintf(w, "Login Success! Your Access Token: %s", tokenResp.AccessToken)

	utils.ResponseJSON(w, StockResponse{
		Code: 1000,
		Msg:  "ok",
		Data: "",
	})
}
