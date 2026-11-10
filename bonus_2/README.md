# Starter СР-2

Реализуйте TODO в пакетах `service` и `repository`. Публичные имена и сигнатуры
менять нельзя: их использует преподавательский набор тестов.

```bash
go test ./...
go test -cover ./...
gofmt -w .
```

Перед сдачей опишите в этом README направление зависимостей между пакетами и
приведите итоговый процент покрытия бизнес-логики.

## Выполненная работа

Направление зависимостей

repository -> service -> model

- `service` импортирует `model`. В нём лежат тарифы, ошибки, калькулятор и интерфейсы `DeliveryPricer` и `TariffRepository`.
- `repository` импортирует `service`. Есть хранилище `Memory`, которое возвращает тарифы типа `DeliveryPricer` из `service`.
- `service` не импортирует `repository`, чтобы не было цикла. Работа идёт через интерфейс `TariffRepository`. 

Тариф в копейках и считается следующим образом:
`база + км * ставка за км + кг * ставка за кг`.

Вес и дистанция должны быть положительными. Дистанция не больше максимальное для тарифа (иначе `ErrDistanceTooLarge`). Граммы округляются вверх до целых кг.

## Покрытие
```text
go test -cover ./...
?       example.com/delivery/model      [no test files]
ok      example.com/delivery/repository 0.003s  coverage: 100.0% of statements
ok      example.com/delivery/service    0.003s  coverage: 100.0% of statements
```
