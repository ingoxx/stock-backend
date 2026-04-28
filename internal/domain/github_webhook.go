package domain

// GitHubTokenResponse 对应 GitHub 返回的 Token 结构
type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

type GithubCallbackRepository interface {
	GitHubCallBackApi() error
}
