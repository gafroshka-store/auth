# Все поднятия приложения, запуск тестов и тд - ЗДЕСЬ
.PHONY: run stop stop-hard run-lint tests

# Запуск контейнеров через docker-compose
run:
	docker-compose up -d

# Остановка контейнеров
stop:
	docker-compose down

# Остановка контейнеров с удалением данных
stop-hard:
	docker-compose down -v

# Линтер
lint:
	golangci-lint run --config .golint.yaml

# Запуск тестов
tests:
	go test -v -cover ./...
