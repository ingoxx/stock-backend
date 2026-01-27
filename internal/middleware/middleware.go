package middleware

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/cmd/server"
	cusErr "github.com/ingoxx/stock-backend/internal/error"
	"github.com/rs/cors"
	"net/http"
)

func AuthMiddleware(next http.Handler, rc map[int]*redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()
		sign := queryParams.Get("sign")

		app := server.NewVerifyApp(rc)
		if err := app.VerifyService.GetAuthData(sign); err != nil {
			var resp = server.VerifyResp{
				Code: 20001,
				Msg:  cusErr.AuthError.Error(),
			}

			b, err := json.Marshal(&resp)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			if _, err := w.Write(b); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			return

		}

		next.ServeHTTP(w, r)
	})
}

func AllowCorsMiddleware(next http.Handler) http.Handler {
	c := cors.New(cors.Options{
		// 允许的域名列表
		AllowedOrigins: []string{"http://localhost:8080", "http://localhost:11806", "http://localhost"},
		// 允许的方法
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		// 允许的 Header
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		// 是否允许 Cookie
		AllowCredentials: true,
		// 开启调试模式（会在控制台打印 CORS 日志）
		Debug: false,
	})

	handler := c.Handler(next)

	return handler
}
