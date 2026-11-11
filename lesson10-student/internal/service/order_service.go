package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if 0 == len(input.Items) {
		return order.Order{}, fmt.Errorf("%w: order must have at least one item", order.ErrInvalidOrder)
	}

	var total int64

	for i, item := range input.Items {
		if "" == strings.TrimSpace(item.Name) {
			return order.Order{}, fmt.Errorf("%w: item %d: empty name", order.ErrInvalidOrder, i)
		}

		if item.Quantity <= 0 {
			return order.Order{}, fmt.Errorf("%w: item %d: quantity <= 0", order.ErrInvalidOrder, i)
		}

		if item.Price <= 0 {
			return order.Order{}, fmt.Errorf("%w, item %d: price <= 0", order.ErrInvalidOrder, i)
		}

		total += item.Price * int64(item.Quantity)
	}

	now := time.Now()

	newOrder := order.Order{
		Items:      input.Items,
		TotalPrice: total,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{Status: order.StatusNew, ChangedAt: now},
		},
	}

	created, err := s.repository.Create(ctx, newOrder)

	if nil != err {
		return order.Order{}, fmt.Errorf("create order: %w", err)
	}

	return created, nil
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	if !isKnownStatus(status) {
		return order.Order{}, fmt.Errorf("%w: %q", order.ErrInvalidStatus, status)
	}

	curr, err := s.repository.GetByID(ctx, id)

	if nil != err {
		return order.Order{}, err
	}

	if !canTransit(curr.Status, status) {
		return order.Order{}, fmt.Errorf("%w: from %s to %s", order.ErrInvalidStatusTransition, curr.Status, status)
	}

	curr.Status = status

	curr.History = append(curr.History, order.StatusChange{
		Status:    status,
		ChangedAt: time.Now(),
	})

	updated, err := s.repository.Update(ctx, curr)

	if nil != err {
		return order.Order{}, fmt.Errorf("update order status: %w", err)
	}

	return updated, nil
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	found, err := s.repository.GetByID(ctx, id)

	if nil != err {
		return nil, err
	}

	return found.History, nil
}

var allowedTransitions = map[order.Status][]order.Status{
	order.StatusNew:        {order.StatusProcessing, order.StatusCancelled},
	order.StatusProcessing: {order.StatusCompleted, order.StatusCancelled},
}

func canTransit(from, to order.Status) bool {
	for _, allowed := range allowedTransitions[from] {
		if to == allowed {
			return true
		}
	}

	return false
}

// Пока что захардкожено, потом сделаю красивее
func isKnownStatus(status order.Status) bool {
	switch status {
	case order.StatusNew, order.StatusProcessing, order.StatusCompleted, order.StatusCancelled:
		return true
	default:
		return false
	}
}
