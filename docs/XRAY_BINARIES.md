# Xray Binary Management

This project uses git submodules and automated scripts to manage Xray binaries instead of committing them directly to the repository.

## Setup

### Quick Setup (Linux x64 only)
```bash
make setup-xray
```

### Full Setup (All platforms)
```bash
make setup-xray-all
```

### Manual Setup
```bash
# Download latest release
./scripts/setup-xray.sh latest

# Download specific version
./scripts/setup-xray.sh v1.8.4

# Download all platforms
./scripts/setup-xray.sh latest all
```

## Directory Structure

After setup, you'll have:
```
third_party/
├── xray-linux-64/
│   ├── xray          # Linux x64 binary
│   └── LICENSE
└── xray-macos-arm64/ # (optional)
    ├── xray          # macOS ARM64 binary
    └── LICENSE
```

## Git Submodule

The `.gitmodules` file references the official Xray-core repository for tracking releases:

```
[submodule "third_party/xray-releases"]
    path = third_party/xray-releases
    url = https://github.com/XTLS/Xray-core.git
    branch = main
```

## Updating Binaries

To update to the latest Xray release:
```bash
make setup-xray
```

## Note

The actual binary files are not committed to this repository. They are downloaded from official releases and should be ignored by git (see `.gitignore`).
