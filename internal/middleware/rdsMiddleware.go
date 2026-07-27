package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/cmd/server"
	cusErr "github.com/ingoxx/stock-backend/internal/error"
	"github.com/ingoxx/stock-backend/utils"
	"github.com/rs/cors"
)

var mapLock sync.RWMutex

func AuthMiddleware(next http.Handler, rc map[int]*redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. 打印详细日志，用于排查请求方法是否被改写
		ip := utils.GetClientIP(r)

		log.Printf("recv [%s] request [%s] %s\n", ip, r.Method, r.URL.String())

		// 2. 优化参数获取：FormValue 可以同时获取 URL 中的 ?sign=... 和 POST 表单中的 sign
		sign := r.FormValue("sign")

		mapLock.RLock()
		app := server.NewVerifyApp(rc)
		mapLock.RUnlock()

		if err := app.VerifyService.GetAuthData(sign); err != nil {
			var resp = server.VerifyResp{
				Code: 403,
				Msg:  cusErr.AuthError.Error(),
			}

			b, err := json.Marshal(&resp)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			// 5. 修复隐患：必须在 Write 之前显式设置 Content-Type 和 403 状态码
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)

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
		AllowedOrigins: []string{"*"},
		// 允许的方法
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions},
		// 允许的 Header
		AllowedHeaders: []string{"*"},
		// 是否允许 Cookie
		AllowCredentials: true,
		// 开启调试模式（会在控制台打印 CORS 日志）
		Debug: false,
	})

	handler := c.Handler(next)

	return handler
}
