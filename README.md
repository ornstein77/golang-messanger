# Серверная часть мессенджера

Небольшая серверная часть мессенджера, написанная на Go. Пользователи регистрируются, указывая **никнейм, пароль и номер телефона**, затем входят в систему и получают токен сессии. После этого они могут просматривать список других пользователей и обмениваться личными сообщениями. В репозитории находится **только серверная часть** — без пользовательского интерфейса.

## Технологии

- Go 1.25, маршрутизатор `net/http` из стандартной библиотеки с поддержкой маршрутов по HTTP-методам (`ServeMux`)
- PostgreSQL через `jackc/pgx/v5` и `pgxpool`
- Версионируемые SQL-миграции (`golang-migrate`)
- Конфигурация из переменных окружения (`kelseyhightower/envconfig`)
- Структурированное логирование (`go.uber.org/zap`)
- Проверка входящих запросов (`go-playground/validator/v10`)
- Документация Swagger / OpenAPI (`swaggo/swag` и `swaggo/http-swagger`)
- Хеширование паролей с помощью bcrypt (`golang.org/x/crypto`)

## Архитектура

```text
cmd/app                     # точка входа (main.go)
docs                        # сгенерированная документация Swagger (docs.go)
internal/
  core/
    config                  # конфигурация из переменных окружения
    domain                  # основные сущности (User, Session, Message)
    errors                  # типизированные ошибки приложения и их сопоставление с HTTP-ответами
    logger                  # создание логгера zap
    repository/postgres/
      pool                  # создание пула подключений pgxpool
      pgx                   # общий интерфейс Querier и обработчики ошибок PostgreSQL
    transport/http/
      middleware            # обработчики recover, логирования, CORS и аутентификации
      request               # декодирование JSON и проверка запросов
      response              # оболочка JSON-ответа или ошибки
      server                # сборка маршрутизатора и жизненный цикл http.Server
      types                 # оболочка ответов и вспомогательные средства для контекста аутентификации
  features/
    auth                    # регистрация, вход, выход и проверка сессии
      repository | service | transport
    users                   # список пользователей и профиль текущего пользователя
      repository | service | transport
    messages                # отправка сообщений и история переписки
      repository | service | transport
migrations                  # версионируемые SQL-миграции (golang-migrate)
out/
  logs                      # файлы журналов приложения
  pgdata                    # данные PostgreSQL (точка подключения тома Docker)
.env                        # локальная конфигурация
Каждая функциональная часть организована как вертикальный срез с тремя слоями, взаимодействующими друг с другом только через интерфейсы: repository (SQL-запросы) → service (бизнес-логика) → transport (HTTP).
Конфигурация
Все параметры читаются из переменных окружения (см. файл .env). Основные значения:
Переменная	Значение по умолчанию	Описание
HTTP_PORT	8080	Порт для HTTP-подключений
POSTGRES_HOST	localhost	Адрес сервера PostgreSQL
POSTGRES_PORT	5432	Порт PostgreSQL
POSTGRES_USER	todoapp	Пользователь PostgreSQL
POSTGRES_PASSWORD	todoapp	Пароль PostgreSQL
POSTGRES_DB	messenger	Имя базы данных
SESSION_TTL	720h	Срок действия токена сессии
LOG_LEVEL	info	Уровень логирования zap


Начало работы
Только при первом запуске: загрузите зависимости (требуется доступ к сети).
make tidy

1. Запустите PostgreSQL, примените миграции и включите перенаправление порта
make compose-up
Команда запускает todoapp-postgres, задачу todoapp-postgres-migrate для применения миграций из каталога migrations/ и перенаправляет подключения к PostgreSQL на localhost:5432 с помощью alpine/socat.
2. Запустите сервер
make run
API будет доступен по адресу http://localhost:8080.
- Проверка работоспособности: GET /health
- Интерфейс Swagger: http://localhost:8080/swagger/index.html
Повторная генерация документации Swagger
Базовая версия docs/docs.go включена в репозиторий. Чтобы заново сгенерировать документацию по аннотациям в коде, выполните:
make swagger          # использует локальную программу swag
# или
make swagger-docker   # использует сервис Swagger в Docker Compose
Обзор API
Базовый путь: /api/v1. Все ответы имеют оболочку { "data": ... } или { "error": ... }. Для доступа к защищённым конечным точкам требуется заголовок Authorization: Bearer <token>.
Метод	Путь	Аутентификация	Описание
POST	/auth/register	нет	Регистрация (никнейм, пароль, телефон)
POST	/auth/login	нет	Вход в систему; возвращает токен сессии
POST	/auth/logout	да	Отзыв текущего токена
GET	/users	да	Получение списка других пользователей
GET	/users/me	да	Получение профиля текущего пользователя
POST	/messages	да	Отправка сообщения пользователю
GET	/messages?with=<uuid>	да	Получение переписки с пользователем


Пример последовательности действий
# Регистрация
curl -s http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"nickname":"alice","password":"s3cr3t-pass","phone":"+14155552671"}'

# Вход в систему -> скопируйте data.token
curl -s http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"nickname":"alice","password":"s3cr3t-pass"}'

TOKEN=... # токен из ответа на вход

# Получение списка пользователей, чтобы выбрать собеседника
curl -s http://localhost:8080/api/v1/users -H "Authorization: Bearer $TOKEN"

# Отправка сообщения
curl -s http://localhost:8080/api/v1/messages \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"recipient_id":"<uuid-другого-пользователя>","body":"Привет!"}'

# Просмотр переписки
curl -s "http://localhost:8080/api/v1/messages?with=<uuid-другого-пользователя>" \
  -H "Authorization: Bearer $TOKEN"
Команды Make
Выполните make help, чтобы увидеть полный список команд. Наиболее используемые:
- make tidy — загрузка и упорядочивание зависимостей
- make run — запуск сервера
- make build — сборка bin/app
- make compose-up — запуск PostgreSQL, миграций и перенаправления порта
- make compose-down — остановка сервисов
- make migrate-up — применение миграций
- make swagger — повторная генерация документации Swagger
Примечания
- Для аутентификации используются непрозрачные токены сессий (UUID). Они хранятся в таблице sessions и проверяются при каждом защищённом запросе.
- Пароли хешируются с помощью bcrypt и никогда не возвращаются через API.
- Изменения схемы базы данных версионируются в каталоге migrations/ и применяются с помощью golang-migrate.
- Образ PostgreSQL 18 хранит данные в структуре каталогов, зависящей от версии. Поэтому в файле Compose один том подключается к /var/lib/postgresql (сопоставленному с out/pgdata), а не к прежнему пути /var/lib/postgresql/data.
