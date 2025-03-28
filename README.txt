Список команд: 

=========================================================================================================================================
Сущность: warehouses (Склады)
Create

$body = @{
    name = "zemla"
    address = "Federation_st_15"
description = "craft_beer_pub"
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:8080/api/warehouse" `
                  -Method Post `
                  -ContentType "application/json" `
                  -Body $body
-----------------------------------------------------------------------------------------------------------------------------------------
delete
$warehouseId = "2179b599-bb0d-4fae-835a-d509ecbfe8c0"
Invoke-RestMethod -Uri "http://localhost:8080/api/warehouse/delete/$warehouseId" -Method Delete
-----------------------------------------------------------------------------------------------------------------------------------------
update
$warehouseId = "b55b9a1a-74ce-45a7-981e-b095ed475607"
$body = @{ location = "Federation 15" } | ConvertTo-Json -Depth 10
Invoke-RestMethod -Uri "http://localhost:8080/api/warehouse/update/$warehouseId" `
                  -Method Put `
                  -Body $body `
                  -ContentType "application/json"
-----------------------------------------------------------------------------------------------------------------------------------------
read
Invoke-RestMethod -Uri "http://localhost:8080/api/warehouses" -Method Get
=========================================================================================================================================
Сущность: products (товары)

create
$body = @{
    name        = "Opora"
    description = "Dark_stout"
    attributes  = @{ "ALC" = "4.0%"; "SOL" = "15%"}
    weight      = 0.5
    barcode     = "1234567890126"
} | ConvertTo-Json -Depth 3

Invoke-RestMethod -Uri "http://localhost:8080/api/product" -Method Post -Headers @{"Content-Type" = "application/json"} -Body $body
-----------------------------------------------------------------------------------------------------------------------------------------
delete
$ProductId = "dfebd807-0058-49ef-9f96-2cf21e575a2b"
Invoke-RestMethod -Uri "http://localhost:8080/api/product/delete/$ProductId " -Method DELETE
-----------------------------------------------------------------------------------------------------------------------------------------
update
$productId = "4a5a73ff-da94-42d2-914f-ac2be241e03c"
$body = @{
    name = "Updated Name Only3"
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "http://localhost:8080/api/product/update/$productId" `
                  -Method Put `
                  -Body $body `
                  -ContentType "application/json"
-----------------------------------------------------------------------------------------------------------------------------------------
read
Invoke-RestMethod -Uri "http://localhost:8080/api/products" -Method Get
=========================================================================================================================================
Сущность: inventory (инвенторизация)

Создание связи товара и склада (указание цены)

$body = @{
    product_id   = "14cce754-3fbc-46e1-b3a0-7e7508981a42"
    warehouse_id = "b55b9a1a-74ce-45a7-981e-b095ed475607"
    quantity     = 100
    price        = 190.00
    discount     = 5.00
} | ConvertTo-Json -Depth 3

Invoke-RestMethod -Uri "http://localhost:8080/api/inventory" -Method POST -Headers @{"Content-Type" = "application/json"} -Body $body
-----------------------------------------------------------------------------------------------------------------------------------------
Обновление количества товара на складе (приход)

$warehouseId = "b55b9a1a-74ce-45a7-981e-b095ed475607"
$productId = "4a5a73ff-da94-42d2-914f-ac2be241e03c"

$body = @{
    quantity = 50
} | ConvertTo-Json -Depth 3

Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/update/$warehouseId/$productId" -Method PUT -Headers @{"Content-Type" = "application/json"} -Body $body
----------------------------------------------------------------------------------------------------------------------------------------
Установка скидки на список товаров

$warehouseId = "b55b9a1a-74ce-45a7-981e-b095ed475607"

$body = @{
    product_ids = @("4a5a73ff-da94-42d2-914f-ac2be241e03c")
    discount    = 3
} | ConvertTo-Json -Depth 3

Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/discount/$warehouseId" -Method PUT -Headers @{"Content-Type" = "application/json"} -Body $body
----------------------------------------------------------------------------------------------------------------------------------------
Получение списка товаров на складе

Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/b55b9a1a-74ce-45a7-981e-b095ed475607?limit=10&offset=0" -Method GET
----------------------------------------------------------------------------------------------------------------------------------------
Получение информации о товаре на складе

$warehouseId = "b55b9a1a-74ce-45a7-981e-b095ed475607"
$productId = "4a5a73ff-da94-42d2-914f-ac2be241e03c"

Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/$warehouseId/$productId" -Method GET
----------------------------------------------------------------------------------------------------------------------------------------
Подсчёт стоимости корзины

$warehouseId = "b55b9a1a-74ce-45a7-981e-b095ed475607"

$body = @{
    items = @{
        "14cce754-3fbc-46e1-b3a0-7e7508981a42" = 2;
	"4a5a73ff-da94-42d2-914f-ac2be241e03c" = 3
    }
} | ConvertTo-Json -Depth 3

Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/calculate/$warehouseId" -Method POST -Headers @{"Content-Type" = "application/json"} -Body $body
----------------------------------------------------------------------------------------------------------------------------------------
Покупка товаров со склада

$warehouseId = "b55b9a1a-74ce-45a7-981e-b095ed475607"

$body = @{
    items = @{
        "14cce754-3fbc-46e1-b3a0-7e7508981a42" = 2;
	"4a5a73ff-da94-42d2-914f-ac2be241e03c" = 3
    }
} | ConvertTo-Json -Depth 3

Invoke-RestMethod -Uri "http://localhost:8080/api/inventory/purchase/$warehouseId" -Method POST -Headers @{"Content-Type" = "application/json"} -Body $body
========================================================================================================================================
Сущность: analytics (аналитика)

Выводит список товаров, проданных со склада, с количеством и общей суммой продаж.

$warehouseId = "b55b9a1a-74ce-45a7-981e-b095ed475607"
Invoke-RestMethod -Uri "http://localhost:8080/api/analytics/$warehouseId" -Method GET
----------------------------------------------------------------------------------------------------------------------------------------
Выводит 10 складов, которые продали товаров на самую большую сумму.

Invoke-RestMethod -Uri "http://localhost:8080/api/analytics/top" -Method GET
========================================================================================================================================
