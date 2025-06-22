# Contributing to Miniflux

This document outlines how to contribute effectively to Miniflux.

## Philosophy

Miniflux follows a **minimalist philosophy**. The feature set is intentionally kept limited to avoid bloatware. Before contributing, please understand that:

- **Improving existing features takes priority over adding new ones**
- **Quality over quantity** - well-implemented, focused features are preferred
- **Simplicity is key** - complex solutions are discouraged in favor of simple, maintainable code

## Before You Start

### Feature Requests

Before implementing a new feature:

- Check if it aligns with Miniflux's philosophy
- Consider if the feature could be implemented differently to maintain simplicity
- Remember that developing software takes significant time, and this is a volunteer-driven project
- If you need a specific feature, the best approach is to contribute it yourself

### Bug Reports

When reporting bugs:

- Search existing issues first to avoid duplicates
- Provide clear reproduction steps
- Include relevant system information (OS, browser, Miniflux version)
- Include error messages, screenshots, and logs when applicable

## Development Setup

### Requirements

- **Git**
- **Go >= 1.24**
- **PostgreSQL**

### Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork locally:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/miniflux.git
   cd miniflux
   ```

3. **Build the application binary:**
   ```bash
   make miniflux
   ```

4. **Run locally in debug mode:**
   ```bash
   make run
   ```

### Database Setup

For development and testing, you can run a local PostgreSQL database with Docker:

```bash
# Start PostgreSQL container
docker run --rm --name miniflux2-db -p 5432:5432 \
  -e POSTGRES_DB=miniflux2 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  postgres
```

You can also use an existing PostgreSQL instance. Make sure to set the `DATABASE_URL` environment variable accordingly.

## Development Workflow

### Code Quality

1. **Run the linter:**
   ```bash
   make lint
   ```
   Requires `staticcheck` and `golangci-lint` to be installed.

2. **Run unit tests:**
   ```bash
   make test
   ```

3. **Run integration tests:**
   ```bash
   make integration-test
   make clean-integration-test
   ```

### Building

- **Current platform:** `make miniflux`
- **All platforms:** `make build`
- **Specific platforms:** `make linux-amd64`, `make darwin-arm64`, etc.
- **Docker image:** `make docker-image`

### Cross-Platform Support

Miniflux supports multiple architectures. When making changes, ensure compatibility across:
- Linux (amd64, arm64, armv7, armv6, armv5)
- macOS (amd64, arm64)
- FreeBSD, OpenBSD, Windows (amd64)

## Pull Request Guidelines

### What Is Preferred

✅ **Good Pull Requests:**

- Focus on a single issue or feature
- Include tests for new functionality
- Maintain or improve performance
- Follow existing code style and patterns
- The commit messages follow the [conventional commit format](https://www.conventionalcommits.org/) (e.g., `feat: add new feature`, `fix: resolve bug`)
- Update documentation when necessary

### What to Avoid

❌ **Pull Requests That Cannot Be Accepted:**

- **Too many changes** - makes review difficult
- **Breaking changes** - disrupts existing functionality
- **New bugs or regressions** - reduces software quality
- **Unnecessary dependencies** - conflicts with minimalist approach
- **Performance degradation** - slows down the software
- **Poor-quality code** - hard to maintain
- **Dependent PRs** - creates review complexity
- **Radical UI changes** - disrupts user experience
- **Conflicts with philosophy** - doesn't align with minimalist approach

### Pull Request Template

When creating a pull request, please include:

- **Description:** What does this PR do?
- **Motivation:** Why is this change needed?
- **Testing:** How was this tested?
- **Breaking Changes:** Are there any breaking changes?
- **Related Issues:** Link to any related issues

## Code Style

- Follow Go conventions and best practices
- Use `gofmt` to format your Go code, and `jshint` for JavaScript
- Write clear, descriptive variable and function names
- Include comments for complex logic
- Keep functions small and focused

## Testing

### Unit Tests
- Write unit tests for new functions and methods
- Ensure tests are fast and don't require external dependencies
- Aim for good test coverage

### Integration Tests
- Add integration tests for new API endpoints
- Tests run against a real PostgreSQL database
- Ensure tests clean up after themselves

## Communication

- **Discussions:** Use GitHub Discussions for general questions and community interaction
- **Issues:** Use GitHub issues for bug reports and feature requests
- **Pull Requests:** Use PR comments for code-specific discussions
- **Philosophy Questions:** Refer to the FAQ for common questions about project direction

## Questions?

- Check the [FAQ](https://miniflux.app/faq.html) for common questions
- Review the [development documentation](https://miniflux.app/docs/development.html) and [internationalization guide](https://miniflux.app/docs/i18n.html)
- Look at existing issues and pull requests for examples
