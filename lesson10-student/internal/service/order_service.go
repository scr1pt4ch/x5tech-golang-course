package service

import (
	"context"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	// TODO: validate input, calculate total, set initial status and history,
	// then save the order through the repository interface.
	return order.Order{}, order.ErrNotImplemented
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	// TODO: implement the scenario through the repository interface.
	return order.Order{}, order.ErrNotImplemented
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	// TODO: implement the scenario through the repository interface.
	return nil, order.ErrNotImplemented
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	// TODO: validate the requested transition, update current status and history,
	// then persist the changed order.
	return order.Order{}, order.ErrNotImplemented
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	// TODO: return status changes in chronological order.
	return nil, order.ErrNotImplemented
}
