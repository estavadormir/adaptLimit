# Contributing to AdaptLimit

Thank you for considering contributing to AdaptLimit! This document outlines the process for contributing to this project.

## Development Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Pull Request Process

1. Ensure your code passes all tests
2. Update the README.md with details of changes if needed
3. The PR should work in Go 1.22 and above
4. Your PR will be merged once it is reviewed and approved

## Code Style

Please follow standard Go style guidelines:
- Run `go fmt` before committing
- Make sure your code passes `golint` and `go vet`
- Follow idiomatic Go practices

## Testing

- Add tests for new features
- Maintain or improve test coverage
- Run `go test -cover ./...` to check coverage

## License

By contributing, you agree that your contributions will be licensed under the project's MIT License.
