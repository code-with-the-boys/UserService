# User Service: локальный запуск, Swagger, примеры curl

HTTP API (grpc-gateway) по умолчанию слушает **`:8080`**, gRPC User — **`:9090`**. Train вызывается по gRPC, адрес задаётся **`TRAIN_GRPC_ADDR`**.

---

## Как запустить локально

### 1. Postgres и Redis (UserService)

Из каталога `UserService`:

```bash
cd UserService

# если порт 6379 на хосте занят — пробрось другой (ниже пример 6380)
REDIS_PORT=6380 docker compose up -d postgres redis
```

Порты и пароли смотри в `docker-compose.yml` и в `.env`. У контейнера Redis проверить фактические переменные:

```bash
docker inspect fitness_redis --format '{{range .Config.Env}}{{println .}}{{end}}' | grep REDIS
```

### 2. Миграции БД (один раз)

```bash
cd UserService/service

goose -dir internal/db/migrations postgres \
  "host=127.0.0.1 port=5433 user=postgres password=postgres dbname=UserService sslmode=disable" up
```

Подставь `user` / `password` / `port`, если у тебя другие (как в `.env`).

### 3. Train (опционально, для маршрутов `/api/v1/train/...`)

```bash
cd TrainService
docker compose up -d
```

По умолчанию gRPC слушает **`127.0.0.1:50051`**.

### 4. Запуск User-сервиса

Из каталога `UserService/service` (удобно завести `.env` рядом с `main` — подхватится `godotenv`):

```bash
cd UserService/service

export POSTGRES_HOST=127.0.0.1
export POSTGRES_PORT=5433
export POSTGRES_USER=fitness_user
export POSTGRES_PASSWORD=user_password
export POSTGRES_DB=UserService
export POSTGRES_SSLMODE=disable

export REDIS_HOST=127.0.0.1
export REDIS_PORT=6380
export REDIS_PASSWORD=secure_redis_password
export REDIS_DB=12

export JWT_ACCESS_SECRET='длинный-секрет-для-access'
export JWT_REFRESH_SECRET='длинный-секрет-для-refresh'

# опционально: PaymentService на :8000 → клиент бьёт в шлюз /api/v1/payment/... с тем же JWT
# export PAYMENT_SERVICE_HTTP_URL='http://127.0.0.1:8000'

# прокси Payment на Litestar (смоук PaymentService/scripts/smoke-e2e.sh через шлюз)
export PAYMENT_SERVICE_HTTP_URL='http://127.0.0.1:8000'

HTTP_ADDR=':8080' GRPC_ADDR=':9090' TRAIN_GRPC_ADDR='127.0.0.1:50051' go run ./cmd/main.go
```

---

## Swagger (OpenAPI)

Сгенерированная спецификация лежит в репозитории:

**`UserService/protos/gen/openapiv2/api.swagger.json`**

Отдельного Swagger UI в процессе приложения нет — это файл. Обновить после изменений в `.proto`:

```bash
UserService/protos/api/gen_buf.sh
```

(Нужны `buf` и плагины `protoc-gen-go`, `protoc-gen-go-grpc`, `protoc-gen-grpc-gateway`, `protoc-gen-openapiv2` в `$(go env GOPATH)/bin`.)

### Просмотр в Swagger UI (Docker)

```bash
docker run --rm -p 8888:8080 \
  -e SWAGGER_JSON=/foo/api.swagger.json \
  -v /absolute/path/to/UserService/protos/gen/openapiv2/api.swagger.json:/foo/api.swagger.json:ro \
  swaggerapi/swagger-ui
```

Открыть в браузере: **http://localhost:8888**

---

## Примеры curl

Базовый URL: **`http://127.0.0.1:8080`**

### Train health (без JWT)

```bash
curl -sS http://127.0.0.1:8080/api/v1/train/health
```

### Регистрация

Нужен **валидный телефон** (11 цифр после нормализации, начало 7 или 8), иначе валидация вернёт ошибку.

```bash
curl -sS http://127.0.0.1:8080/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"secret123","phone":"79001234567"}'
```

### Логин

В JSON ответа токен в поле **`accessToken`** (camelCase, не `access_token`).

```bash
curl -sS http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"you@example.com","password":"secret123"}'
```

### Запросы с Bearer

```bash
TOKEN='<accessToken из ответа логина>'

curl -sS http://127.0.0.1:8080/api/v1/users/me/id \
  -H "Authorization: Bearer $TOKEN"

curl -sS 'http://127.0.0.1:8080/api/v1/train/plans' \
  -H "Authorization: Bearer $TOKEN"

curl -sS 'http://127.0.0.1:8080/api/v1/train/plans?status=active' \
  -H "Authorization: Bearer $TOKEN"
```

Профиль ( `{user_id}` должен совпадать с пользователем из JWT):

```bash
curl -sS "http://127.0.0.1:8080/api/v1/users/<USER_UUID>/profile" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Заметки

- Защищённые маршруты: заголовок **`Authorization: Bearer <accessToken>`**.
- Тела и ответы в REST — в основном **camelCase** (`userId`, `accessToken`, …).
- Сборка **Docker-образа приложения** с текущим `replace ../protos` в `go.mod` требует, чтобы в контекст сборки попадала папка `protos`; иначе используй локальный `go run` или поправь `docker-compose` / `Dockerfile`.
