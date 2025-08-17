# 📂 Caching Web Server

Демонстрационный проект на **Go** для работы с документами: хранение, загрузка и выдача файлов через HTTP API.  
Используются **PostgreSQL** и **MinIO** для хранения данных и файлов.

---

## 🚀 Запуск проекта

Перед запуском убедитесь, что установлен **Docker** и **Docker Compose**.

### 1️⃣ Настройка окружения

В корне проекта уже есть файл `.env` с примером конфигурации:

```env
ADMIN_TOKEN="test"
LOG_LEVEL="0"
ADDR=":8080"
TOKEN_SALT="Document"
MAX_SIZE_FILE="50"

⚠️ В реальных условиях значения должны храниться безопасно (например, через переменные окружения или секреты).

⸻

2️⃣ Сборка и запуск

Используйте Makefile для удобства:

# Запуск в Docker
make run_docker

Эта команда:
	•	Поднимет контейнеры postgres, minio и golang
	•	Соберёт образ приложения из Dockerfile
	•	Запустит проект на http://localhost:8080

# Остановка и удаление контейнеров
make stop_docker

# Запуск тестирования с покрытием и выводом в HTML
make cover

🛠️ Технологический стек
	•	Go
	•	PostgreSQL
	•	MinIO
	•	Docker / Docker Compose

📦 Сервисы

•	API: http://localhost:8080
•	PostgreSQL: localhost:5432 (БД documents)
•	MinIO:
•	API: http://localhost:9000
•	Консоль: http://localhost:9001
•	Логин: admin / password

🗂️ Визуальная схема Docker Compose

┌───────────────┐          ┌───────────────┐
│   Golang App  │─────────▶│  PostgreSQL   │
│  (API Server) │          │  (DB)         │
│ localhost:8080│          │ localhost:5432│
└───────┬───────┘          └───────────────┘
        │
        │
        ▼
┌───────────────┐
│    MinIO      │
│ (Object Store)│
│localhost:9000 │
└───────────────┘
