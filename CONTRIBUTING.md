# Contributing to Repoan-CLI 🤝

Thank you for your interest in improving Repoan! All help is welcome – from bug reports to new features.

## 🛠️ Development Setup

1. **Install Go:** Version 1.22 or higher is recommended.
2. **Clone the repository:**
   ```bash
   git clone https://github.com/repoan/repoan.git
   cd repoan
   ```
3. **Install dependencies:**
   ```bash
   go mod tidy
   ```
4. **Build:**
   ```bash
   go build -o repoan .
   ```

## 🧪 Running Tests

We value stability. Please ensure all tests pass:
```bash
go test -v ./...
```

## 📝 Pull Request Process

1. Create a **feature branch** (`git checkout -b feature/my-cool-feature`).
2. Implement your changes and add **unit tests**.
3. Document new features in `README.md` or in `docs/`.
4. Submit a **Pull Request** against the `main` branch.

## 🐛 Reporting Issues

If you find a bug or have a suggestion for improvement, please create an [issue](https://github.com/repoan/repoan/issues) with:
- A clear description.
- Steps to reproduce (for bugs).
- Expected vs. actual behavior.

---

Thank you for being part of the Repoan community!
