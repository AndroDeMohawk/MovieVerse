# 🎬 Movie Service

> **Status:** 🚧 Active Development

A high-performance backend service for a movie and TV series platform (similar to IMDb), built using microservice architecture and a modern Go tech stack.

---

## 🛠 Tech Stack

* **Language:** Go 1.27+
* **Database:** PostgreSQL (Driver: `pgx/v5`, Generator: `sqlc`)
* **Transport:** gRPC (contracts hosted in a dedicated repository)
* **Caching & Storage:** Redis
* **Message Broker:** Apache Kafka (asynchronous processing for comments)
* **Logging & Configuration:** `log/slog`, `cleanenv`
* **Task Runner:** `go-task` (`Taskfile.yml`)

---

## 🏗 Project Architecture

The project follows **Clean Architecture** principles to maintain strict layer separation:
```text
MovieVerse/
├── cmd/app/          # Entry point, dependency injection, and server startup
├── internal/
│   ├── config/       # Application configuration loader
│   ├── repository/   # Data access layer (PostgreSQL via sqlc, Redis)
│   ├── usecase/      # Core business logic
│   ├── transport/    # gRPC servers and Kafka consumers/producers
│   └── client/       # gRPC clients for external microservices (e.g., Auth SSO)
├── migrations/       # PostgreSQL SQL migration files
└── Taskfile.yml      # Task automation runner
```
---

## 📋 Roadmap

- [x] **Database Design:** Formulated schema for profiles, movies, genres, ratings, and comments.
- [x] **gRPC Contracts:** Extracted `.proto` definitions into an independent repository for decoupled generation.
- [x] **Data Access Layer:** Configured type-safe Go code generation via `sqlc` with `pgx/v5`.
- [x] **Application Skeleton:** Built modular configuration loading (`local.yaml`) and structured logging (`slog`).
- [ ] **Authentication:** Integrate gRPC SSO Auth microservice using a custom JWT Interceptor.
- [ ] **Core Business Logic (Usecase):** Implement CRUD operations for movies, genres, and user profiles.
- [ ] **Kafka Integration:** Asynchronous event streaming and worker processing for movie comments.
- [ ] **Redis Caching:** In-memory caching for Top-100 lists and real-time view counts.

---

## 🚀 Local Development

### Prerequisites

* Go `1.27+`
* Taskfile (`go-task`)
* PostgreSQL, Redis, and Apache Kafka running locally or via Docker

### Quick Start

1. Clone the repository:
   git clone https://github.com/AndroDeMohawk/MovieVerse.git
   cd MovieVerse

2. Run the application:
   go run cmd/app/main.go -config=config/local.yaml

3. Generate SQL query code (using sqlc):
   task sqlc
