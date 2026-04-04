# Observability Configuration Guide

## ✅ Что было исправлено

### 1. **Grafana Datasources**
   - ✅ Создан `prometheus.yaml` для метрик VictoriaMetrics (тип: `prometheus`)
   - ✅ Исправлен `victoriametrics.yaml` - изменен тип с `victoriametrics-logs-datasource` на `loki` для логов
   - ✅ Добавлена конфигурация `loki.yaml` (уже была)

**Проблема была:** Использовался неправильный плагин для логов, и отсутствовал основной datasource для метрик.

### 2. **VictoriaMetrics Scrape Configuration** (`victoriametrics/scrape.yml`)
   - ✅ Добавлен путь `/metrics` к конфигурации скрейпинга
   - ✅ Добавлены параметры `metrics_path` и `scrape_timeout`

**Проблема была:** VictoriaMetrics пытался скрейпить `app:8080` без указания пути к метрикам.

### 3. **Vector Configuration** (`vector/vector.toml`)
   - ✅ Добавлена отправка логов в VictoriaLogs (основной сток)
   - ✅ Сохранена отправка в Loki (резервный сток)

**Структура потоков:**
```
Docker Logs → Parse JSON → VictoriaLogs (HTTP) + Loki
```

### 4. **Docker Compose Зависимости и Healthchecks**

#### App
- ✅ Добавлен healthcheck через `/metrics` эндпоинт
- ✅ Задержка запуска 15s для полной инициализации

#### VictoriaMetrics
- ✅ Добавлена зависимость от `app` (service_healthy)
- ✅ Добавлен healthcheck
- ✅ Гарантирует что приложение готово перед скрейпингом

#### VictoriaLogs
- ✅ Добавлен healthcheck на `/health`

#### Vector
- ✅ Добавлены зависимости: `victorialogs`, `loki`, `app`
- ✅ Гарантирует что все хранилища готовы перед сбором логов

#### Grafana
- ✅ Добавлены правильные зависимости (service_healthy)
- ✅ Удален плагин `victoriametrics-logs-datasource` (использует встроенный Loki)
- ✅ Добавлены переменные `GF_USERS_ALLOW_SIGN_UP=false` и `GF_PATHS_PROVISIONING`
- ✅ Добавлен healthcheck

#### Loki
- ✅ Добавлен healthcheck

## 📋 Порядок запуска контейнеров

```
1. PostgreSQL & Redis (базы данных)
   ↓
2. App (зависит от баз)
   ↓
3. VictoriaMetrics & VictoriaLogs (зависят от App)
   ↓
4. Vector (зависит от VictoriaLogs, Loki и App)
   ↓
5. Grafana (зависит от VictoriaMetrics, VictoriaLogs и Loki)
```

## 🚀 Как использовать

### Запуск
```bash
docker compose down -v  # Очистить старые данные (если нужно)
docker compose up -d
```

### Проверка статуса
```bash
docker compose ps

# Все контейнеры должны быть в статусе:
# "running" и healthcheck "healthy" (если задан)
```

### Доступ к сервисам

| Сервис | URL | Учетные данные |
|--------|-----|---|
| Grafana | http://localhost:3000 | admin / admin |
| VictoriaMetrics | http://localhost:8428 | - |
| VictoriaLogs | http://localhost:9428 | - |
| Loki | http://localhost:3100 | - |
| App HTTP | http://localhost:8080 | - |
| App gRPC | localhost:9090 | - |

### Проверка компонентов

**Проверить метрики приложения:**
```bash
curl http://localhost:8080/metrics | head -20
```

**Проверить что VictoriaMetrics скрейпит:**
```bash
curl http://localhost:8428/api/v1/targets
```

**Проверить логи в Vector:**
```bash
docker logs fitness_vector --follow
```

**Проверить логи в VictoriaLogs:**
```bash
curl "http://localhost:9428/api/v1/query_range?query=*&start=0&end=$(date +%s)"
```

## 📊 Добавление дашбордов в Grafana

1. Откройте Grafana: http://localhost:3000
2. Перейдите в **Dashboards** → **New** → **New Dashboard**
3. Создайте панели используя:
   - **Datasource:** VictoriaMetrics (для метрик)
   - **Datasource:** VictoriaLogs или Loki (для логов)

**Примеры запросов:**

- Метрики приложения:
```promql
rate(grpc_server_handled_total[5m])
```

- Логи из приложения:
```promql
{container_name="fitness_app"}
```

## 🔧 Переменные окружения (.env)

Убедитесь что в `.env` файле присутствуют:
```bash
GRAFANA_ADMIN_USER=admin
GRAFANA_ADMIN_PASSWORD=admin
POSTGRES_ADMIN_USER=postgres
POSTGRES_ADMIN_PASSWORD=postgres
```

## ✋ Возможные проблемы и решения

### Grafana не может подключиться к VictoriaMetrics
1. Проверьте что VictoriaMetrics запущена: `docker ps | grep victoriametrics`
2. Проверьте connectivity: `docker exec fitness_grafana curl -v http://victoriametrics:8428`
3. Перезагрузите Grafana: `docker restart fitness_grafana`

### Vector отправляет ошибки
1. Проверьте конфиг: `docker exec fitness_vector cat /etc/vector/vector.toml`
2. Проверьте логи: `docker logs fitness_vector`
3. Убедитесь что VictoriaLogs и Loki запущены и healthy

### Нет данных в Grafana
1. Убедитесь что приложение генерирует метрики на `/metrics`
2. Проверьте что VictoriaMetrics скрейпит правильный эндпоинт: `curl http://localhost:8428/api/v1/targets`
3. Дождитесь несколько минут пока метрики начнут собираться

### "data source not found" ошибка
- ✅ Исправлено! Были добавлены все необходимые datasources с правильными типами и URL

## 📁 Структура файлов

```
grafana/
├── provisioning/
│   ├── datasources/
│   │   ├── loki.yaml          # Loki datasource
│   │   ├── prometheus.yaml     # VictoriaMetrics datasource (НОВЫЙ)
│   │   └── victoriametrics.yaml # VictoriaLogs datasource (ИСПРАВЛЕН)
│   └── dashboards/
│       └── dashboards.yaml     # Dashboard provisioning config
└── dashboards/
    └── (ваши JSON дашборды)

vector/
└── vector.toml               # Vector config (ИСПРАВЛЕН - добавлен VictoriaLogs sink)

victoriametrics/
└── scrape.yml               # Prometheus scrape config (ИСПРАВЛЕН - добавлен /metrics path)

docker-compose.yml           # (ИСПРАВЛЕН - добавлены healthchecks и зависимости)
```

## 🎯 Следующие шаги

1. Перезапустите контейнеры с новыми конфигурациями
2. Дождитесь пока все сервисы будут healthy
3. Проверьте что метрики собираются в VictoriaMetrics
4. Создайте дашборды в Grafana
5. Настройте оповещения по необходимости
