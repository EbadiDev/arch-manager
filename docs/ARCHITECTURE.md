# Arch-Manager Architecture Documentation

## Overview

Arch-Manager is a Go-based proxy management platform that serves as the central orchestrator for distributed Xray-core nodes. This document provides a detailed breakdown of how the system works internally.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                       Arch-Manager                             │
│                                                                 │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   HTTP Server   │  │   Coordinator   │  │    Database     │ │
│  │                 │  │                 │  │                 │ │
│  │ - Web UI        │  │ - Node Sync     │  │ - JSON File     │ │
│  │ - API Endpoints │  │ - Config Push   │  │ - Users/Nodes   │ │
│  │ - Auth Layer    │  │ - Stats Collect │  │ - Settings      │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│           │                     │                     │         │
│           │                     │                     │         │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │     Writer      │  │     Xray        │  │    Licensor     │ │
│  │                 │  │                 │  │                 │ │
│  │ - Config Gen    │  │ - Local Core    │  │ - License Check │ │
│  │ - Local/Remote  │  │ - Stats API     │  │ - Validation    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                 │
                     ┌───────────┴───────────┐
                     │                       │
              ┌─────────────┐         ┌─────────────┐
              │ Arch-Node-1 │   ...   │ Arch-Node-N │
              │             │         │             │
              │ - Xray Core │         │ - Xray Core │
              │ - HTTP API  │         │ - HTTP API  │
              └─────────────┘         └─────────────┘
```

## Core Components

### 1. Application (internal/app/app.go)

The main application struct that initializes and manages all components:

```go
type App struct {
    Context     context.Context
    Cancel      context.CancelFunc
    Config      *config.Config
    Logger      *logger.Logger
    HttpClient  *client.Client
    HttpServer  *server.Server
    Database    *database.Database
    Writer      *writer.Writer
    Coordinator *coordinator.Coordinator
    Xray        *xray.Xray
    Enigma      *enigma.Enigma
    Licensor    *licensor.Licensor
}
```

**Initialization Flow:**
1. Load configuration from JSON files
2. Initialize logger with structured logging
3. Create database connection (JSON file-based)
4. Setup Xray-core instance
5. Initialize HTTP client for node communication
6. Create coordinator for node management
7. Start HTTP server for web interface and API

### 2. Database Layer (internal/database/)

The system uses a **JSON file-based database** stored at `storage/database/app.json`.

**Database Structure:**
```go
type Content struct {
    Settings *Settings  // System configuration
    Stats    *Stats     // Usage statistics  
    Users    []*User    // User accounts
    Nodes    []*Node    // Connected nodes
}
```

**Key Features:**
- **Thread-safe**: Uses mutex for concurrent access
- **Auto-backup**: Hourly backups with 7-day retention
- **Atomic writes**: Ensures data consistency
- **Validation**: Struct validation on load/save

**User Management:**
```go
type User struct {
    Id                  int     // Unique identifier
    Identity            string  // UUID for client configs
    Name                string  // Display name
    Quota               float64 // Data limit in GB
    Usage               float64 // Current usage in GB
    UsageBytes          int64   // Raw bytes consumed
    UsageResetAt        int64   // Monthly reset timestamp
    Enabled             bool    // Active status
    ShadowsocksPassword string  // Unique password
    ShadowsocksMethod   string  // Encryption method
    CreatedAt           int64   // Creation timestamp
}
```

**Node Management:**
```go
type Node struct {
    Id         int        // Unique identifier
    Host       string     // IP/hostname
    HttpToken  string     // Authentication token
    HttpPort   int        // API port
    Usage      float64    // Consumed bandwidth (GB)
    UsageBytes int64      // Raw bytes
    PushStatus NodeStatus // Config sync status
    PullStatus NodeStatus // Health check status
    PushedAt   int64      // Last config push
    PulledAt   int64      // Last health check
}
```

### 3. Configuration Management (internal/config/)

**Configuration Sources (Priority Order):**
1. `configs/main.example.json` (local override)
2. `configs/main.defaults.json` (defaults)

**Environment Variables:**
- Config paths are managed through `internal/config/env.go`
- Default paths: `./configs/`, `./storage/`, `./third_party/`

**Configuration Structure:**
```go
type Config struct {
    HttpServer struct {
        Host string // Bind address (0.0.0.0)
        Port int    // Listen port (8080)
    }
    HttpClient struct {
        Timeout int // Request timeout (30s)
    }
    Logger struct {
        Level  string // debug/info/warn/error
        Format string // Timestamp format
    }
    Xray struct {
        LogLevel string // Xray log level
    }
}
```

### 4. Coordinator (internal/coordinator/)

The coordinator is the **heart of the system**, managing all node operations through multiple background workers:

**Worker Processes:**
1. **Config Sync Worker** (10s interval): Pushes configurations to outdated nodes
2. **Status Pull Worker** (1m interval): Checks node health and marks unavailable nodes
3. **Local Stats Worker** (1m interval): Collects stats from local Xray instance
4. **Remote Stats Worker** (1m interval): Collects stats from remote nodes
5. **Backup Worker** (1h interval): Creates database backups
6. **Usage Reset Worker** (1h interval): Resets monthly user quotas

**Node Communication:**
```go
// Push configuration to node
func (c *Coordinator) syncRemoteConfig(node *database.Node) {
    url := fmt.Sprintf("http://%s:%d/v1/configs", node.Host, node.HttpPort)
    xc := c.writer.RemoteConfig(node, c.state.XrayUpdatedAt(), c.state.XraySharedPassword())
    
    _, err := c.hc.Do(http.MethodPost, url, node.HttpToken, xc)
    if err == nil {
        node.PushStatus = database.NodeStatusAvailable
    } else {
        node.PushStatus = database.NodeStatusUnavailable
    }
}
```

**Stats Collection:**
```go
// Collect stats from local Xray
func (c *Coordinator) syncLocalStats() error {
    queryStats, err := c.xray.QueryStats()
    
    // Parse user traffic: "user>>>123>>>traffic>>>downlink"
    // Parse node traffic: "inbound>>>internal-1>>>traffic>>>downlink"
    
    // Update database with collected stats
    // Disable users who exceed quota
    // Trigger config resync if needed
}
```

### 5. Configuration Writer (internal/writer/)

The writer generates Xray configurations for both local and remote instances:

**Local Configuration:**
```go
func (w *Writer) LocalConfig() (*xray.Config, error) {
    clients := w.clients() // Active users
    
    xc := xray.NewConfig(w.c.Xray.LogLevel)
    
    // Create inbounds for different connection types:
    // - Relay: Routes traffic through nodes
    // - Reverse: Incoming connections from nodes  
    // - Direct: Direct connections (no relay)
    
    // Setup routing rules and load balancers
    // Configure reverse proxy portals for nodes
    
    return xc, nil
}
```

**Remote Configuration:**
```go
func (w *Writer) RemoteConfig(node *database.Node, lastUpdate time.Time, password string) *xray.Config {
    xc := xray.NewConfig(w.c.Xray.LogLevel)
    
    xc.Metadata = &xray.Metadata{
        UpdatedAt: lastUpdate.Format(time.RFC3339),
        UpdatedBy: w.database.Content.Settings.Host,
    }
    
    // Configure relay inbound (receives from manager)
    // Configure reverse bridge (connects back to manager)
    // Configure remote inbound (direct user connections)
    
    return xc
}
```

### 6. HTTP Server (internal/http/server/)

**Route Structure:**
```
/                          -> Static web files
/profile                   -> User profile page

