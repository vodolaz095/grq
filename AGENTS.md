# AGENTS.md

## Основные методы структуры `RedisQueue` (Client)

Структура `RedisQueue` предоставляет следующие методы:

### Управление клиентом
| Метод | Описание |
|-------|----------|
| `Ping(ctx context.Context) error` | Проверка соединения с Redis |
| `GetID() string` | Получение ID клиента |
| `String() string` | Встроение строкового представления |
| `GetQueueName() string` | Получение имени очереди |
| `Close() error` | Закрытие соединения |

### Публикация задач (publish.go)
| Метод | Описание |
|-------|----------|
| `Publish(initialCtx context.Context, p any) error` | Публикация задачи в очередь |
| `PublishFirst(initialCtx context.Context, p any) error` | Публикация задачи в начало очереди |
| `Count(initialCtx context.Context) (int64, error)` | Получение количества задач в очереди |
| `Purge(initialCtx context.Context) error` | Очистка всей очереди |

### Потребление задач (consume.go)
| Метод | Описание |
|-------|----------|
| `SetHeartbeat(interval time.Duration)` | Установка интервала heartbeat |
| `SetConsumerTimeout(interval time.Duration)` | Установка таймаута потребителя |
| `GetTask(initialCtx context.Context) (string, bool, error)` | Получение следующей задачи |
| `Age() (time.Duration, error)` | Получение возраста задачи |
| `ListConsumers(initialCtx context.Context) (map[string]time.Duration, error)` | Получение списка активных потребителей |
| `ConsumeConcurrently(initialCtx context.Context, worker WorkerFunc, concurrency int) error` | Событие задачи с несколькими воркерами |

### Отложенные задачи (defer.go)
| Метод | Описание |
|-------|----------|
| `DeferAt(initialCtx context.Context, on time.Time, p any) error` | Запланировать задачу на конкретное время |
| `DeferAfter(initialCtx context.Context, delay time.Duration, p any) error` | Запланировать задачу через интервал |
| `ConsumeDeffered(initialCtx context.Context) (string, bool, error)` | Потребить готовую deferred задачу |

### Внутренние методы
- `presence(ctx context.Context) error` — проверка присутствия
- `wrapWorker(input WorkerFunc) WorkerFunc` — обертка для worker функций

### Ошибки
| Ошибка | Описание |
|--------|----------|
| `ErrWrongDefer` | Ошибка при попытке запланировать задачу в прошлом |

## Запуск компонентов для тестирования кода
Для корректного запуска компонентов и тестирования кода в проекте, следуйте следующим шагам:

1. **Подготовка окружения**:
   - Убедитесь, что у вас установлены все необходимые зависимости проекта.
   - Запустите команду `make deps` для установки всех зависимостей, так как это Go-проект.
   - Если установлен `docker`, то запустите команду `make docker/up`
   - Если установлен `podman`, то запустите команду `make podman/up`
   - Убедитесь, что база данных `redis` доступна, выполнив команду `redis-cli ping`

2. **Запуск тестов**:
   - Убедитесь, что база данных `redis` доступна, выполнив команду `redis-cli ping`
   - Для запуска тестов используйте команду `make test` или `go test ./...`.
   - При необходимости, укажите конкретные тесты или директории для тестирования.

3. **Доступные примеры**:
   - **consumer**: пример потребителя задач из очереди. Запускается командой `go run examples/consumer/main.go`.
   - **publisher**: пример издателя задач в очередь. Запускается командой `go run examples/publisher/main.go`.
   - **full**: пример одновременной работы издателя и потребителя. Запускается командой `go run examples/full/main.go`.

4. **Документация и поддержка**:
   - Для получения дополнительной информации по использованию агентов и работе с проектом, обратитесь к документации в README.md и примерам, предоставленным в проекте.
   - Проект использует Redis в качестве бэкенда для очереди задач.
   - Модуль доступен по адресу `github.com/vodolaz095/grq`.
