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

func TestOrderRepository_CreateAssignsSequentialIDs(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	first, err := repo.Create(ctx, order.Order{ID: 100})
	require.NoError(t, err)
	second, err := repo.Create(ctx, order.Order{})
	require.NoError(t, err)

	assert.Equal(t, int64(1), first.ID)
	assert.Equal(t, int64(2), second.ID)
}

func TestOrderRepository_NotFound(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 42)
	assert.ErrorIs(t, err, order.ErrOrderNotFound)

	_, err = repo.Update(ctx, order.Order{ID: 42})
	assert.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestOrderRepository_Update(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	created, err := repo.Create(ctx, order.Order{Status: order.StatusNew})
	require.NoError(t, err)

	created.Status = order.StatusProcessing
	_, err = repo.Update(ctx, created)
	require.NoError(t, err)

	list, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, order.StatusProcessing, list[0].Status)
}

func TestOrderRepository_EmptyList(t *testing.T) {
	list, err := memory.NewOrderRepository().List(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, list)
	assert.Empty(t, list)
}

func TestOrderRepository_CreateCopiesInput(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	items := []order.Item{{Name: "Keyboard", Quantity: 1, Price: 500000}}
	created, err := repo.Create(ctx, order.Order{Items: items})
	require.NoError(t, err)

	items[0].Price = 1

	stored, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(500000), stored.Items[0].Price)
}

func TestOrderRepository_GetByIDReturnsCopy(t *testing.T) {
	repo := memory.NewOrderRepository()
	ctx := context.Background()

	created, err := repo.Create(ctx, order.Order{
		Items:   []order.Item{{Name: "Keyboard", Quantity: 1, Price: 500000}},
		History: []order.StatusChange{{Status: order.StatusNew, ChangedAt: time.Now()}},
	})
	require.NoError(t, err)

	fetched, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	fetched.Items[0].Price = 100
	fetched.History[0].Status = order.StatusCancelled

	again, err := repo.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(500000), again.Items[0].Price)
	assert.Equal(t, order.StatusNew, again.History[0].Status)
}
