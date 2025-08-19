package queries

const SearchById = `
SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
       o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
       d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
       p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
FROM orders o
JOIN deliveries d ON d.order_uid = o.order_uid
JOIN payments   p ON p.order_uid = o.order_uid
WHERE o.order_uid = $1
`

const GetItems = `
SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
FROM order_items
WHERE order_uid = $1
ORDER BY chrt_id
`

const GetUids = `
SELECT order_uid
FROM orders
ORDER BY date_created DESC
LIMIT $1
`
