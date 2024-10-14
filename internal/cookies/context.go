package cookies

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/gophermart/internal/domain"
)

type ctxAuthKey string

const AuthKey ctxAuthKey = "auth"

func Context(parent context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(parent, AuthKey, uid)
}

func UIDFromContext(ctx context.Context) (*domain.UserID, bool) {
	val, ok := ctx.Value(AuthKey).(domain.UserID)
	if !ok || val == 0 {
		return nil, false
	}

	uid := val
	return &uid, true
}