/v1/sign-in               -> Authentication
/v1/profile               -> User profile API

Admin Routes (requires auth):
/v1/users                 -> User management
/v1/nodes                 -> Node management  
/v1/stats                 -> Statistics
/v1/settings              -> System settings
/v1/information           -> License info
/v1/imports               -> Data import
```

**Authentication Flow:**
1. Client sends credentials to `/v1/sign-in`
2. Server validates against admin password
3. Returns encrypted token (using Enigma)
4. Client includes token in Authorization header
5. Middleware validates token for protected routes

### 7. Node Connection Architecture

**Connection Types:**

1. **Relay Mode** (`SsRelayPort`):
   - Manager creates outbound connections to nodes
   - Nodes accept inbound connections from manager
   - Traffic flows: Client → Manager → Node → Internet

2. **Reverse Mode** (`SsReversePort`):
   - Nodes establish persistent connections to manager
   - Manager creates reverse proxy portals
   - Traffic flows: Client → Manager → [Reverse Tunnel] → Node → Internet

3. **Direct Mode** (`SsDirectPort`):
   - Manager provides direct access (no relay)
   - Traffic flows: Client → Manager → Internet

4. **Remote Mode** (`SsRemotePort`):
   - Nodes provide direct user access
   - Traffic flows: Client → Node → Internet

**Configuration Synchronization:**
1. Manager generates node-specific Xray configs
2. Configs pushed via HTTP API to nodes
3. Nodes restart Xray with new configuration
4. Status tracking via push/pull status fields

## Data Flow

### User Traffic Flow
```
1. User connects with Shadowsocks client
2. Traffic hits Manager's inbound (relay/reverse/direct)
3. Manager routes based on configuration:
   - Relay: Forward to specific node
   - Reverse: Use reverse tunnel to node
   - Direct: Process locally
4. Node processes traffic and sends to internet
5. Response follows reverse path
6. Stats collected at each hop
```

### Configuration Update Flow
```
1. Admin changes settings via Web UI
2. HTTP handler updates database
3. Coordinator detects change
4. Writer generates new configurations
5. Local Xray restarted with new config
6. Remote configs pushed to all nodes
7. Nodes restart with updated configs
8. Status monitoring confirms success
```

### Statistics Collection Flow
```
1. Xray cores report stats via API
2. Local stats collected every minute
3. Remote stats pulled from nodes
4. Data aggregated and stored in database
5. User quotas checked and enforced
6. Overuse triggers config updates
7. Web dashboard displays real-time data
```

## Security Architecture

**Authentication:**
- Admin password stored in database settings
- API tokens encrypted using Enigma (Ed25519)
- Session-based authentication for web interface

**Node Security:**
- Each node has unique HTTP token
- Token-based API authentication
- Optional proxy routing for communication

**Traffic Encryption:**
- Shadowsocks encryption for user traffic
- Shadowsocks 2022 for inter-node communication
- TLS for HTTP API communication

## High Availability Features

**Automatic Failover:**
- Multiple nodes in load balancer
- Unhealthy nodes automatically removed
- Traffic redistributed to healthy nodes

**Data Persistence:**
- Hourly automated backups
- 7-day retention policy
- Atomic database operations

**Monitoring & Recovery:**
- Health checks every minute
- Automatic service restart capabilities
- Comprehensive logging and alerting

## Performance Optimizations

**Connection Pooling:**
- HTTP client reuse for node communication
- Persistent connections where possible

**Efficient Routing:**
- Load balancing across multiple nodes
- Intelligent path selection
- Traffic optimization algorithms

**Resource Management:**
- Background worker coordination
- Memory-efficient data structures
- Optimized JSON parsing/serialization

This architecture provides a robust, scalable foundation for managing distributed proxy infrastructure while maintaining high performance and reliability.
