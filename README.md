# Arch-Manager

[![Go Version](https://img.shields.io/badge/go-1.19+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE.md)
[![Docker](https://img.shields.io/badge/docker-supported-blue.svg)](https://hub.docker.com)
[![Status](https://img.shields.io/badge/status-production-brightgreen.svg)](https://github.com/ebadidev/arch-manager)

**Arch-Manager** is a powerful, enterprise-grade proxy management platform built with Go that orchestrates Xray-core nodes for scalable, high-performance proxy infrastructure. It provides centralized management, monitoring, and coordination for distributed proxy networks.

> ⚠️ **Development Status**: This project is currently in active development and is **not production ready**.

## ✨ Features

- **🎯 Centralized Management**: Unified control panel for managing multiple proxy nodes
- **👥 User Administration**: Comprehensive user management with profiles and access control
- **📊 Real-time Analytics**: Advanced insights and statistics dashboard
- **🔄 Node Coordination**: Seamless orchestration of distributed multicore Arch-Node instances
- **⚡ Xray Integration**: Full Xray-core support with automatic binary management and multicore optimization
- **🔒 License Management**: Built-in licensing system for commercial deployments
- **🌐 Modern Web UI**: Responsive, intuitive web interface
- **📈 Auto-scaling**: Dynamic scaling based on load and demand
- **🔐 Security First**: Enterprise-grade security with authentication and encryption
- **💾 Smart Backups**: Automated hourly backups with one-week retention
- **🐳 Docker Ready**: Production-ready containerization support
- **📝 Comprehensive Logging**: Structured logging with multiple output formats

## 🏗️ Architecture

Arch-Manager serves as the central orchestrator in the Arch Net ecosystem:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Admin Panel   │    │   API Client    │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │     Arch-Manager          │
                    │  ┌─────────────────────┐  │
                    │  │   HTTP Server       │  │
                    │  │   User Management   │  │
                    │  │   Node Coordinator  │  │
                    │  │   License Manager   │  │
                    │  │   Database Layer    │  │
                    │  └─────────────────────┘  │
                    └─────────────┬─────────────┘
                                  │
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
    ┌─────▼─────┐           ┌─────▼─────┐           ┌─────▼─────┐
    │Arch-Node-1│           │Arch-Node-2│           │Arch-Node-N│
    │(Multicore │           │(Multicore │           │(Multicore │
    │ Xray Core)│           │ Xray Core)│           │ Xray Core)│
    └───────────┘           └───────────┘           └───────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Operating System**: Debian 10+ or Ubuntu 18.04+
- **Architecture**: amd64 (x86_64)
- **Memory**: 2 GB RAM minimum (4 GB recommended)
- **CPU**: 2 cores minimum (4+ cores recommended)
- **Storage**: 10 GB free space minimum
- **Network**: Stable internet connection with public IP

### Installation

1. **System Dependencies**
   ```bash
   apt-get update && apt-get install -y \
     make wget jq curl vim git openssl cron unzip
   ```

2. **BBR TCP Optimization** (Optional but recommended)
   ```bash
   echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf
   echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf
   sysctl -p
   ```

3. **Install Arch-Manager**
   ```bash
   git clone https://github.com/ebadidev/arch-manager.git
   cd arch-manager
   make setup
   ```

4. **Setup Xray Binaries**
   ```bash
   make setup-xray
   ```

5. **Access Web Interface**
   ```bash
   # Default: http://your-server-ip:8080
   # Default credentials: admin / password
   ```

### Docker Deployment

1. **Using Docker Compose**
   ```bash
   # Pull and start
   docker compose up -d
   
   # View logs
   docker compose logs -f
   ```

2. **Custom Configuration**
   ```bash
   # Edit configuration
   vim configs/main.json
   
   # Restart with new config
   docker compose restart
   ```

## ⚙️ Configuration

### Main Configuration (`configs/main.json`)

```json
{
  "http_server": {
    "host": "0.0.0.0",
    "port": 8080
  },
  "http_client": {
    "timeout": 30000
  },
  "logger": {
    "level": "info",
    "format": "2006-01-02 15:04:05.000"
  },
  "xray": {
    "log_level": "warning"
  }
}
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ARCH_MANAGER_PORT` | HTTP server port | `8080` |
| `ARCH_MANAGER_HOST` | Server bind address | `0.0.0.0` |
| `ARCH_MANAGER_LOG_LEVEL` | Application log level | `info` |

## 🎮 Web Interface

### Admin Panel Features

Access at `http://your-server:8080` with default credentials:
- **Username**: `admin`
- **Password**: `password`

#### 📋 Dashboard Sections

| Section | Description | Key Features |
|---------|-------------|--------------|
| **👥 Users** | User management | Create, edit, delete users; view profiles |
| **🖥️ Nodes** | Node management | Add/remove nodes, monitor status |
| **📊 System** | System overview | Server stats, resource usage |
| **📈 Insights** | Analytics | Traffic analysis, usage reports |
| **⚙️ Settings** | Configuration | System settings, preferences |

## 🔧 Management

### Service Operations

```bash
# Check service status
systemctl status arch-manager

# Start/stop service
systemctl start arch-manager
systemctl stop arch-manager

# Restart service
systemctl restart arch-manager

# Enable auto-start
systemctl enable arch-manager
```

### Updates

```bash
# Automatic update (recommended)
make update

# Manual steps
git fetch --all
git reset --hard
git clean -fd
git pull
make setup
systemctl restart arch-manager
```

### Maintenance Commands

```bash
# Clean logs
make clean

# Fresh installation (⚠️ destroys data)
make fresh

# Restore from backup
make recover

# Schedule automatic reboots
make schedule-reboot
```

## 📊 Monitoring & Logs

### Real-time Monitoring

```bash
# Service logs
journalctl -f -u arch-manager

# Application logs
tail -f ./storage/logs/app-std.log
tail -f ./storage/logs/app-err.log

# Xray logs
tail -f ./storage/logs/xray-access.log
tail -f ./storage/logs/xray-error.log
```

### Log Files Location

```
storage/logs/
├── app-std.log         # Application standard output
├── app-err.log         # Application errors
├── xray-access.log     # Xray access logs
└── xray-error.log      # Xray error logs
```

## 💾 Backup & Recovery

### Automatic Backups

Arch-Manager creates hourly database backups:

```
storage/database/
├── app.json                    # Current database
├── backup-monday-00.json       # Monday 12:00 AM
├── backup-monday-01.json       # Monday 01:00 AM
├── ...                         # Every hour for 7 days
└── backup-sunday-23.json       # Sunday 11:00 PM
```

### Manual Recovery

```bash
# Restore latest backup
make recover

# Manual restoration
systemctl stop arch-manager
cp storage/database/backup-{day}-{hour}.json storage/database/app.json
systemctl start arch-manager
```

## 🔌 API Reference

### Authentication

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### Core Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/auth/*` | POST | Authentication operations |
| `/api/v1/users` | GET/POST | User management |
| `/api/v1/nodes` | GET/POST | Node operations |
| `/api/v1/stats` | GET | Statistics and metrics |
| `/api/v1/settings` | GET/POST | System configuration |
| `/api/v1/insights` | GET | Analytics and reports |

## 🛠️ Development

### Local Development

```bash
# Setup development environment
make local-setup

# Run locally
make local-run

# Build binary
make build

# Update dependencies
go mod tidy
```

### Project Structure

```
arch-manager/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   └── start.go           # Start command
├── configs/               # Configuration files
├── internal/              # Internal packages
│   ├── app/              # Application core
│   ├── config/           # Configuration management
│   ├── coordinator/      # Node coordination
│   ├── database/         # Data persistence
│   ├── http/             # HTTP server & API
│   ├── licensor/         # License management
│   └── utils/            # Utilities
├── scripts/              # Setup and maintenance
├── storage/              # Runtime data
├── third_party/          # External binaries
└── web/                  # Frontend assets
```

## 🔒 Security

- **🔐 Authentication**: Token-based API authentication
- **🛡️ Authorization**: Role-based access control
- **🔒 Encryption**: TLS encryption for all communications
- **🔑 License Protection**: Hardware-bound licensing
- **📝 Audit Logs**: Comprehensive security logging
- **🚫 Rate Limiting**: API rate limiting and DDoS protection

## 🤝 Integration

### Adding Arch-Nodes

1. **Deploy Arch-Node instance**
2. **Get node information**:
   ```bash
   # On the node server
   make info
   ```
3. **Register in Arch-Manager**:
   - Navigate to **Nodes** section in web interface
   - Click **Add Node**
   - Enter node details and credentials

### API Integration

```javascript
// Example: Fetch user statistics
const response = await fetch('/api/v1/stats/users', {
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});
const stats = await response.json();
```

## 📋 Troubleshooting

### Common Issues

1. **Service won't start**
   ```bash
   # Check logs for errors
   journalctl -u arch-manager --no-pager
   
   # Verify configuration
   ./arch-manager start --dry-run
   ```

2. **Web interface not accessible**
   ```bash
   # Check if service is running
   systemctl status arch-manager
   
   # Verify port availability
   netstat -tlnp | grep :8080
   
   # Check firewall
   ufw status
   ```

3. **Node connectivity issues**
   ```bash
   # Test node connection
   curl -I http://node-ip:8080/
   
   # Check node logs
   journalctl -f -u arch-node-1
   ```

### Performance Optimization

```bash
# Monitor resource usage
htop
iostat 1
iftop

# Optimize for high traffic
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf
sysctl -p
```

## 🤝 Contributing

We welcome contributions! Please follow these steps:

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature-amazing-feature`
3. **Commit** your changes: `git commit -m 'Add amazing feature'`
4. **Push** to the branch: `git push origin feature-amazing-feature`
5. **Open** a Pull Request

### Development Guidelines

- Follow Go conventions and best practices
- Add tests for new functionality
- Update documentation for API changes
- Use conventional commits for messages

## 📚 Related Projects

- **[Arch-Node](https://github.com/ebadidev/arch-node)** - High-performance multicore proxy nodes
- **[Xray-core](https://github.com/XTLS/Xray-core)** - Advanced proxy protocols

## 📄 License

This project is licensed under the terms specified in the [LICENSE](LICENSE.md) file.

---

**Made with ❤️ by the Arch Net team**
