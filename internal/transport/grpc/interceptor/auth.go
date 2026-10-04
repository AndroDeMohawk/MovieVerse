package interceptor

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const UserIDKey ctxKey = "user_id"

type CustomClaims struct {
	jwt.RegisteredClaims
	AppID int32  `json:"app_id"`
	Email string `json:"email"`
	UID   int64  `json:"uid"`
}

// AuthUnaryInterceptor создаёт UnaryInterceptor для валидации JWT токена
func AuthUnaryInterceptor(secret string, expectedAppID int32) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		tokenStr, err := extractTokenFromMetadata(ctx)
		if err != nil {

			return handler(ctx, req)
		}

		claims, err := parseToken(tokenStr, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization token")
		}

		if claims.AppID != expectedAppID {
			return nil, status.Error(codes.Unauthenticated, "token belongs to another application")
		}

		ctx = context.WithValue(ctx, UserIDKey, claims.UID)

		return handler(ctx, req)
	}
}

func parseToken(tokenStr, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	return claims, nil
}

func extractTokenFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("missing metadata")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", fmt.Errorf("missing authorization header")
	}

	authHeader := authHeaders[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid authorization format")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// UserIDFromContext достаёт ID авторизованного пользователя из контекста
func UserIDFromContext(ctx context.Context) (int64, bool) {
	uid, ok := ctx.Value(UserIDKey).(int64)
	return uid, ok
}
