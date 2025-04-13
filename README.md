# Сервис для работы с ПВЗ
## Настройка и запуск
1. Клонируйте репозиторий
```
git clone https://github.com/spanwalla/pvz
cd pvz
```
2. Создайте файл `.env` в корне проекта (можете скопировать [`.env.example`](.env.example)).
3. Выполните команду
```
docker compose up --build -d
```
4. Для остановки используйте команду
```
docker compose down --remove-orphans
```
* Для запуска интеграционных тестов выполните команду `make integration-test`.
* Для запуска обычных тестов используйте команду `go test -v ./internal/...`.

## Функциональные требования
Сервис предоставляет API для работы с пунктами выдачи заказов.

Спецификация API: [ссылка](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-spring-2025/swagger.yaml).

Реализована авторизация с JWT-токенами и ролями.

## Нефункциональные требования
* Покрытие сервисов тестами: __94.6%__.
* Реализован интеграционный тест для одного сценария.
* Настроен GitHub Workflows на запуск линтера и тестов.
* Добавлено логирование.
* Подключён Prometheus (метрики называются `points_created_total`, `products_created_total`, `receptions_created_total`).
* Реализован gRPC-метод.
* Настроена генерация DTO для обработчиков из спецификации OpenAPI.