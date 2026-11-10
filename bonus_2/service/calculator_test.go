package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/delivery/model"
	"example.com/delivery/repository"
	"example.com/delivery/service"
)

func TestCalculatorAllTariffs(t *testing.T) {
	t.Parallel()

	calculator := service.NewCalculator(repository.Default())

	tests := []struct {
		tariff string
		want   int64
	}{
		{"courier", 50_000},
		{"pickup", 21_000},
		{"express", 110_000},
	}

	for _, tt := range tests {
		t.Run(tt.tariff, func(t *testing.T) {
			t.Parallel()

			order := model.Order{ID: "order-1", Tariff: tt.tariff, WeightGrams: 2_500, DistanceKM: 10}
			got, err := calculator.Calculate(order)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculatorErrors(t *testing.T) {
	t.Parallel()

	calculator := service.NewCalculator(repository.Default())

	tests := []struct {
		name    string
		order   model.Order
		wantErr error
	}{
		{
			name:    "неизвестный тариф",
			order:   model.Order{ID: "order-2", Tariff: "drone", WeightGrams: 1_000, DistanceKM: 5},
			wantErr: service.ErrUnknownTariff,
		},
		{
			name:    "неверный вес",
			order:   model.Order{ID: "order-3", Tariff: "courier", WeightGrams: 0, DistanceKM: 5},
			wantErr: service.ErrInvalidWeight,
		},
		{
			name:    "слишком далеко",
			order:   model.Order{ID: "order-4", Tariff: "pickup", WeightGrams: 1_000, DistanceKM: 51},
			wantErr: service.ErrDistanceTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := calculator.Calculate(tt.order)

			assert.Zero(t, got)
			require.ErrorIs(t, err, tt.wantErr)

			var priceErr *service.OrderPriceError
			require.ErrorAs(t, err, &priceErr)
			assert.Equal(t, tt.order.ID, priceErr.OrderID)
			assert.Equal(t, tt.order.Tariff, priceErr.Tariff)

			assert.Contains(t, err.Error(), tt.order.ID)
			assert.Contains(t, err.Error(), tt.order.Tariff)
			assert.Contains(t, err.Error(), tt.wantErr.Error())
		})
	}
}

func TestOrderPriceErrorUnwrap(t *testing.T) {
	t.Parallel()

	err := &service.OrderPriceError{OrderID: "order-5", Tariff: "courier", Err: service.ErrInvalidWeight}

	assert.Equal(t, service.ErrInvalidWeight, err.Unwrap())
	assert.Equal(t, "заказ order-5, тариф courier: некорректный вес", err.Error())
}
