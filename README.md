# Pollz WebSocket Server (Go)

## Setup

1. **Clone and navigate to websocket server**
```bash
cd websocket-go/
```

2. **Install dependencies**
```bash
go mod download
```

3. **Setup environment variables**
```bash
cp .env.example .env
# Edit .env with your database credentials
```

4. **Run with Docker (Recommended)**
```bash
docker-compose up --build
```

**OR run locally:**
```bash
# Ensure PostgreSQL and Redis are running
make run
```

Server will be available at `ws://localhost:1401`

### WebSocket
- `ws://localhost:1401/ws/chat/live` - Main chat WebSocket


