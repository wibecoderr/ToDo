# 📝 ToDo — REST API Backend

A fully functional **ToDo application backend** built with **Go** and **PostgreSQL**. Clean folder structure, Docker support, and middleware included — ready to connect to any frontend.

---

## 🚀 Tech Stack

| Layer | Technology |
|-------|------------|
| Language | Go (Golang) |
| Database | PostgreSQL |
| Containerization | Docker / Docker Compose |

---

## 📁 Project Structure

```
ToDo/
├── cmd/              # Application entry point
├── database/         # DB connection and queries
├── handler/          # HTTP route handlers
├── model/            # Data models / structs
├── middleware.go     # Middleware (auth, logging, etc.)
├── utils.go          # Utility/helper functions
├── docker-compose.yaml
├── go.mod
└── go.sum
```

---

## ⚙️ Getting Started

### Prerequisites

- [Go](https://go.dev/dl/) 1.20+
- [Docker](https://www.docker.com/) & Docker Compose
- PostgreSQL (or use the Docker setup below)

---

### 1. Clone the Repository

```bash
git clone https://github.com/wibecoderr/ToDo.git
cd ToDo
```

### 2. Run with Docker (Recommended)

The easiest way — spins up the app and PostgreSQL together:

```bash
docker-compose up -d
```

### 3. Run Locally (Without Docker)

Set up your PostgreSQL database, then:

```bash
# Install dependencies
go mod tidy

# Run the app
go run ./cmd
```

Make sure your database connection details are configured correctly in the `database/` folder or via environment variables.

---

## 🔌 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/todos` | Get all todos |
| GET | `/todos/:id` | Get a single todo |
| POST | `/todos` | Create a new todo |
| PUT | `/todos/:id` | Update a todo |
| DELETE | `/todos/:id` | Delete a todo |

> Endpoints may vary — check `handler/` folder for the full list.

---


## 📦 Dependencies

Dependencies are managed via Go modules. Install them with:

```bash
go mod tidy
```

---

## 🐳 Docker Compose

The `docker-compose.yaml` sets up both the Go app and PostgreSQL in one command:

```bash
docker-compose up --build      # Start everything
docker-compose down            # Stop everything
docker-compose down -v         # Stop and remove volumes
```

---

## 👤 Author

**wibecoderr** — [GitHub](https://github.com/wibecoderr)

---

## 📄 License

This project is open source. Feel free to use and modify.
