package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/repository/memory"
)

func TestOrderRepositoryCreateAssignsSequentialIDs(t *testing.T) {
	tests := []struct {
		name    string
		inputs  []order.Order
		wantIDs []int64
	}{
		{"один заказ, свой ID игнорируется", []order.Order{{ID: 100}}, []int64{1}},
		{"три заказа подряд", []order.Order{{}, {}, {}}, []int64{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewOrderRepository()

			for i, input := range tt.inputs {
				created, err := repo.Create(context.Background(), input)

				require.NoError(t, err)
				assert.Equal(t, tt.wantIDs[i], created.ID)
			}
		})
	}
}

func TestOrderRepositoryNotFound(t *testing.T) {
	tests := []struct {
		name string
		call func(repo *memory.OrderRepository) error
	}{
		{"GetByID", func(repo *memory.OrderRepository) error {
			_, err := repo.GetByID(context.Background(), 42)
			return err
		}},
		{"Update", func(repo *memory.OrderRepository) error {
			_, err := repo.Update(context.Background(), order.Order{ID: 42})
			return err
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(memory.NewOrderRepository())

			assert.ErrorIs(t, err, order.ErrOrderNotFound)
		})
	}
}

func TestOrderRepositoryList(t *testing.T) {
	tests := []struct {
		name    string
		create  int
		wantLen int
	}{
		{"пустой репозиторий", 0, 0},
		{"один заказ", 1, 1},
		{"три заказа", 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewOrderRepository()
			ctx := context.Background()

			for i := 0; i < tt.create; i++ {
				_, err := repo.Create(ctx, order.Order{})
				require.NoError(t, err)
			}

			list, err := repo.List(ctx)

			require.NoError(t, err)
			assert.NotNil(t, list)
			assert.Len(t, list, tt.wantLen)
		})
	}
}

func TestOrderRepositoryUpdate(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	created, err := repo.Create(ctx, order.Order{Status: order.StatusNew})
	require.NoError(t, err)

	created.Status = order.StatusProcessing
	_, err = repo.Update(ctx, created)
	require.NoError(t, err)

	stored, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, order.StatusProcessing, stored.Status)
}

func TestOrderRepositoryCopies(t *testing.T) {
	tests := []struct {
		name   string
		change func(input, fetched order.Order)
	}{
		{"меняем входной заказ после Create", func(input, _ order.Order) {
			input.Items[0].Price = 1
			input.History[0].Status = order.StatusCancelled
		}},
		{"меняем заказ, полученный из GetByID", func(_, fetched order.Order) {
			fetched.Items[0].Price = 1
			fetched.History[0].Status = order.StatusCancelled
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := memory.NewOrderRepository()
			ctx := context.Background()

			input := order.Order{
				Items:   []order.Item{{Name: "Keyboard", Quantity: 1, Price: 500000}},
				History: []order.StatusChange{{Status: order.StatusNew, ChangedAt: time.Now()}},
			}

			created, err := repo.Create(ctx, input)
			require.NoError(t, err)

			fetched, err := repo.GetByID(ctx, created.ID)
			require.NoError(t, err)

			tt.change(input, fetched)

			again, err := repo.GetByID(ctx, created.ID)
			require.NoError(t, err)
			assert.Equal(t, int64(500000), again.Items[0].Price)
			assert.Equal(t, order.StatusNew, again.History[0].Status)
		})
	}
}
