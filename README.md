# Alerts2Incidents

Проект направлен на реализацию автоматической обработки инцидентов на основе систем мониторинга (Grafana + Prometheus (VMetrics), Zabbix), с дополнительной возможностью ручного управления инцидентами.

`ВАЖНО! Проект является каркасом, требующим доработки под Ваши нужды.`

## Как развернуть
1. Копируем репозиторий с помощью команды `git clone` 
2. Создаем внутри корневой директории проекта директорию `configs`
3. В этой директории создаем и настраиваем файлы `handler.env`, `server.env`, `bot.env`
4. Из корневой директории проекта билдим и запускаем сервисы командой `docker-compose up -d` 

## Примеры файлов конфигурации
### handler.env (Комментарии удалить, при необходимости)
```
DATABASE_HOST=YOUR_DB_IP_OR_HOST # Имя хоста (если есть DNS) или явный IP адрес
DATABASE_PORT=YOUR_DB_PORT # Порт базы данных
DATABASE_NAME=YOUR_DB_NAME # Имя базы данных
DATABASE_USER=YOUR_DB_USER # Юзер базы данных
DATABASE_PASSWORD=YOUR_DB_USER_PASSWORD # Пасс базы данных
DATABASE_MAX_CONNECTIONS=100 # Минимум 1; Максимум 100

SERVICE_CHANNEL_DATA_MAX_SIZE=100 # Минимум 1; Максимум 100
SERVICE_CHANNEL_ALERTS_MAX_SIZE=100 # Минимум 1; Максимум 100

SERVICE_COLLECTOR_GRAFANA_IS_ACTIVE=true # true / false
SERVICE_COLLECTOR_GRAFANA_INCLUDE_PROMETHEUS_ALERTS=true # true / false
SERVICE_COLLECTOR_GRAFANA_PROMETHEUS_UIDS=YOUR_PROMETHEUS_UIDS # Один или несколько (через пробел в одну строку)
SERVICE_COLLECTOR_GRAFANA_API_URL=YOUR_GRAFANA_URL # Пример: https://example.com (слеш в конце не нужен)
SERVICE_COLLECTOR_GRAFANA_TOKEN=YOUR_GRAFANA_API_TOKEN
SERVICE_COLLECTOR_GRAFANA_COLLECT_INTERVAL=YOUR_COLLECT_INTERVAL # Пример: 5s; Минимум 5s
SERVICE_COLLECTOR_GRAFANA_REQUEST_TIMEOUT=YOUR_REQUEST_TIMEOUT # Пример: 5s; Минимум 1s, максимум 5s

SERVICE_COLLECTOR_ZABBIX_IS_ACTIVE=true # true / false
SERVICE_COLLECTOR_ZABBIX_API_URL=YOUR_ZABBIX_URL # Пример: https://example.com (слеш в конце не нужен)
SERVICE_COLLECTOR_ZABBIX_TOKEN=YOUR_ZABBIX_API_TOKEN
SERVICE_COLLECTOR_ZABBIX_TRIGGER_MIN_LEVEL=YOUR_ZABBIX_TRIGGER_MIN_LEVEL # От 1 до 5 (3 = High)
SERVICE_COLLECTOR_ZABBIX_COLLECT_INTERVAL=YOUR_COLLECT_INTERVAL # Пример: 5s; Минимум 5s
SERVICE_COLLECTOR_ZABBIX_REQUEST_TIMEOUT=YOUR_REQUEST_TIMEOUT # Пример: 5s; Минимум 1s, максимум 5s

SERVICE_PARSER_AGGREGATION_INTERVAL=YOUR_AGGREGATION_INTERVAL # Пример: 5s; Минимум 5s
SERVICE_PARSER_GRAFANA_AM_PARSE_FIELD=YOUR_FIELD # Поле для парсинга алертов Grafana Alertmanager (summary или description)
SERVICE_PARSER_GRAFANA_PROMETHEUS_PARSE_FIELD=YOUR_FIELD # Поле для парсинга алертов Grafana Prometheus (summary или description)

SERVICE_CACHE_INCIDENTS_MAX_SIZE=100 # Размер кэша инцидентов (Минимум 1; Максимум 100)
SERVICE_CACHE_RULES_MAX_SIZE=-1 # Размер кэша правил (-1 = бесконечный; Максимум 100)
```

