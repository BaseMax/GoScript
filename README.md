# GoScript

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

GoScript is a lightweight scripting language implemented in Go. It provides a simple, expressive syntax for common scripting tasks and serves as an excellent platform for learning language design and building domain-specific languages.

## Features

- **Interpreted Scripting:** Write and execute scripts without a separate compilation step.
- **Simple Syntax:** Clean and intuitive syntax for arithmetic, logic, conditionals, loops, functions, and more.
- **Modular Design:** Easily extend or integrate new features thanks to the well-organized code structure.
- **Cross-Platform:** Built in Go, GoScript can run on any platform that supports Go.

## Syntax

**Hello World:**

```
print("Hey World!")
```

**Fibonacci:**

```
fn f(n) {
    if n <= 1 { 1 }
    n * f(n-1)
}

println(f(5))
```

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) Tested on 1.22.4

### Installation

Clone the repository:

```bash
git clone https://github.com/BaseMax/GoScript.git
cd GoScript
```

Build the project:

```bash
go build -o goscript
or
go build -o goscript.exe
```

Alternatively, you can run it directly with:

```bash
go run goscript.go
```

### Running a Script

After building the executable, you can run a GoScript file by providing its path as an argument:

```bash
./goscript.exe path/to/your/script.gos
```

For example, to run one of the provided examples:

```bash
./goscript.exe examples/hello.gos
```

## Project Structure

```bash
GoScript/
├── .gitignore           # Git ignore file
├── evaluator.go         # Evaluator: Executes the AST nodes
├── examples/            # Example GoScript programs
├── go.mod               # Go module file
├── goscript.exe         # Compiled executable (Windows)
├── lexer.go             # Lexer: Tokenizes the source code
├── LICENSE              # MIT License file
├── goscript.go          # Entry point of the interpreter
└── parser.go            # Parser: Builds the AST from tokens
```

## Contributing

Contributions to GoScript are welcome! If you have ideas for improvements, new features, or bug fixes, please follow these steps:

- Fork the repository.
- Create a new branch for your feature or bugfix.
- Commit your changes with clear messages.
- Open a pull request describing your changes.
- Feel free to open issues for any bugs or feature requests.

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Copyright

© 2025 Max Base (Seyyed Ali Mohammadiyeh)