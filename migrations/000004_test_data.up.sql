INSERT INTO inventory (id, product_id, warehouse_id, quantity, price, discount)
VALUES
    (gen_random_uuid(), '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 100, 50.00, 10.0),
    (gen_random_uuid(), '33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 200, 30.00, 0.0);
