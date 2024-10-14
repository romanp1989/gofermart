package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/gophermart/internal/api/balance"
	"github.com/romanp1989/gophermart/internal/api/order"
	"github.com/romanp1989/gophermart/internal/api/user"
)

func NewRoutes(u *user.Handler, o *order.Handler, b *balance.Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/api/user/register", u.RegisterHandler)
	r.Post("/api/user/login", u.LoginHandler)
	r.Post("/api/user/orders", o.CreateOrderHandler)
	r.Get("/api/user/orders", o.ListOrdersHandler)
	r.Get("/api/user/balance", b.GetBalanceHandler)
	r.Post("/api/user/balance/withdraw", b.WithdrawHandler)
	r.Get("/api/user/withdrawals", b.GetWithdrawHandler)
	return r
}
