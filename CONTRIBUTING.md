# Contributing to GoTune

First off, thank you for considering contributing to GoTune! Your help is greatly appreciated.

This document provides guidelines for contributing to the project. Please read it carefully to ensure a smooth and effective contribution process.

## How to Contribute

There are many ways to contribute to GoTune, including:

- **Reporting Bugs:** If you find a bug, please open an issue and provide as much detail as possible.
- **Suggesting Enhancements:** If you have an idea for a new feature or an improvement to an existing one, open an issue to discuss it.
- **Writing Code:** If you want to fix a bug or implement a new feature, you can submit a pull request.
- **Improving Documentation:** If you find any errors or omissions in the documentation, please submit a pull request with your changes.

## Getting Started

Before you start working on a contribution, please make sure you have read the following documents:

- **[BUILD.md](docs/BUILD.md):** This document explains how to build and run the project.
- **[DEVELOPMENT.md](docs/DEVELOPMENT.md):** This document provides a guide for developers, covering the development workflow, testing, code quality, and architecture.
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md):** This document provides an in-depth explanation of the project's architecture.

## Pull Request Process

1.  **Fork the repository** and create your branch from `main`.
2.  **Make your changes.** Please follow the coding style and conventions used in the project.
3.  **Ensure the tests pass.** Run `make test` to run all tests.
4.  **Ensure the linter passes.** Run `make lint` to check for linting errors.
5.  **Submit a pull request.** Provide a clear and descriptive title and a detailed description of your changes.

## Coding Style

- **Go:** Follow the standard Go conventions. Use `go fmt` to format your code.
- **Fyne:** Follow the Fyne conventions for UI code.
- **Logging:** Use the structured logger (`slog`) for all logging. See `DEVELOPMENT.md` for more details.

## Questions?

If you have any questions, feel free to open an issue or reach out to the maintainers.
