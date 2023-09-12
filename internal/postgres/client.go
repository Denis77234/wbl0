package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"

	"awesomeProject/internal/model"
)

type Client struct {
	db *sql.DB
}

func New(cfg Config) (*Client, error) {
	db, err := sql.Open("pgx", cfg.connString())
	if err != nil {
		return nil, err
	}

	c := &Client{db}

	return c, nil
}

func (c Client) Close() error {
	err := c.db.Close()

	return err
}

func (c *Client) CreateOrder(ctx context.Context, ord model.Order) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin:%w", err)
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO orders(order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shard_key, sm_id, date_created, oof_shard) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)",
		ord.OrderUid, ord.TrackNumber, ord.Entry, ord.Locale, ord.InternalSignature, ord.CustomerId, ord.DeliveryService, ord.Shardkey, ord.SmId, ord.DateCreated, ord.Shardkey)
	if err != nil {
		return fmt.Errorf("insert orders:%w", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO deliveries(name, phone, zip, city, address, region, email, order_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)",
		ord.Delivery.Name, ord.Delivery.Phone, ord.Delivery.Zip, ord.Delivery.City, ord.Delivery.Address, ord.Delivery.Region, ord.Delivery.Email, ord.OrderUid)
	if err != nil {
		return fmt.Errorf("insert deliveries:%w", err)
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO payments(transaction, request_id, currency , provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee, order_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)",
		ord.Payment.Transaction, ord.Payment.RequestId, ord.Payment.Currency, ord.Payment.Provider, ord.Payment.Amount, ord.Payment.PaymentDt, ord.Payment.Bank, ord.Payment.DeliveryCost, ord.Payment.GoodsTotal, ord.Payment.CustomFee, ord.OrderUid)
	if err != nil {
		return fmt.Errorf("insert payments:%w", err)
	}

	for i, itm := range ord.Items {
		_, err = tx.ExecContext(ctx, "INSERT INTO items(chrt_id, track_number, price , rid, name, sale, size, total_price, nm_id, brand, status, order_id) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)",
			itm.ChrtId, itm.TrackNumber, itm.Price, itm.Rid, itm.Name, itm.Sale, itm.Size, itm.TotalPrice, itm.NmId, itm.Brand, itm.Status, ord.OrderUid)
		if err != nil {
			return fmt.Errorf("insert item №%v:%w", i, err)
		}

	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("tx commit:%w", err)
	}

	return nil

}

func (c Client) SelectOrders(ctx context.Context) ([]model.Order, error) {
	ids, err := c.db.QueryContext(ctx, "SELECT order_uid FROM orders")
	if err != nil {
		return nil, fmt.Errorf("select ids:%w", err)
	}

	orders := make([]model.Order, 0, 5)

	for ids.Next() {
		var id string

		err = ids.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("scan ids:%w", err)
		}
		ord, err := c.SelectOrderById(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("select order by id:%w", err)
		}

		orders = append(orders, ord)
	}

	return orders, nil
}

func (c *Client) SelectOrderById(ctx context.Context, orderId string) (model.Order, error) {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return model.Order{}, fmt.Errorf("begin:%w", err)
	}

	defer tx.Rollback()

	var ord model.Order
	if err = tx.QueryRowContext(ctx, "SELECT * FROM orders WHERE order_uid = $1", orderId).
		Scan(&ord.OrderUid, &ord.TrackNumber, &ord.Entry, &ord.Locale, &ord.InternalSignature, &ord.CustomerId, &ord.DeliveryService, &ord.Shardkey, &ord.SmId, &ord.DateCreated, &ord.OofShard); err != nil {
		return model.Order{}, fmt.Errorf("select order:%w", err)
	}

	var del model.Delivery
	if err = tx.QueryRowContext(ctx, "SELECT * FROM deliveries WHERE order_id = $1", orderId).
		Scan(&del.Name, &del.Phone, &del.Zip, &del.City, &del.Address, &del.Region, &del.Email, &orderId); err != nil {
		return model.Order{}, fmt.Errorf("select delivery:%w", err)
	}

	ord.Delivery = del

	var pay model.Payment
	if err = tx.QueryRowContext(ctx, "SELECT * FROM payments WHERE order_id = $1", orderId).
		Scan(&pay.Transaction, &pay.RequestId, &pay.Currency, &pay.Provider, &pay.Amount, &pay.PaymentDt, &pay.Bank, &pay.DeliveryCost, &pay.GoodsTotal, &pay.CustomFee, &orderId); err != nil {
		return model.Order{}, fmt.Errorf("select delivery:%w", err)
	}

	ord.Payment = pay

	rows, err := tx.QueryContext(ctx, "SELECT * FROM items WHERE order_id = $1", orderId)
	if err != nil {
		return model.Order{}, fmt.Errorf("select items:%w", err)
	}

	items := make([]model.Item, 0, 5)

	for rows.Next() {
		var item model.Item

		if err = rows.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status, &orderId); err != nil {
			return model.Order{}, fmt.Errorf("select item:%w", err)
		}

		items = append(items, item)
	}

	ord.Items = items

	if err = tx.Commit(); err != nil {
		return model.Order{}, fmt.Errorf("tx commit:%w", err)
	}

	return ord, nil
}
