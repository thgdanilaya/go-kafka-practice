package repository

import (
	"context"
	"errors"
	"time"

	"wbl0/internal/model"
	"wbl0/internal/queries"
)

func (repo *Repository) GetOrder(ctx context.Context, id string) (model.Order, error) {
	var order model.Order
	err := repo.db.QueryRow(ctx, queries.SearchById, id).Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDT, &order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		return model.Order{}, err
	}

	rows, err := repo.db.Query(ctx, queries.GetItems, id)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	order.Items = order.Items[:0]
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status); err != nil {
			return model.Order{}, err
		}
		order.Items = append(order.Items, item)
	}
	return order, rows.Err()
}

func (repo *Repository) ListRecent(ctx context.Context, limit int) ([]model.Order, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be > 0")
	}
	rows, err := repo.db.Query(ctx, queries.GetUids, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make([]string, 0, limit)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]model.Order, 0, len(ids))
	for _, id := range ids {
		ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
		order, err := repo.GetOrder(ctx2, id)
		cancel()
		if err == nil {
			out = append(out, order)
		}
	}
	return out, nil
}
