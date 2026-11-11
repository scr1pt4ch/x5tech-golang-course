package memory

import (
	"context"
	"fmt"

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
	toSave := cloneOrder(value)
	toSave.ID = r.nextID
	r.nextID++

	r.items = append(r.items, toSave)

	return cloneOrder(toSave), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	for _, item := range r.items {
		if id == item.ID {
			return cloneOrder(item), nil
		}
	}

	return order.Order{}, fmt.Errorf("order %d: %w", id, order.ErrOrderNotFound)
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	res := make([]order.Order, 0, len(r.items))

	for _, item := range r.items {
		res = append(res, cloneOrder(item))
	}

	return res, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	for i := range r.items {
		if value.ID == r.items[i].ID {
			toSave := cloneOrder(value)
			r.items[i] = toSave

			return cloneOrder(toSave), nil
		}
	}

	return order.Order{}, fmt.Errorf("order %d: %w", value.ID, order.ErrOrderNotFound)
}

// Нужен, чтобы потом было проще добавить работу с бд, да и в общем для идиоматичности
func cloneOrder(o order.Order) order.Order {
	cloned := o

	if nil != o.Items {
		cloned.Items = make([]order.Item, len(o.Items))
		copy(cloned.Items, o.Items)
	}

	if nil != o.History {
		cloned.History = make([]order.StatusChange, len(o.History))
		copy(cloned.History, o.History)
	}

	return cloned
}
