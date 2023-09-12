package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"

	"awesomeProject/internal/model"
)

type Cache interface {
	PutOrder(order model.Order) error
}

type OrderStorage interface {
	CreateOrder(ctx context.Context, ord model.Order) error
}

type OrderManager struct {
	cache   Cache
	storage OrderStorage

	Broker stan.Conn
}

func NewOrderManager(ch Cache, strg OrderStorage, natsUrl, clusterId, clientId string) (*OrderManager, error) {
	conn, err := nats.Connect(natsUrl)
	if err != nil {
		return nil, err
	}

	sc, err := stan.Connect(clusterId, clientId, stan.NatsConn(conn),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}))
	if err != nil {
		return nil, err
	}

	om := &OrderManager{
		cache:   ch,
		storage: strg,
		Broker:  sc,
	}

	return om, nil
}

func (o *OrderManager) ListenOrders(ctx context.Context) {
	go func() {
		sub, err := o.Broker.Subscribe("wbOrders", func(m *stan.Msg) {
			var order model.Order
			err := json.Unmarshal(m.Data, &order)
			if err != nil {
				log.Println(err)
				return
			}

			if !order.Validate() {
				log.Println("invalid order")
				return
			}

			err = o.cache.PutOrder(order)
			if err != nil {
				log.Println(err)
				return
			}

			err = o.storage.CreateOrder(ctx, order)
			if err != nil {
				log.Println(err)
				return
			}
		})
		if err != nil {
			log.Fatal(err)
			return
		}
		<-ctx.Done()
		sub.Close()
	}()
}
