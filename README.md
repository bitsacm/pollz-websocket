# Pollz WebSocket Server (Go)

## Prerequisites
- Linux or macOS (preferred). On Windows, use **WSL** or dual boot.
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/)

> ⚠️ Note: Your system may support either `docker-compose` or `docker compose`.  
> Use the same consistently. Prepend `sudo` if you encounter permission errors.

---


## Forking and cloning

### Option 1: Clone All Repositories (recommended for full-stack setup)

1. **Fork** the repositories on GitHub (backend, frontend, websocket) from the original organization:  

   * Backend: [bitsacm/pollz-backend](https://github.com/bitsacm/pollz-backend)  
   * Frontend: [bitsacm/pollz-frontend](https://github.com/bitsacm/pollz-frontend)  
   * Websocket: [bitsacm/pollz-websocket](https://github.com/bitsacm/pollz-websocket)  

2. **Clone your forks** into a single `pollz` folder (replace `<your-github-username>` with yours):

   ```bash
   # Create a parent folder to keep all Pollz repos together
   mkdir pollz
   cd pollz

   # Clone backend
   git clone https://github.com/<your-github-username>/pollz-backend.git

   # Clone frontend
   git clone https://github.com/<your-github-username>/pollz-frontend.git

   # Clone websocket
   git clone https://github.com/<your-github-username>/pollz-websocket.git

3. **Add upstream remotes** to fetch updates from the official repos:

   ```bash
   cd pollz-backend
   git remote add upstream https://github.com/bitsacm/pollz-backend.git
   cd ..

   cd pollz-frontend
   git remote add upstream https://github.com/bitsacm/pollz-frontend.git
   cd ..

   cd pollz-websocket
   git remote add upstream https://github.com/bitsacm/pollz-websocket.git
   cd ..
   ```

---

### Option 2: Clone websocket Only

1. **Fork** the repository on GitHub.

2. **Clone your fork** (replace `<your-github-username>`):

   ```bash
   git clone https://github.com/<your-github-username>/pollz-websocket.git
   cd pollz-websocket
   ```

3. *(Optional but recommended for contributors)* Add the original repo as upstream:

   ```bash
   git remote add upstream https://github.com/bitsacm/pollz-websocket.git
   git fetch upstream
   ```

---
## Setup

1. **Install dependencies**
```bash
go mod download
```

2. **Setup environment variables**
```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. **Run with Docker (Recommended)**
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


