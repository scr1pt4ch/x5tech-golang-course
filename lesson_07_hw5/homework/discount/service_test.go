package discount

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent int
		wantErr error
	}{
		{name: "Ноль допустим", percent: 0, wantErr: nil},
		{name: "Обычный процент", percent: 10, wantErr: nil},
		{name: "Сто допустимо", percent: 100, wantErr: nil},
		{name: "Минус один ошибка", percent: -1, wantErr: ErrInvalidPercent},
		{name: "Сто один ошибка", percent: 101, wantErr: ErrInvalidPercent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidatePercent(tt.percent)

			if nil != tt.wantErr {
				assert.ErrorIs(t, err, tt.wantErr)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestApply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		price   int
		percent int
		want    int
		wantErr error
	}{
		{name: "Обычная скидка", price: 1000, percent: 20, want: 800},
		{name: "Скидка 0%", price: 1000, percent: 0, want: 1000},
		{name: "Скидка 100%", price: 1000, percent: 100, want: 0},
		{name: "Нулевая цена", price: 0, percent: 30, want: 0},
		{name: "Дробная часть отбрасывается", price: 99, percent: 10, want: 89},
		{name: "Отрицательная цена", price: -1, percent: 0, wantErr: ErrNegativePrice},
		{name: "Процент меньше 0", price: 1000, percent: -1, wantErr: ErrInvalidPercent},
		{name: "Процент больше 100", price: 1000, percent: 101, wantErr: ErrInvalidPercent},
		{name: "Сначала проверяется цена", price: -1, percent: 101, wantErr: ErrNegativePrice},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := Apply(tt.price, tt.percent)

			if nil != tt.wantErr {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFinalPrice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		items   []Item
		want    int
		wantErr error
	}{
		{name: "nil вместо списка", items: nil, want: 0},
		{name: "Пустой список", items: []Item{}, want: 0},
		{
			name:  "Один товар",
			items: []Item{{Price: 200, DiscountPercent: 50}},
			want:  100,
		},
		{
			name: "Несколько корректных товаров",
			items: []Item{
				{Price: 1000, DiscountPercent: 10},
				{Price: 100, DiscountPercent: 0},
				{Price: 50, DiscountPercent: 100},
				{Price: 19, DiscountPercent: 50},
			},
			want: 1009,
		},
		{
			name: "Товар с некорректной ценой",
			items: []Item{
				{Price: 100, DiscountPercent: 10},
				{Price: -100, DiscountPercent: 10},
			},
			wantErr: ErrNegativePrice,
		},
		{
			name: "Товар с некорректной скидкой",
			items: []Item{
				{Price: 100, DiscountPercent: 10},
				{Price: 100, DiscountPercent: 101},
			},
			wantErr: ErrInvalidPercent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := FinalPrice(tt.items)

			if nil != tt.wantErr {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Zero(t, got)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
