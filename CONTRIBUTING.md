# Contributing to MinifluxNg

MinifluxNg is a fork of [Miniflux](https://github.com/miniflux/v2). Contributions that belong to upstream (bug fixes, core improvements) should go to the upstream project first.

Fork-specific contributions (AI summary, web scraper, Pinchtab JS rendering) are welcome here.

## Development Setup

### Requirements

- **Git**
- **Go >= 1.24**
- **PostgreSQL**

### Getting Started

1. **Clone the repository:**
   ```bash
   git clone https://github.com/naiba-forks/miniflux.git
   cd miniflux
   ```

2. **Build the application binary:**
   ```bash
   make miniflux
   ```

3. **Run locally in debug mode:**
   ```bash
   make run
   ```

### Database Setup

```bash
docker run --rm --name miniflux2-db -p 5432:5432 \
  -e POSTGRES_DB=miniflux2 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  postgres
```

## Code Quality

```bash
make lint    # requires staticcheck and golangci-lint
make test
make integration-test
make clean-integration-test
```

## Pull Request Guidelines

- Focus on a single issue or feature
- Include tests for new functionality
- Follow existing code style and patterns
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
- No unnecessary dependencies
- No breaking changes without discussion