### server.env (Комментарии удалить, при необходимости)
```
API_SERVER_HOST=YOUR_SERVER_HOST # Хост или IP адрес, на котором будет работать сервер (может быть пустым)
API_SERVER_PORT=8080 # Порт, на котором будет работать сервер (должен совпадать с портом, который открывается в docker контейнере)
API_LDAP_HOST=YOUR_LDAP_HOST # Хост или IP адрес до LDAP сервера
API_LDAP_PORT=YOUR_LDAP_PORT # Порт LDAP сервера
API_LDAP_BASE_DN=YOUR_BASE_DN # Пример: "dc=domain,dc=com"
API_LDAP_ALLOWED_GROUPS=YOUR_ALLOWED_GROUPS # Разрешенные группы для авторизации; Пример: "acs_statuspage_admin acs_statuspage_editor"
API_JWT_SECRET_KEY=YOUR_JWT_SECRET_KEY # Секретный ключ для JWT токенов в формате base64 строки
API_JWT_TOKEN_EXPIRATION_INTERVAL=YOUR_TOKEN_EXPIRATION_INTERVAL # Время жизни JWT токенов; Минимум 1 час, максимум 24 часа; Пример: 8h

DATABASE_HOST=YOUR_DB_IP_OR_HOST # Имя хоста (если есть DNS) или явный IP адрес
DATABASE_PORT=YOUR_DB_PORT # Порт базы данных
DATABASE_NAME=YOUR_DB_NAME # Имя базы данных
DATABASE_USER=YOUR_DB_USER # Юзер базы данных
DATABASE_PASSWORD=YOUR_DB_USER_PASSWORD # Пасс базы данных
DATABASE_MAX_CONNECTIONS=100 # Минимум 1; Максимум 100
```

### bot.env (Комментарии удалить, при необходимости)
```
TELEGRAM_BOT_TOKEN=YOUR_TELEGRAM_TOKEN
TELEGRAM_CHATS=YOUR_TELEGRAM_CHATS # Чаты Telegram в формате "ID1:ThreadID1 ID2"; ThreadID опционально.
TELEGRAM_MESSAGE_CACHE_MAX_SIZE=100 # Размер кэша сообщений Telegram (Минимум 1; Максимум 100)
TELEGRAM_MESSAGE_PARSE_MODE=YOUR_PARSE_MODE # Режим парсинга сообщений Telegram (HTML / Markdown / MarkdownV2)
TELEGRAM_MESSAGE_TEMPLATE_FILEPATH=./templates/bk_incident_ru_mdv2.tmpl # Путь до файла шаблона сообщения (лежат в директории "templates")
TELEGRAM_REQUEST_DELAY=YOUR_DELAY # Интервал задержки перед каждым запросом к Telegram API; Минимум 1s, максимум 30s

DATABASE_HOST=YOUR_DB_IP_OR_HOST # Имя хоста (если есть DNS) или явный IP адрес
DATABASE_PORT=YOUR_DB_PORT # Порт базы данных
DATABASE_NAME=YOUR_DB_NAME # Имя базы данных
DATABASE_USER=YOUR_DB_USER # Юзер базы данных
DATABASE_PASSWORD=YOUR_DB_USER_PASSWORD # Пасс базы данных
DATABASE_MAX_CONNECTIONS=100 # Минимум 1; Максимум 100
```

## Зависимости
* ЯП Golang 1.22+
* Docker + Compose
* PostgreSQL
