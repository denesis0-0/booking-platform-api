# booking-platform-api

Backend API для бронирования ресурсов: комнат, рабочих мест, специалистов и оборудования.

Главная техническая идея проекта — **безопасное бронирование при одновременных запросах**. API не позволяет забронировать один и тот же слот дважды, даже если несколько пользователей пытаются сделать это одновременно.

## Возможности

* создание и получение списка ресурсов;
* создание и получение слотов для ресурса;
* получение только доступных слотов;
* создание бронирований;
* отмена бронирований;
* защита от двойного бронирования одного слота;
* хранение данных в PostgreSQL;
* запуск через Docker Compose;
* тест на конкурентное бронирование.

## Стек

* Go
* PostgreSQL
* Docker
* Docker Compose
* pgx / pgxpool
* REST API

## Структура проекта

```text
booking-platform-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── httpapi/
│   └── storage/
├── migrations/
│   └── 001_create_tables.sql
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── README.md
```

## Как устроен проект

В проекте есть три основные сущности:

```text
resources  -> ресурсы, которые можно бронировать
slots      -> временные интервалы для ресурса
bookings   -> бронирования пользователей
```

Ресурсом может быть комната, рабочее место, специалист или оборудование.

Слот принадлежит ресурсу и описывает конкретный временной интервал.

Бронирование принадлежит слоту и имеет статус:

```text
confirmed
cancelled
```

## Защита от двойного бронирования

Проект защищает систему от двойного бронирования на уровне базы данных.

В таблице `bookings` используется partial unique index:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS unique_confirmed_booking_per_slot
ON bookings(slot_id)
WHERE status = 'confirmed';
```

Это означает, что PostgreSQL разрешает только одно активное бронирование для одного слота.

Если бронирование отменяется, его статус становится `cancelled`, и этот же слот снова можно забронировать.

Такой подход важен, потому что целостность данных не зависит только от кода приложения. Даже при одновременных запросах база данных гарантирует, что один слот не будет забронирован дважды.

## API Endpoints

### Health Check

```http
GET /health
```

Пример ответа:

```json
{
  "status": "ok",
  "app": "booking-platform-api",
  "database": "ok"
}
```

### Resources

Создать ресурс:

```http
POST /resources
```

Тело запроса:

```json
{
  "name": "Room 101",
  "type": "room",
  "description": "Small meeting room"
}
```

Получить список ресурсов:

```http
GET /resources
```

### Slots

Создать слот для ресурса:

```http
POST /resources/{resource_id}/slots
```

Тело запроса:

```json
{
  "starts_at": "2026-06-29T10:00:00+02:00",
  "ends_at": "2026-06-29T11:00:00+02:00"
}
```

Получить все слоты ресурса:

```http
GET /resources/{resource_id}/slots
```

Получить только доступные слоты:

```http
GET /resources/{resource_id}/available-slots
```

### Bookings

Создать бронирование:

```http
POST /bookings
```

Тело запроса:

```json
{
  "slot_id": "slot-uuid",
  "user_name": "Timofey"
}
```

Если слот уже забронирован, API вернет:

```http
409 Conflict
```

```json
{
  "error": "slot already booked"
}
```

Получить список бронирований:

```http
GET /bookings
```

Отменить бронирование:

```http
DELETE /bookings/{booking_id}
```

После отмены статус бронирования меняется на `cancelled`, а слот снова становится доступным для бронирования.

## Локальный запуск

Запустить PostgreSQL:

```bash
docker-compose up -d postgres
```

Применить миграции:

```bash
docker exec -i booking-platform-postgres psql -U booking -d booking < migrations/001_create_tables.sql
```

Запустить API:

```bash
go run ./cmd/api
```

Проверить, что приложение работает:

```bash
curl http://localhost:8080/health
```

## Запуск через Docker Compose

Собрать и запустить весь проект:

```bash
docker-compose up --build
```

Или запустить в фоне:

```bash
docker-compose up --build -d
```

Проверить API:

```bash
curl http://localhost:8080/health
```

Если используется Docker Compose v2, вместо `docker-compose` нужно писать:

```bash
docker compose
```

## Переменные окружения

Пример файла `.env.example`:

```env
PORT=8080
DATABASE_URL=postgres://booking:booking@localhost:5432/booking?sslmode=disable
```

Для запуска внутри Docker Compose используется другой `DATABASE_URL`, потому что API подключается к PostgreSQL по имени сервиса `postgres`:

```env
DATABASE_URL=postgres://booking:booking@postgres:5432/booking?sslmode=disable
```

## Тесты

Перед запуском тестов нужно убедиться, что PostgreSQL запущен и миграции применены.

Запуск всех тестов:

```bash
go test ./...
```

В проекте есть тест на конкурентное бронирование:

```text
TestConcurrentBookingSameSlot
```

Он создает один слот и запускает несколько goroutine, которые одновременно пытаются забронировать этот слот.

Ожидаемый результат:

```text
1 успешное бронирование
остальные попытки завершаются ошибкой slot already booked
```

Этот тест показывает, что система корректно обрабатывает одновременные запросы и не допускает двойного бронирования.

## Пример сценария

Создать ресурс:

```bash
curl -X POST http://localhost:8080/resources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Room 101",
    "type": "room",
    "description": "Small meeting room"
  }'
```

Создать слот:

```bash
curl -X POST http://localhost:8080/resources/RESOURCE_ID/slots \
  -H "Content-Type: application/json" \
  -d '{
    "starts_at": "2026-06-29T10:00:00+02:00",
    "ends_at": "2026-06-29T11:00:00+02:00"
  }'
```

Забронировать слот:

```bash
curl -X POST http://localhost:8080/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "slot_id": "SLOT_ID",
    "user_name": "Timofey"
  }'
```

Попробовать забронировать тот же слот еще раз:

```bash
curl -X POST http://localhost:8080/bookings \
  -H "Content-Type: application/json" \
  -d '{
    "slot_id": "SLOT_ID",
    "user_name": "Another User"
  }'
```

Ожидаемый ответ:

```json
{
  "error": "slot already booked"
}
```

Отменить бронирование:

```bash
curl -X DELETE http://localhost:8080/bookings/BOOKING_ID
```

После отмены слот снова появится в списке доступных:

```bash
curl http://localhost:8080/resources/RESOURCE_ID/available-slots
```

## Статус проекта

Реализовано:

* REST API;
* подключение к PostgreSQL;
* миграция базы данных;
* Dockerfile;
* Docker Compose;
* создание ресурсов;
* создание слотов;
* создание бронирований;
* отмена бронирований;
* список доступных слотов;
* защита от двойного бронирования;
* тест на конкурентное бронирование.

Возможные улучшения:

* JWT-аутентификация;
* полноценные пользователи;
* автоматический запуск миграций при старте приложения;
* пагинация;
* middleware для логирования запросов;
* OpenAPI / Swagger-документация;
* CI pipeline через GitHub Actions.
