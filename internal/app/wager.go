package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq" // postgresql implementation package in go

	"wager/internal/domain"
)

const (
	maxWagerInPage = 20
)

type (
	// App application struct
	App struct {
		e    *echo.Echo
		repo domain.WagerRepository
	}
)

// New application
func New(repo domain.WagerRepository) *App {
	app := &App{
		e:    echo.New(),
		repo: repo,
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
func (app *App) Run(port int) error {
	log.Printf("Start the server at :%d", port)
	return app.e.Start(fmt.Sprintf(":%d", port))
}

// Close app and all the resources
func (app *App) Close(ctx context.Context) error {
	log.Println("Close the app")
	if err := app.e.Shutdown(ctx); err != nil {
		// we should panic here, but we have a db dependency, that's why I try to log it out
		log.Printf("Shutdown http app error: %s\n", err.Error())
	}

	// close db connection
	return app.repo.Close(ctx)
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

// all the handlers will have the same pattern
// First bind the request
// Second validate it
// Third call repository to persist the data

func (app *App) placeWager(ctx echo.Context) error {
	log.Printf("Process a place wager request")

	wager := domain.Wager{}
	if err := ctx.Bind(&wager); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	if err := wager.Validate(ctx.Request().Context()); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	res, err := app.repo.Create(ctx.Request().Context(), wager)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Description: err.Error()})
	}

	return ctx.JSON(http.StatusCreated, res)
}

// GetWagersRequest ...
type getWagersRequest struct {
	Page  int `json:"page" query:"page"`   // This one should be the wager id
	Limit int `json:"limit" query:"limit"` // There should be a maximum value for limit
}

func (app *App) getWagers(ctx echo.Context) error {
	log.Printf("Process a get wagers request")

	req := getWagersRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}
	log.Printf("Start get wagers from %d limit %d", req.Page, req.Limit)

	// In my opinion, we should limit the number of returned wagers
	//if the limit is less than or equal zero, I change it to max number returned wagers
	// but the requirement does not say it so I assume I have to reject the large limt
	// in this case I set max limit to 20
	if req.Limit <= 0 || req.Limit > maxWagerInPage {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Description: fmt.Sprintf("limit must be less than %d", maxWagerInPage),
		})
	}

	if req.Page <= 0 {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Description: "Page can not be less than or equal to zero",
		})
	}

	wagers, _, err := app.repo.Get(ctx.Request().Context(), req.Page, req.Limit)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Description: err.Error()})
	}

	return ctx.JSON(http.StatusOK, wagers)
}

func (app *App) buyWager(ctx echo.Context) error {
	log.Printf("Process buy wager request")

	purchase := domain.Purchase{}
	if err := ctx.Bind(&purchase); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	if err := purchase.Validate(ctx.Request().Context()); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	res, err := app.repo.Purchase(ctx.Request().Context(), purchase.WagerID, purchase.BuyingPrice)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Description: err.Error()})
	}

	return ctx.JSON(http.StatusCreated, res)
}
