package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"wager/internal/domain"
	"wager/internal/domain/mocks"
)

func TestNew(t *testing.T) {
	assert.NotNil(t, New(&mocks.WagerRepository{}))
}

func TestClose(t *testing.T) {
	mockRepo := &mocks.WagerRepository{}
	mockRepo.On("Close", mock.Anything).Return(nil)

	app := New(mockRepo)
	assert.NotNil(t, app)
	require.NoError(t, app.Close(context.Background()))
}

func TestRun(t *testing.T) {
	mockRepo := &mocks.WagerRepository{}
	mockRepo.On("Close", mock.Anything).Return(nil)

	app := New(mockRepo)
	assert.NotNil(t, app)

	go app.Run(rand.Intn(100))

	require.NoError(t, app.Close(context.Background()))
}

func TestPlaceWager(t *testing.T) {
	tcs := []struct {
		name       string
		in         domain.Wager
		statusCode int
	}{
		{
			name: "normal",
			in: domain.Wager{
				TotalWagerValue:   10,
				Odds:              1,
				SellingPercentage: 10,
				SellingPrice:      decimal.NewFromFloat(10.11),
			},
			statusCode: 201,
		},
		{
			name: "negative totat wager value",
			in: domain.Wager{
				TotalWagerValue:   -1,
				Odds:              1,
				SellingPercentage: 10,
				SellingPrice:      decimal.NewFromFloat(10.11),
			},
			statusCode: 400,
		},
		{
			name: "bad selling price",
			in: domain.Wager{
				Odds:              1,
				SellingPrice:      decimal.NewFromFloat(10.11),
				TotalWagerValue:   100,
				SellingPercentage: 100,
			},
			statusCode: 400,
		},
	}

	mockRepo := &mocks.WagerRepository{}
	mockRepo.On("Close", mock.Anything).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(domain.Wager{}, nil)

	app := New(mockRepo)
	assert.NotNil(t, app)

	go app.Run(rand.Intn(100))
	defer app.Close(context.Background())

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			data, _ := json.Marshal(tc.in)
			req := httptest.NewRequest(http.MethodPost, "/wagers", bytes.NewBuffer(data))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			app.placeWager(ctx)
			assert.Equal(t, rec.Code, tc.statusCode)
		})
	}
}

func TestBuyWager(t *testing.T) {
	tcs := []struct {
		name       string
		in         domain.Wager
		statusCode int
	}{
		{
			name: "normal",
			in: domain.Wager{
				TotalWagerValue:   10,
				Odds:              1,
				SellingPercentage: 10,
				SellingPrice:      decimal.NewFromFloat(10.11),
			},
			statusCode: 201,
		},
		{
			name: "negative totat wager value",
			in: domain.Wager{
				TotalWagerValue:   -1,
				Odds:              1,
				SellingPercentage: 10,
				SellingPrice:      decimal.NewFromFloat(10.11),
			},
			statusCode: 400,
		},
		{
			name: "bad selling price",
			in: domain.Wager{
				Odds:              1,
				SellingPrice:      decimal.NewFromFloat(10.11),
				TotalWagerValue:   100,
				SellingPercentage: 100,
			},
			statusCode: 400,
		},
	}

	mockRepo := &mocks.WagerRepository{}
	mockRepo.On("Close", mock.Anything).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(domain.Wager{}, nil)

	app := New(mockRepo)
	assert.NotNil(t, app)

	go app.Run(rand.Intn(100))
	defer app.Close(context.Background())

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			data, _ := json.Marshal(tc.in)
			req := httptest.NewRequest(http.MethodPost, "/wagers", bytes.NewBuffer(data))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			app.placeWager(ctx)
			assert.Equal(t, rec.Code, tc.statusCode)
		})
	}
}

func TestGetWagers(t *testing.T) {
	tcs := []struct {
		name       string
		page       int
		limit      int
		statusCode int
		hasRes     bool
	}{
		{
			name:       "normal",
			page:       1,
			limit:      10,
			statusCode: 200,
		},
		{
			name:       "invalid limit",
			limit:      200,
			page:       1,
			statusCode: 400,
		},
	}

	mockRepo := &mocks.WagerRepository{}
	mockRepo.On("Close", mock.Anything).Return(nil)
	mockRepo.On("Get", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(make([]domain.Wager, 10), nil)

	app := New(mockRepo)
	assert.NotNil(t, app)

	go app.Run(rand.Intn(100))
	defer app.Close(context.Background())

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, "/wagers", nil)
			req.URL.Query().Add("page", fmt.Sprint(tc.page))
			req.URL.Query().Add("limit", fmt.Sprint(tc.limit))

			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			app.getWagers(ctx)
			assert.Equal(t, rec.Code, tc.statusCode)
		})
	}
}
