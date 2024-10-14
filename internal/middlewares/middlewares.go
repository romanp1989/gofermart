package middlewares

import (
	"github.com/romanp1989/gophermart/internal/api/user"
)

type Middlewares struct {
	userHandler *user.Handler
}

func New(userHandler *user.Handler) *Middlewares {
	return &Middlewares{userHandler: userHandler}
}

//func (m *Middlewares) AuthMiddleware(h http.HandlerFunc) http.HandlerFunc {
//	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
//
//		ctx := request.Context()
//
//		cookie, err := request.Cookie("Token")
//		if err != nil {
//			m.userHandler.Log.Log.Error(err)
//			writer.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//
//		key, err := m.userHandler.GetKey(ctx, cookie.Value)
//		if err != nil || len(key) == 0 {
//			m.userHandler.Log.Log.Error(err)
//			writer.WriteHeader(http.StatusInternalServerError)
//			return
//		}
//
//		if ok := cookies.Validation(cookie.Value, key); !ok {
//			m.userHandler.Log.Log.Info("unauthorized user")
//			writer.WriteHeader(http.StatusUnauthorized)
//			return
//		}
//
//		h.ServeHTTP(writer, request.WithContext(ctx))
//	})
//
//}
