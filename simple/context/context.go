package context

import (
	"context"
	"webapp1/simple/models"
)

const userKey privateKey = "user"

type privateKey string

func WithUser(ctx context.Context, user *models.Users) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func User(ctx context.Context) *models.Users {
	if temp := ctx.Value(userKey); temp != nil {
		if user, ok := temp.(*models.Users); ok {
			return user
		}
	}
	return nil
}
