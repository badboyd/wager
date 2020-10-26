package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq" // postgresql implementation package in go
	"github.com/shopspring/decimal"
	"gopkg.in/go-playground/validator.v9"

	"wager/internal/domain"
)

type (
	// App application struct
	App struct {
		e         *echo.Echo
		repo      domain.WagerRepository
		validator *validator.Validate
	}
)

// New application
func New(repo domain.WagerRepository) *App {
	app := &App{
		e:         echo.New(),
		repo:      repo,
		validator: validator.New(),
	}

	// create customr validator for numeric type
	app.initNumericValidator()

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

// Init validator for buying_price
func (app *App) initNumericValidator() {
	app.validator.RegisterCustomTypeFunc(func(f reflect.Value) interface{} {
		fieldDecimal, ok := f.Interface().(decimal.Decimal)
		if ok {
			val, _ := fieldDecimal.Float64()
			return val
		}

		return nil
	}, decimal.Decimal{})

	app.validator.RegisterValidationCtx("v_selling_price", func(ctx context.Context, fi validator.FieldLevel) bool {
		sellingPrice := decimal.NewFromFloat(fi.Field().Float())

		totalWagerValue := fi.Parent().FieldByName("TotalWagerValue").Int()
		sellingPercentage := fi.Parent().FieldByName("SellingPercentage").Int()

		return sellingPrice.GreaterThan(decimal.NewFromInt(totalWagerValue * sellingPercentage / 100))
	})
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
	Description string `json:"description"`
}

func (e *ErrorResponse) Error() string {
	return e.Description
}

// all the handlers will have the same pattern
// First bind the request
// Second validate it
// Third call repository to persist the data

func (app *App) placeWager(ctx echo.Context) error {
	wager := domain.Wager{}
	if err := ctx.Bind(&wager); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	if err := app.validator.StructCtx(ctx.Request().Context(), wager); err != nil {
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
	Page  int `json:"page" query:"page" validate:"min=1"`          // This one should be the wager id
	Limit int `json:"limit" query:"limit" validate:"min=1,max=20"` // There should be a maximum value for limit
}

func (app *App) getWagers(ctx echo.Context) error {
	req := getWagersRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	if err := app.validator.StructCtx(ctx.Request().Context(), req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	wagers, _, err := app.repo.Get(ctx.Request().Context(), req.Page, req.Limit)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Description: err.Error()})
	}

	return ctx.JSON(http.StatusOK, wagers)
}

// PurchaseRequest ...
type purchaseRequest struct {
	WagerID     int             `json:"wager_id" param:"wager_id" validate:"required"`
	BuyingPrice decimal.Decimal `json:"buying_price" validate:"gt=0"`
}

func (app *App) buyWager(ctx echo.Context) error {
	req := purchaseRequest{}
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	if err := app.validator.StructCtx(ctx.Request().Context(), req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Description: err.Error()})
	}

	res, err := app.repo.Purchase(ctx.Request().Context(), req.WagerID, req.BuyingPrice)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Description: err.Error()})
	}

	return ctx.JSON(http.StatusOK, res)
}
