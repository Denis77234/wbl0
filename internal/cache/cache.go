package cache

import (
	"context"
	"errors"

	"awesomeProject/internal/model"
)

type orderStorage interface {
	SelectOrders(ctx context.Context) ([]model.Order, error)
}

type Cache struct {
	storage map[string]model.Order
}

func New(strg orderStorage) (*Cache, error) {
	mp := make(map[string]model.Order)

	orders, err := strg.SelectOrders(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, order := range orders {
		mp[order.OrderUid] = order
	}

	cache := &Cache{storage: mp}

	return cache, nil
}

func (c *Cache) GetOrder(id string) (model.Order, error) {
	order, ok := c.storage[id]
	if !ok {
		return model.Order{}, errors.New("unknown id")
	}

	return order, nil
}

func (c *Cache) PutOrder(order model.Order) error {
	_, ok := c.storage[order.OrderUid]
	if ok {
		return errors.New("order already exists")
	}

	c.storage[order.OrderUid] = order

	return nil
}
