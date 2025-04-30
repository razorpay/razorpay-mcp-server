# Contributing to Razorpay MCP Server

Thank you for your interest in contributing to the Razorpay MCP Server! This document outlines the process for contributing to this project.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project. We strive to maintain a welcoming and inclusive environment for all contributors.

## TLDR;

```
make test
make fmt
make lint
make build
make run
```

We use Cursor to contribute - our AI developer. Look at `~/.cusor/rules`. It understands the standards we have defined and codes with that.

## Development Process

We use a fork-based workflow for all contributions:

1. **Fork the repository**: Start by forking the [razorpay-mcp-server repository](https://github.com/razorpay/razorpay-mcp-server) to your GitHub account.

2. **Clone your fork**: Clone your fork to your local machine:
   ```bash
   git clone https://github.com/YOUR-USERNAME/razorpay-mcp-server.git
   cd razorpay-mcp-server
   ```

3. **Add upstream remote**: Add the original repository as an "upstream" remote:
   ```bash
   git remote add upstream https://github.com/razorpay/razorpay-mcp-server.git
   ```

4. **Create a branch**: Create a new branch for your changes:
   ```bash
   git checkout -b username/feature
   ```
   Use a descriptive branch name that includes your username followed by a brief feature description.

5. **Make your changes**: Implement your changes, following the code style guidelines.

6. **Write tests**: Add tests for your changes when applicable.

7. **Run tests and linting**: Make sure all tests pass and the code meets our linting standards.

8. **Commit your changes**: Make commits with clear messages following this format:
   ```bash
   git commit -m "[type]: description of the change"
   ```
   Where `type` is one of:
   - `chore`: for tasks like adding linter config, GitHub Actions, addressing PR review comments, etc.
   - `feat`: for adding new features like a new fetch_payment tool
   - `fix`: for bug fixes
   - `ref`: for code refactoring
   - `test`: for adding UTs or E2Es
   
   Example: `git commit -m "feat: add payment verification tool"`

9. **Keep your branch updated**: Regularly sync your branch with the upstream repository:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

10. **Push to your fork**: Push your changes to your fork:
    ```bash
    git push origin username/feature
    ```

11. **Create a Pull Request**: Open a pull request from your fork to the main repository.

## Pull Request Process

1. Fill out the pull request template with all relevant information.
2. Link any related issues in the pull request description.
3. Ensure all status checks pass.
4. Wait for review from maintainers.
5. Address any feedback from the code review.
6. Once approved, a maintainer will merge your changes.

## Local Development Setup

### Prerequisites

- Go 1.21 or later
- Docker (for containerized development)
- Git

### Setting up the Development Environment

1. Clone your fork of the repository (see above).

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up your environment variables:
   ```bash
   export RAZORPAY_KEY_ID=your_key_id
   export RAZORPAY_KEY_SECRET=your_key_secret
   ```

### Running the Server Locally

There are `make` commands also available now for the below, refer TLDR; above.

To run the server in development mode:

```bash
go run ./cmd/razorpay-mcp-server/main.go stdio
```

### Running Tests

To run all tests:

```bash
go test ./...
```

To run tests with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Quality and Linting

We use golangci-lint for code quality checks. To run the linter:

```bash
# Install golangci-lint if you don't have it
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run the linter
golangci-lint run
```

Our linting configuration is defined in `.golangci.yaml` and includes:
- Code style checks (gofmt, goimports)
- Static analysis (gosimple, govet, staticcheck)
- Security checks (gosec)
- Complexity checks (gocyclo)
- And more

Please ensure your code passes all linting checks before submitting a pull request.

## Documentation

When adding new features or modifying existing ones, please update the documentation accordingly. This includes:

- Code comments
- README updates
- Tool documentation

## Adding New Tools

When adding a new tool to the Razorpay MCP Server:

1. Review the detailed developer guide at [pkg/razorpay/README.md](pkg/razorpay/README.md) for complete instructions and examples.
2. Create a new function in the appropriate resource file under `pkg/razorpay` (or create a new file if needed).
3. Implement the tool following the patterns in the developer guide.
4. Register the tool in `server.go`.
5. Add appropriate tests.
6. Update the main README.md to document the new tool.

The developer guide for tools includes:
- Tool structure and patterns
- Parameter definition and validation
- Examples for both GET and POST endpoints
- Best practices for naming and organization

## Releasing

Releases are managed by the maintainers. We use [GoReleaser](https://goreleaser.com/) for creating releases.

## Getting Help

If you have questions or need help with the contribution process, please use [GitHub Discussions](https://github.com/razorpay/razorpay-mcp-server/discussions) to ask for assistance.

Thank you for contributing to the Razorpay MCP Server! 