# 🔍 OtelCompare

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

OtelCompare is a command-line tool designed to automate the process of analyzing and commenting on traces in GitHub Pull Requests. This tool can operate in two modes:

1. **Compare Mode** 🔄: Compares traces between different files
2. **Info Mode** ℹ️: Provides detailed information about traces in a single file

## 🚀 Features

- Direct GitHub Pull Requests integration
- Automatic OpenTelemetry trace analysis
- Detailed comment generation
- Compare mode for change analysis
- Info mode for trace documentation
- Dry-run mode to preview comments

## 📋 Prerequisites

- Go 1.23 or higher
- GitHub Personal Access Token with PR commenting permissions

## 🛠️ Installation

```bash
go install github.com/lpcalisi/otelcompare@latest
```

## 💻 Usage

### Compare Mode

```bash
otelcompare compare -i examples/modified.json -i examples/baseline.json [--pr <pr-number> --owner <owner> --repo <repo>] [--attribute <attr>]
```

The compare command requires at least two input files to compare. You can specify the attribute to use for trace identification with `--attribute` (default: "trace_id").

### Info Mode

```bash
otelcompare info -i examples/baseline.json [--pr <pr-number> --owner <owner> --repo <repo>]
```

The info command analyzes a single trace file and generates a detailed report. The GitHub-specific flags (`--pr`, `--owner`, and `--repo`) are only required when posting to GitHub.

### Dry Run Mode

Both commands support a `--dry-run` flag that will print the comment to stdout without posting it to GitHub:

```bash
otelcompare compare -i examples/modified.json -i examples/baseline.json --dry-run
otelcompare info -i examples/baseline.json --dry-run
```

When using `--dry-run`, the GitHub-specific flags (`--pr`, `--owner`, and `--repo`) are not required.

## ⚙️ Configuration

The tool requires a GitHub token to be set in environment variables:

```bash
export GITHUB_TOKEN=your-token-here
```

## 🤝 Contributing

Contributions are welcome. Please open an issue first to discuss the changes you would like to make.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 