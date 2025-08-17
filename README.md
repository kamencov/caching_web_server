# 📂 Caching Web Server

Демонстрационный проект на Go для работы с документами: хранение, загрузка и выдача файлов через HTTP API.  
Используются **PostgreSQL** и **MinIO** для хранения данных и файлов.

## 🚀 Запуск проекта

Перед запуском убедитесь, что установлен **Docker** и **docker compose**.

### 1. Настройка окружения
В корне проекта уже есть `.env` с примером конфигурации:

```env
ADMIN_TOKEN="test"
LOG_LEVEL="0"
ADDR=":8080"
TOKEN_SALT="Document"
MAX_SIZE_FILE="50"

⚠️ В реальных условиях значения должны храниться безопасно (например, в переменных окружения или секретах).

### 2. Сборка и запуск

Используйте Makefile для удобства:
# Запуск в Docker
make run_docker
Эта команда:
	•	поднимет контейнеры postgres, minio и golang
	•	соберёт образ приложения из Dockerfile
	•	запустит проект на http://localhost:8080

# Остановка и удаление контейнеров
make stop_docker

# Запуск тестирования с покрытием и выводом в html
make cover

🛠️ Стек технологий
	•	Go
	•	PostgreSQL
	•	MinIO
	•	Docker

📦 Сервисы
	•	API: http://localhost:8080
	•	PostgreSQL: localhost:5432 (БД documents)
	•	MinIO:
	•	API: http://localhost:9000
	•	Консоль: http://localhost:9001
	•	Логин: admin / password
