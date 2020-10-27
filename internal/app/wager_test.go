package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		hasErr     bool
		err        ErrorResponse
	}{
		{
			name: "successful place",
			in: domain.Wager{
				TotalWagerValue:   10,
				Odds:              1,
				SellingPercentage: 10,
				SellingPrice:      decimal.NewFromFloat(10.11),
			},
			statusCode: 201,
		},
		{
			name: "negative total_wager_value",
			in: domain.Wager{
				TotalWagerValue:   -1,
				Odds:              1,
				SellingPercentage: 10,
				SellingPrice:      decimal.NewFromFloat(10.11),
			},
			statusCode: 400,
			hasErr:     true,
			err: ErrorResponse{
				Description: domain.ErrInvalidTotalWagerValue,
			},
		},
		{
			name: "invalid selling_price",
			in: domain.Wager{
				Odds:              1,
				SellingPrice:      decimal.NewFromFloat(10.11),
				TotalWagerValue:   100,
				SellingPercentage: 100,
			},
			statusCode: 400,
			hasErr:     true,
			err: ErrorResponse{
				Description: domain.ErrInvalidSellingPrice,
			},
		},
		{
			name: "invalid selling_price scale",
			in: domain.Wager{
				Odds:              1,
				SellingPrice:      decimal.NewFromFloat(10.111),
				TotalWagerValue:   100,
				SellingPercentage: 100,
			},
			statusCode: 400,
			hasErr:     true,
			err: ErrorResponse{
				Description: domain.ErrInvalidSellingPrice,
			},
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

			if tc.hasErr {
				var errRes ErrorResponse
				assert.Nil(t, json.Unmarshal(rec.Body.Bytes(), &errRes))
				assert.Equal(t, errRes, tc.err)
			}

			assert.Equal(t, rec.Code, tc.statusCode)
		})
	}
}

func TestBuyWager(t *testing.T) {
	tcs := []struct {
		name       string
		in         domain.Purchase
		statusCode int
		hasErr     bool
		err        ErrorResponse
	}{
		{
			name: "successful purchase",
			in: domain.Purchase{
				WagerID:     1,
				BuyingPrice: decimal.NewFromFloat(11.11),
			},
			statusCode: 201,
		},
		{
			name: "invalid wager_id",
			in: domain.Purchase{
				WagerID:     -1,
				BuyingPrice: decimal.NewFromFloat(11.11),
			},
			statusCode: 400,
			hasErr:     true,
			err: ErrorResponse{
				Description: domain.ErrInvalidWagerID,
			},
		},
		{
			name: "invalid buying_price",
			in: domain.Purchase{
				WagerID:     1,
				BuyingPrice: decimal.NewFromFloat(-1.00),
			},
			statusCode: 400,
			hasErr:     true,
			err: ErrorResponse{
				Description: domain.ErrInvalidBuyingPrice,
			},
		},
	}

	mockRepo := &mocks.WagerRepository{}
	mockRepo.On("Close", mock.Anything).Return(nil)
	mockRepo.On("Purchase", mock.Anything, mock.Anything, mock.Anything).Return(domain.Purchase{}, nil)

	app := New(mockRepo)
	assert.NotNil(t, app)

	go app.Run(rand.Intn(100))
	defer app.Close(context.Background())

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			data, _ := json.Marshal(tc.in)
			wagerID := tc.in.WagerID
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/buy/%d", wagerID), bytes.NewBuffer(data))

			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			app.buyWager(ctx)

			if tc.hasErr {
				var errRes ErrorResponse
				assert.Nil(t, json.Unmarshal(rec.Body.Bytes(), &errRes))
				assert.Equal(t, errRes, tc.err)
			}
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
		hasErr     bool
		err        ErrorResponse
	}{
		{
			name:       "get wagers sucessfully",
			page:       1,
			limit:      10,
			statusCode: 200,
		},
		{
			name:       "invalid limit",
			limit:      200,
			page:       1,
			statusCode: 400,
			hasErr:     true,
			err: ErrorResponse{
				Description: fmt.Sprintf("limit must be less than %d", maxWagerInPage),
			},
		},
	}

	mockRepo := &mocks.WagerRepository{}
	mockRepo.On("Close", mock.Anything).Return(nil)
	mockRepo.On("Get", mock.Anything, mock.AnythingOfType("int"), mock.AnythingOfType("int")).
		Return(make([]domain.Wager, 10, 10), 10, nil)

	app := New(mockRepo)
	assert.NotNil(t, app)

	go app.Run(rand.Intn(100))
	defer app.Close(context.Background())

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodGet, "/wagers", nil)
			values := url.Values{
				"page":  []string{fmt.Sprint(tc.page)},
				"limit": []string{fmt.Sprint(tc.limit)},
			}
			req.URL.RawQuery += values.Encode()

			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)

			app.getWagers(ctx)
			assert.Equal(t, rec.Code, tc.statusCode)

			if tc.hasErr {
				var errRes ErrorResponse
				assert.Nil(t, json.Unmarshal(rec.Body.Bytes(), &errRes))
				assert.Equal(t, errRes, tc.err)
			}
		})
	}
}
