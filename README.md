# alex-metrics-service

Сервис сбора runtime-метрик: HTTP-сервер хранит gauge/counter, агент периодически снимает `runtime.MemStats` и отправляет их на сервер.

Часть трека Яндекс.Практикума «Сервер сбора метрик и алертинга».

## Требования

- Go 1.26+

## Быстрый старт

```bash
# сервер (по умолчанию localhost:8080)
go run ./cmd/server

# агент (в другом терминале)
go run ./cmd/agent
```

Сборка:

```bash
go build -o bin/server ./cmd/server
go build -o bin/agent ./cmd/agent
```

## Флаги

| Флаг | По умолчанию | Описание |
|------|--------------|----------|
| `-a` | `localhost:8080` | адрес сервера (listen для server, target для agent) |
| `-p` | `2` | интервал опроса метрик агентом (секунды) |
| `-r` | `10` | интервал отправки метрик на сервер (секунды) |

Примеры:

```bash
go run ./cmd/server -a localhost:8080
go run ./cmd/agent -a localhost:8080 -p 2 -r 10
```

## HTTP API

| Метод | Путь | Описание |
|-------|------|----------|
| `POST` | `/update/{type}/{name}/{value}` | обновить метрику (`gauge` / `counter`) |
| `GET` | `/value/{type}/{name}` | получить значение метрики |
| `GET` | `/` | HTML-список всех метрик |

Примеры:

```bash
# counter
curl -X POST http://localhost:8080/update/counter/PollCount/1
curl http://localhost:8080/value/counter/PollCount

# gauge
curl -X POST http://localhost:8080/update/gauge/Alloc/12345.6
curl http://localhost:8080/value/gauge/Alloc

# все метрики
curl http://localhost:8080/
```

Типы метрик:

- **counter** — целое число, при обновлении **суммируется** с текущим значением
- **gauge** — число с плавающей точкой, при обновлении **перезаписывается**

## Структура

```
cmd/
  server/          # точка входа сервера
  agent/           # точка входа агента
internal/
  app/             # сборка и запуск HTTP-сервера
  agent/           # обёртка агента
  config/          # флаги конфигурации
  handler/         # HTTP-handlers + HTML-шаблоны
  service/         # бизнес-логика метрик и агента
  repository/      # доступ к хранилищу
  storage/         # in-memory storage
  model/           # модели метрик
```

## Тесты

```bash
go test ./...
```

Локальные автотесты Практикума (бинарник в корне репозитория):

```bash
./metricstest-darwin-arm64 -test.v -test.run=^TestIteration1$ \
  -agent-binary-path=cmd/agent/agent \
  -binary-path=cmd/server/server \
  -source-path=.
```

Ветки для CI называйте `iterN` (например `iter3`) — так запускаются автотесты инкрементов с 1 по N. Подробнее: [go-autotests](https://github.com/Yandex-Practicum/go-autotests).
