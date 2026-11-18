package repository

import (
	"errors"
	"sort"
	"sync"

	"example.com/lesson9practice/internal/model"
)

var ErrNotFound = errors.New("order not found")

type Memory struct {
	mu     sync.Mutex
	nextID int
	orders map[int]model.Order
}

func New() *Memory { return &Memory{nextID: 1, orders: make(map[int]model.Order)} }

func (m *Memory) Create(customer string) (model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order := model.Order{ID: m.nextID, Customer: customer, Status: "new"}
	m.orders[order.ID] = order
	m.nextID++
	return order, nil
}
func (m *Memory) ByID(id int) (model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	order, ok := m.orders[id]
	if !ok {
		return model.Order{}, ErrNotFound
	}
	return order, nil
}

func (m *Memory) List(limit, offset int) ([]model.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	all := make([]model.Order, 0, len(m.orders))
	for _, order := range m.orders {
		all = append(all, order)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })

	if len(all) <= offset {
		return []model.Order{}, nil
	}

	page := all[offset:]
	if limit < len(page) {
		page = page[:limit]
	}
	return page, nil
}
