package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/delivery/service"
)

func TestTariffPrices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pricer   service.DeliveryPricer
		weight   int
		distance int
		want     int64
	}{
		{"курьер, обычный заказ", service.NewCourier(), 2_500, 10, 50_000},
		{"курьер, 1 грамм = 1 кг", service.NewCourier(), 1, 1, 26_500},
		{"курьер, ровно 1 кг", service.NewCourier(), 1_000, 1, 26_500},
		{"курьер, 1001 г = 2 кг", service.NewCourier(), 1_001, 1, 31_500},
		{"курьер, максимальная дистанция", service.NewCourier(), 1_000, 100, 175_000},

		{"самовывоз, обычный заказ", service.NewPickup(), 2_500, 10, 21_000},
		{"самовывоз, максимальная дистанция", service.NewPickup(), 1_000, 50, 37_000},

		{"экспресс, обычный заказ", service.NewExpress(), 2_500, 10, 110_000},
		{"экспресс, максимальная дистанция", service.NewExpress(), 1_000, 300, 960_000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.pricer.Price(tt.weight, tt.distance)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTariffValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pricer   service.DeliveryPricer
		weight   int
		distance int
		wantErr  error
	}{
		{"нулевой вес", service.NewCourier(), 0, 10, service.ErrInvalidWeight},
		{"отрицательный вес", service.NewPickup(), -1, 10, service.ErrInvalidWeight},
		{"нулевая дистанция", service.NewCourier(), 1_000, 0, service.ErrInvalidDistance},
		{"отрицательная дистанция", service.NewExpress(), 1_000, -5, service.ErrInvalidDistance},
		{"курьер дальше 100 км", service.NewCourier(), 1_000, 101, service.ErrDistanceTooLarge},
		{"самовывоз дальше 50 км", service.NewPickup(), 1_000, 51, service.ErrDistanceTooLarge},
		{"экспресс дальше 300 км", service.NewExpress(), 1_000, 301, service.ErrDistanceTooLarge},
		{"сначала проверяется вес", service.NewCourier(), 0, 0, service.ErrInvalidWeight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.pricer.Price(tt.weight, tt.distance)

			assert.ErrorIs(t, err, tt.wantErr)
			assert.Zero(t, got)
		})
	}
}
