# Contributing to the PingOne MCP Server

We appreciate your help! We welcome contributions in the form of creating issues or pull requests.

## Before You Start

Know that:

1. If you have any questions, please ask! We'll help as best we can.
2. While we appreciate perfect PRs, it's not essential. We'll fix up any housekeeping changes before merge. Any PRs that need further work, we'll point you in the right direction or can take on ourselves.
3. We may not be able to respond quickly; our development cycles are on a priority basis.
4. We base our priorities on customer need and the number of votes on issues/PRs by the number of 👍 reactions. If there is an existing issue or PR for something you'd like, please vote!

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) 1.25+ (to build and test the server)
- Access to a PingOne environment for testing

### Development Setup

1. **Clone the repository:**

```bash
git clone https://github.com/pingidentity/pingone-mcp-server.git
cd pingone-mcp-server
```

2. **Install dependencies:**

```bash
go mod download
```

3. **Build the server:**

```bash
make build
```

4. **Run tests:**

```bash
make test
```

## Development Workflow

### Creating a New Feature or Bug Fix

1. **Create a new branch** using the naming convention is advised but not essential:
   ```
   <type>/<description>
   ```
   
   Types:
   - `feature/` - For new features or enhancements
   - `bugfix/` - For bug fixes
   - `hotfix/` - For urgent production fixes
   - `chore/` - For maintenance tasks, dependencies, tooling
   - `docs/` - For documentation changes
   - `refactor/` - For code refactoring without changing functionality
   - `test/` - For adding or updating tests

   Example:
   ```bash
   git checkout -b feature/add-user-management
   ```

2. **Make your changes** following our code standards:
   - Follow Go best practices and idioms
   - Write clear, descriptive commit messages
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes:**

```bash
# Run tests
make test

# Run linting
make lint

# Build to verify
make build
```

4. **Commit your changes:**

```bash
git add .
git commit -m "feat: add user management functionality

- Implemented user CRUD operations
- Added integration tests
- Updated documentation"
```

5. **Push your branch:**

```bash
git push origin feature/add-user-management
```

1. **Open a Pull Request** on GitHub following the pull request template, providing:
   - A clear title describing the change
   - A detailed description of what was changed and why
   - Any relevant issue numbers
   - Test results or screenshots if applicable

## Code Standards

### Go Code Style

- Follow standard Go formatting (use `gofmt`)
- Use meaningful variable and function names
- Keep functions focused and concise
- Write comprehensive godoc comments for exported functions and types
- Handle errors properly and provide context

### Copyright Notice

All Go files must include the copyright notice at the top:

```go
// Copyright © 2025 Ping Identity Corporation
```

### Testing

- Write unit tests for all new functionality
- Ensure all tests pass before submitting a PR
- Aim for good test coverage of critical paths
- Use table-driven tests where appropriate

### Documentation

- Update README.md if adding new features or changing usage
- Add godoc comments for all exported functions, types, and packages
- Include examples where helpful
- Keep documentation clear and concise

## Pull Request Process

1. **Before Submitting:**
   - Ensure all tests pass (`make test`)
   - Run linting (`make lint`)
   - Update documentation as needed
   - Rebase on the latest main branch

2. **PR Description Should Include:**
   - What was changed and why
   - How to test the changes
   - Any breaking changes
   - Related issue numbers

3. **Review Process:**
   - A maintainer will review your PR
   - Address any feedback or requested changes
   - Once approved, a maintainer will merge your PR

4. **After Merge:**
   - Your contribution will be included in the next release
   - Delete your feature branch

## Reporting Issues

When reporting issues, please include:

- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Go version and OS
- Relevant logs or error messages
- Any other context that might be helpful

## Code of Conduct

Please be respectful and constructive in all interactions. We're all here to build something great together.

## Questions?

If you have questions about contributing, feel free to:

- Open an issue with the `question` label
- Reach out to the maintainers

Thank you for contributing to the PingOne MCP Server!
