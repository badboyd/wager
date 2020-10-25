package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq" // postgresql implementation package in go
	"github.com/shopspring/decimal"

	"wager/config"
	"wager/internal/domain"
	repository "wager/internal/repository/postgres"
)

type (
	// App application struct
	App struct {
		e    *echo.Echo
		cfg  *config.Schema
		repo *repository.PostgresRepository
	}
)

// New application
func New() *App {
	cfg, err := config.Load()
	if err != nil {
		log.Panicf("Cannot load configuration: %s\n", err.Error())
	}

	log.Printf("%+v", cfg)

	dbConfig := fmt.Sprintf("user=%s dbname=%s host=%s port=%d sslmode=disable",
		cfg.Database.Username, cfg.Database.Database, cfg.Database.Host, cfg.Database.Port)
	log.Printf("Init db with these param %v", dbConfig)

	app := &App{
		cfg:  cfg,
		e:    echo.New(),
		repo: repository.New(sqlx.MustConnect("postgres", dbConfig)),
	}

	// handle recover
	app.e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))

	// health and live check
	app.e.GET("/health", app.healthCheck)
	app.e.GET("/live", app.liveCheck)

	// init app routing
	app.e.GET("/wagers", app.getWagers)
	app.e.POST("/wagers", app.placeWager)
	app.e.POST("/buy/:wager_id", app.buyWager)

	return app
}

// Run application
func (app *App) Run() error {
	go func() {
		if err := app.e.Start(fmt.Sprintf(":%d", app.cfg.Service.Port)); err != nil {
			// we shoud panic here, but I prefer a gracfully stop
			// there will be live/health check in production environment
			// if 1 of them is failed then the instance will be killed
			log.Println("Shutting down the server")
		}
	}()

	// wait for the signal
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, os.Kill)

	log.Printf("Received signal %s", <-ch)
	defer cancel()

	return app.e.Shutdown(ctx)
}

// this is for health check procedure
func (app *App) healthCheck(e echo.Context) error {
	return e.JSON(http.StatusOK, "OK")
}

func (app *App) liveCheck(e echo.Context) error {
	return e.JSON(http.StatusOK, "OK")
}

// ErrorResponse ...
type ErrorResponse struct {
	Description string `json:"error"`
}

func (e *ErrorResponse) Error() string {
	return e.Description
}

func (app *App) placeWager(ctx echo.Context) error {
	wager := domain.Wager{}
	if err := ctx.Bind(&wager); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	if err := app.repo.Create(ctx.Request().Context(), &wager); err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Description: err.Error()})
	}

	return ctx.JSON(http.StatusCreated, wager)
}

// GetWagersRequest ...
type GetWagersRequest struct {
	Page  int `json:"page" query:"page"` // This one should be the wager id
	Limit int `json:"limit" query:"limit"`
}

func (app *App) getWagers(ctx echo.Context) error {
	req := GetWagersRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	if req.Limit == 0 || req.Limit > 20 {
		req.Limit = 10
	}
	wagers, _, err := app.repo.Get(ctx.Request().Context(), req.Page, req.Limit)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Description: err.Error()})
	}

	return ctx.JSON(http.StatusOK, wagers)
}

// PurchaseRequest ...
type PurchaseRequest struct {
	WagerID     int             `json:"wager_id" param:"wager_id" validate:"required"`
	BuyingPrice decimal.Decimal `json:"buying_price" validate:"gt=0"`
}

func (app *App) buyWager(ctx echo.Context) error {
	return nil
}
