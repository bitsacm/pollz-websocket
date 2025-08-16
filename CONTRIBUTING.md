# Contributing to Pollz WebSocket Server

## Setup
1. Fork the repository
2. Clone your fork: `git clone <your-fork-url>`
3. Create `.env` from `.env.example`
4. Install Go 1.19+
5. Run: `docker-compose up --build` or `make run`

## Development Workflow
1. Create feature branch: `git checkout -b feature/your-feature`
2. Make changes
3. Run tests: `make test`
4. Format code: `make fmt`
5. Commit with clear message
6. Push and create PR (see PR Guidelines below)

## Code Standards
- Follow Go conventions and `gofmt`
- Write unit tests for new functions
- Use meaningful variable and function names
- Keep functions small and focused
- Handle errors properly
- Add comments for exported functions

## Before Submitting PR
- [ ] Tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
- [ ] No hardcoded credentials
- [ ] WebSocket connections work
- [ ] Go.mod updated if needed
- [ ] Docker setup works

## GitHub PR Guidelines

### PR Title Format
Use conventional commit format: `type(scope): description`

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements

**Examples:**
- `feat(websocket): add message rate limiting`
- `fix(connection): resolve memory leak in hub`
- `perf(database): optimize message queries`
- `refactor(handlers): improve error handling`

### PR Description
- Clearly describe what changes were made
- Include performance impact if applicable
- Reference related issues using `Fixes #123`
- List any breaking changes
- Mention if go.mod was updated

### Branch Naming
- `feature/feature-name` - New features
- `fix/bug-description` - Bug fixes
- `docs/documentation-update` - Documentation
- `perf/performance-improvement` - Performance optimizations
- `refactor/component-name` - Refactoring

### WebSocket Specific Guidelines
- Test with multiple concurrent connections
- Verify message delivery and ordering
- Check memory usage under load
- Ensure proper connection cleanup
- Test reconnection scenarios

## Project Structure
- `cmd/server/` - Application entry point
- `internal/config/` - Configuration management
- `internal/handlers/` - HTTP and WebSocket handlers
- `internal/hub/` - WebSocket connection management
- `internal/models/` - Data structures
- `internal/database/` - Database operations