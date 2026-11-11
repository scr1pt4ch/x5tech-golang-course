package memory

import (
	"context"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	// TODO: assign an ID and store the order.
	return order.Order{}, order.ErrNotImplemented
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	// TODO: return ErrOrderNotFound when the order does not exist.
	return order.Order{}, order.ErrNotImplemented
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	// TODO: return all stored orders in creation order without exposing mutable internal state.
	return nil, order.ErrNotImplemented
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	// TODO: replace an existing order or return ErrOrderNotFound.
	return order.Order{}, order.ErrNotImplemented
}
