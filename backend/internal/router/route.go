package router

import (
	"github.com/labstack/echo/v5"
	"github.com/o-ga09/go-backend-template/internal/database/mysql"
	"github.com/o-ga09/go-backend-template/internal/handler"
	"github.com/o-ga09/go-backend-template/pkg/config"
	"github.com/o-ga09/go-backend-template/pkg/session"
)

type route struct {
	rooAPI  *echo.Group
	user    handler.IUser
	auth    handler.IAuth
	product handler.IProduct
}

type IRouting interface {
	SetupApplicationRoute()
	SetupSystemRoute()
}

func New(root *echo.Group, cfg config.Config) IRouting {
	userRepo := mysql.NewUserRepository()
	productRepo := mysql.NewProductRepository()
	sessionMgr := session.NewManager(cfg.SessionSecret, session.SessionTTL)
	return &route{
		rooAPI:  root,
		user:    handler.NewUserHandler(userRepo, sessionMgr),
		auth:    handler.NewAuthHandler(userRepo),
		product: handler.NewProductHandler(productRepo),
	}
}
