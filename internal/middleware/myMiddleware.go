package middleware

import (
	"context"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/cmd/server"
	"github.com/ingoxx/stock-backend/utils"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
)

// JWTAuthMiddleware 验证 Token 并注入 userID 到 Context 的中间件
func JWTAuthMiddleware(next http.Handler, rc map[int]*redis.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sign := r.FormValue("sign")
		if sign == "" {
			utils.ResponseJSON(w, utils.Response{
				Code: 401,
				Msg:  "未提供 Token，请先登录",
				Data: "",
			})
			return
		}

		user := r.FormValue("uid")
		if user == "" {
			utils.ResponseJSON(w, utils.Response{
				Code: 401,
				Msg:  "未提供用户，请先登录",
				Data: "",
			})
			return
		}

		mapLock.RLock()
		app := server.NewVerifyApp(rc)
		mapLock.RUnlock()

		if err := app.VerifyService.GetJwtToken(user, sign); err != nil {
			utils.ResponseJSON(w, utils.Response{
				Code: 401,
				Msg:  err.Error(),
				Data: "",
			})
			return
		}

		claims, err := utils.ParseToken(sign)
		if err != nil {
			if err := app.VerifyService.DelJwtToken(user); err != nil {
				utils.ResponseJSON(w, utils.Response{
					Code: 401,
					Msg:  err.Error(),
					Data: "",
				})

				return
			}

			utils.ResponseJSON(w, utils.Response{
				Code: 401,
				Msg:  "Token 无效或已过期",
				Data: "",
			})
			return
		}

		// 4. 解析成功：将 UserID 存入 Request Context 中
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		r = r.WithContext(ctx)

		// 5. 放行，继续执行后续 Handler
		next.ServeHTTP(w, r)
	})
}
