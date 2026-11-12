package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
)

// Чтобы тестировать именно сервис
type dummyRepository struct {
	orders []order.Order
	nextID int64
}

func newDummyRepository() *dummyRepository {
	return &dummyRepository{nextID: 1}
}

func (d *dummyRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	value.ID = d.nextID
	d.nextID++
	d.orders = append(d.orders, value)
	return value, nil
}

func (d *dummyRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, o := range d.orders {
		if id == o.ID {
			return o, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (d *dummyRepository) List(_ context.Context) ([]order.Order, error) {
	return d.orders, nil
}

func (d *dummyRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i := range d.orders {
		if value.ID == d.orders[i].ID {
			d.orders[i] = value
			return value, nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (d *dummyRepository) addOrder(status order.Status) int64 {
	id := d.nextID
	d.nextID++
	d.orders = append(d.orders, order.Order{
		ID:      id,
		Status:  status,
		History: []order.StatusChange{{Status: status, ChangedAt: time.Now().UTC()}},
	})
	return id
}

func TestCreate(t *testing.T) {
	svc := service.NewOrderService(newDummyRepository())

	created, err := svc.Create(context.Background(), order.CreateOrderInput{
		Items: []order.Item{
			{Name: "     Keyboard", Quantity: 1, Price: 500000},
			{Name: "Mouse", Quantity: 2, Price: 150000},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), created.ID)
	assert.Equal(t, "Keyboard", created.Items[0].Name)
	assert.Equal(t, order.StatusNew, created.Status)
	assert.Equal(t, int64(800000), created.TotalPrice)
	require.Len(t, created.History, 1)
	assert.Equal(t, order.StatusNew, created.History[0].Status)
}

func TestCreate_Validation(t *testing.T) {
	tests := []struct {
		name  string
		items []order.Item
	}{
		{"нет позиций", nil},
		{"пустое название", []order.Item{{Name: " ", Quantity: 1, Price: 100}}},
		{"нулевое количество", []order.Item{{Name: "Mouse", Quantity: 0, Price: 100}}},
		{"отрицательное количество", []order.Item{{Name: "Mouse", Quantity: -1, Price: 100}}},
		{"нулевая цена", []order.Item{{Name: "Mouse", Quantity: 1, Price: 0}}},
		{"отрицательная цена", []order.Item{{Name: "Mouse", Quantity: 1, Price: -100}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := service.NewOrderService(newDummyRepository())

			_, err := svc.Create(context.Background(), order.CreateOrderInput{Items: tt.items})

			assert.ErrorIs(t, err, order.ErrInvalidOrder)
		})
	}
}

func TestGetByIDAndList(t *testing.T) {
	repo := newDummyRepository()
	id := repo.addOrder(order.StatusNew)
	svc := service.NewOrderService(repo)

	found, err := svc.GetByID(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, id, found.ID)

	_, err = svc.GetByID(context.Background(), 999)
	assert.ErrorIs(t, err, order.ErrOrderNotFound)

	list, err := svc.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name    string
		from    order.Status
		to      order.Status
		wantErr error
	}{
		{"from new to processing", order.StatusNew, order.StatusProcessing, nil},
		{"from new to cancelled", order.StatusNew, order.StatusCancelled, nil},
		{"from processing to completed", order.StatusProcessing, order.StatusCompleted, nil},
		{"from processing to cancelled", order.StatusProcessing, order.StatusCancelled, nil},
		{"from new to completed", order.StatusNew, order.StatusCompleted, order.ErrInvalidStatusTransition},
		{"from processing to new", order.StatusProcessing, order.StatusNew, order.ErrInvalidStatusTransition},
		{"from completed to cancelled", order.StatusCompleted, order.StatusCancelled, order.ErrInvalidStatusTransition},
		{"from cancelled to processing", order.StatusCancelled, order.StatusProcessing, order.ErrInvalidStatusTransition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newDummyRepository()
			id := repo.addOrder(tt.from)
			svc := service.NewOrderService(repo)

			updated, err := svc.UpdateStatus(context.Background(), id, tt.to)

			if nil != tt.wantErr {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.to, updated.Status)
		})
	}
}

func TestHistory(t *testing.T) {
	svc := service.NewOrderService(newDummyRepository())
	ctx := context.Background()

	created, err := svc.Create(ctx, order.CreateOrderInput{
		Items: []order.Item{{Name: "Keyboard", Quantity: 1, Price: 500000}},
	})
	require.NoError(t, err)

	_, err = svc.UpdateStatus(ctx, created.ID, order.StatusProcessing)
	require.NoError(t, err)
	_, err = svc.UpdateStatus(ctx, created.ID, order.StatusCompleted)
	require.NoError(t, err)

	history, err := svc.History(ctx, created.ID)
	require.NoError(t, err)

	require.Len(t, history, 3)
	assert.Equal(t, order.StatusNew, history[0].Status)
	assert.Equal(t, order.StatusProcessing, history[1].Status)
	assert.Equal(t, order.StatusCompleted, history[2].Status)
	assert.False(t, history[2].ChangedAt.Before(history[1].ChangedAt))
	assert.False(t, history[1].ChangedAt.Before(history[0].ChangedAt))
}

func TestUpdateStatus_UnknownStatus(t *testing.T) {
	repo := newDummyRepository()
	id := repo.addOrder(order.StatusNew)
	svc := service.NewOrderService(repo)

	_, err := svc.UpdateStatus(context.Background(), id, "shipped")

	assert.ErrorIs(t, err, order.ErrInvalidStatus)
}

func TestHistory_NotFound(t *testing.T) {
	svc := service.NewOrderService(newDummyRepository())

	_, err := svc.History(context.Background(), 999)

	assert.ErrorIs(t, err, order.ErrOrderNotFound)
}
