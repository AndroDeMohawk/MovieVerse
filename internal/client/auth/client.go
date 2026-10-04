package auth

import (
	"context"
	"fmt"

	ssov1 "github.com/AndroDeMohawk/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"log/slog"
)

type Client struct {
	api    ssov1.AuthClient
	appID  int32
	logger *slog.Logger
}

func New(ctx context.Context, addr string, appID int32, logger *slog.Logger) (*Client, error) {
	const op = "client.auth.New"
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Client{
		api:    ssov1.NewAuthClient(conn),
		appID:  appID,
		logger: logger,
	}, nil
}

func (c *Client) Register(ctx context.Context, email, password string) (int64, error) {
	const op = "client.auth.Register"
	resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return resp.GetUserId(), nil
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	const op = "client.auth.Login"
	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{
		Email:    email,
		Password: password,
		AppId:    c.appID,
	})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resp.GetToken(), nil
}

func (c *Client) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "client.auth.IsAdmin"
	resp, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: userID,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.GetIsAdmin(), nil
}
