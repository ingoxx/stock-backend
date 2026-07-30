package middleware

import (
	"context"
	"net/http"

	"github.com/ingoxx/stock-backend/utils"
)

type contextKey string

const UserIDKey contextKey = "userID"

// JWTAuthMiddleware 验证 Token 并注入 userID 到 Context 的中间件
func JWTAuthMiddleware(next http.Handler) http.Handler {
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

		// 2. 检查格式是否为：Bearer <token>
		//parts := strings.SplitN(sign, " ", 2)
		//if !(len(parts) == 2 && parts[0] == "Bearer") {
		//	utils.ResponseJSON(w, utils.Response{
		//		Code: 401,
		//		Msg:  "Token 格式错误(格式应为: Bearer <token>)",
		//		Data: "",
		//	})
		//	return
		//}

		// 3. 校验并解析 Token
		claims, err := utils.ParseToken(sign)
		if err != nil {
			utils.ResponseJSON(w, utils.Response{
				Code: 401,
				Msg:  "Token 无效或已过期: " + err.Error(),
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
