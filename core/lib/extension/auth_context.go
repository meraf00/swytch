package extension

import (
	"context"
	"errors"
)

type authContextKey int

const (
	authContext authContextKey = iota
)

type AuthContext struct {
	Credentials struct {
		UID   string
		Scope []string
	}
	Artifacts struct {
		AccessToken string
	}
}

func WithAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContext, authCtx)
}

func GetAuthContext(ctx context.Context) (*AuthContext, error) {
	authCtx, ok := ctx.Value(authContext).(*AuthContext)
	if !ok || authCtx == nil {
		return nil, errors.New("unauthorized: missing auth context")
	}

	return authCtx, nil
}
