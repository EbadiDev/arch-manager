# Xray Binary Management

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

## Updating Binaries

To update to the latest Xray release:
```bash
make setup-xray
```
