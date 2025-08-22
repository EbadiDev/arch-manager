# Node Connection and Communication

## Overview

This document explains how Arch-Manager connects to and communicates with Arch-Node instances in the distributed proxy network. The system supports multiple connection modes to handle different network topologies and use cases.

## Connection Architecture

### System Components

```
┌─────────────────┐     HTTP API     ┌─────────────────┐
│   Arch-Manager  │ ←────────────→  │   Arch-Node     │
│                 │                  │                 │
│ ┌─────────────┐ │                  │ ┌─────────────┐ │
│ │ Coordinator │ │   Config Push    │ │ HTTP Server │ │
│ │   Worker    │ │ ──────────────→  │ │   /v1/*     │ │
│ └─────────────┘ │                  │ └─────────────┘ │
│ ┌─────────────┐ │   Stats Pull     │ ┌─────────────┐ │
│ │   Writer    │ │ ←──────────────  │ │ Xray Core   │ │
│ │  Generator  │ │                  │ │   Stats     │ │
│ └─────────────┘ │                  │ └─────────────┘ │
└─────────────────┘                  └─────────────────┘
```

## Node Registration Process

### 1. Manual Node Addition

**Via Web Interface:**
1. Navigate to **Nodes** section in admin panel
2. Click **Add Node** button
3. Enter node details:
   ```
   Host: 192.168.1.100
   HTTP Port: 8080
   HTTP Token: secure-random-token
   ```
4. System validates connection and saves to database

**Via API:**
```bash
curl -X POST http://manager:8080/v1/nodes \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "host": "192.168.1.100",
    "http_port": 8080,
    "http_token": "secure-random-token"
  }'
```

### 2. Node Database Entry

When a node is added, it creates a database record:

```go
type Node struct {
    Id         int        `json:"id"`          // Auto-generated: 1, 2, 3...
    Host       string     `json:"host"`        // "192.168.1.100"
    HttpToken  string     `json:"http_token"`  // "secure-random-token"
    HttpPort   int        `json:"http_port"`   // 8080
    Usage      float64    `json:"usage"`       // 0.0 (GB)
    UsageBytes int64      `json:"usage_bytes"` // 0
    PushStatus NodeStatus `json:"push_status"` // "" (processing)
    PullStatus NodeStatus `json:"pull_status"` // "" (processing)
    PushedAt   int64      `json:"pushed_at"`   // 0
    PulledAt   int64      `json:"pulled_at"`   // 0
}
```

## Communication Protocols

### 1. Configuration Push (Manager → Node)

**API Endpoint:** `POST http://{node_host}:{node_port}/v1/configs`

**Authentication:** 
```
Authorization: Bearer {node.HttpToken}
```

**Request Body:** Complete Xray configuration JSON
```json
{
  "log": { "loglevel": "info" },
  "inbounds": [...],
  "outbounds": [...],
  "routing": {...},
  "reverse": {...},
  "metadata": {
    "updated_at": "2025-08-22T10:30:00Z",
    "updated_by": "192.168.1.10"
  }
}
```

**Implementation:**
```go
func (c *Coordinator) syncRemoteConfig(node *database.Node) {
    url := fmt.Sprintf("http://%s:%d/v1/configs", node.Host, node.HttpPort)
    xc := c.writer.RemoteConfig(node, c.state.XrayUpdatedAt(), c.state.XraySharedPassword())
    
    _, err := c.hc.Do(http.MethodPost, url, node.HttpToken, xc)
    if err == nil {
        node.PushStatus = database.NodeStatusAvailable
        node.PushedAt = time.Now().UnixMilli()
    } else {
        node.PushStatus = database.NodeStatusUnavailable
    }
}
```

### 2. Statistics Collection (Node → Manager)

**API Endpoint:** `GET http://{node_host}:{node_port}/v1/stats`

**Response Format:**
```json
[
  {
    "name": "user>>>123>>>traffic>>>downlink",
    "value": 1048576
  },
  {
    "name": "inbound>>>remote>>>traffic>>>downlink", 
    "value": 2097152
  }
]
```

**Stats Parsing:**
```go
func (c *Coordinator) syncRemoteNodeStats(node *database.Node) {
    url := fmt.Sprintf("http://%s:%d/v1/stats", node.Host, node.HttpPort)
    
    response, err := c.hc.Do(http.MethodGet, url, node.HttpToken, nil)
    if err != nil {
        return
    }
    
    var queryStats []*command.Stat
    json.Unmarshal(response, &queryStats)
    
    users := map[string]int64{}
    var nodeUsageBytes int64
    
    for _, qs := range queryStats {
        parts := strings.Split(qs.GetName(), ">>>")
        if parts[0] == "user" {
            users[parts[1]] += qs.GetValue()  // User traffic
        } else if parts[0] == "inbound" && parts[1] == "remote" {
            nodeUsageBytes += qs.GetValue()   // Node total traffic
        }
    }
    
    // Update user usage and check quotas
    // Update node usage statistics
}
```

## Connection Modes

### 1. Relay Mode (`SsRelayPort`)

**Traffic Flow:** Client → Manager → Node → Internet

**Manager Configuration:**
```go
// Manager creates outbound to node
xc.Outbounds = append(xc.Outbounds, xc.MakeShadowsocksOutbound(
    fmt.Sprintf("relay-%d", node.Id),  // "relay-1"
    node.Host,                         // "192.168.1.100"
    key,                               // Generated key
    config.Shadowsocks2022Method,      // "2022-blake3-aes-128-gcm"
    outboundRelayPort,                 // Random port
))

// Add to load balancer
xc.FindBalancer("relay").Selector = append(
    xc.FindBalancer("relay").Selector,
    fmt.Sprintf("relay-%d", node.Id),
)
```

**Node Configuration:**
```go
// Node accepts inbound from manager
relayOutbound := w.xray.Config().FindOutbound(fmt.Sprintf("relay-%d", node.Id))
xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
    "direct",
    relayOutbound.Settings.Servers[0].Password,  // Same key as manager
    relayOutbound.Settings.Servers[0].Method,    // Same method
    "tcp",
    relayOutbound.Settings.Servers[0].Port,      // Same port
    nil,
))
```

### 2. Reverse Mode (`SsReversePort`)

**Traffic Flow:** Client → Manager → [Reverse Tunnel] → Node → Internet

**Manager Configuration:**
```go
// Manager creates internal inbound for reverse connection
xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
    fmt.Sprintf("internal-%d", node.Id),  // "internal-1"
    key,                                  // Generated key
    config.Shadowsocks2022Method,         // Node-to-node encryption
    "tcp",
    inboundPort,                          // Random local port
    nil,
))

// Create reverse proxy portal
xc.Reverse.Portals = append(xc.Reverse.Portals, &xray.ReverseItem{
    Tag:    fmt.Sprintf("portal-%d", node.Id),  // "portal-1"
    Domain: fmt.Sprintf("s%d.reverse.proxy", node.Id), // "s1.reverse.proxy"
})
```

**Node Configuration:**
```go
// Node creates outbound connection back to manager
internalOutbound := w.xray.Config().FindInbound(fmt.Sprintf("internal-%d", node.Id))
xc.Outbounds = append(xc.Outbounds, xc.MakeShadowsocksOutbound(
    "internal",
    w.database.Content.Settings.Host,           // Manager IP
    internalOutbound.Settings.Password,         // Same key
    internalOutbound.Settings.Method,           // Same method
    internalOutbound.Port,                      // Same port
))

// Create reverse bridge
xc.Reverse.Bridges = append(xc.Reverse.Bridges, &xray.ReverseItem{
    Tag:    "bridge",
    Domain: fmt.Sprintf("s%d.reverse.proxy", node.Id),
})
```

### 3. Direct Mode (`SsDirectPort`)

**Traffic Flow:** Client → Manager → Internet (no nodes)

**Manager Configuration:**
```go
xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
    "direct",
    key,                          // Generated key
    config.ShadowsocksMethod,     // Standard user encryption
    "tcp,udp",
    w.database.Content.Settings.SsDirectPort,
    clients,                      // User list
))

xc.Routing.Rules = append(xc.Routing.Rules, &xray.Rule{
    InboundTag:  []string{"direct"},
    OutboundTag: "out",  // Direct to internet
})
```

### 4. Remote Mode (`SsRemotePort`)

**Traffic Flow:** Client → Node → Internet (bypasses manager)

**Node Configuration:**
```go
xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
    "remote",
    password,                                   // Shared password
    config.ShadowsocksMethod,                   // User encryption
    "tcp",
    w.database.Content.Settings.SsRemotePort,
    w.clients(),                                // Full user list
))
```

## Health Monitoring

### Status Tracking

Each node maintains two status indicators:

**Push Status** (Configuration sync):
- `""` (empty): Processing/initial state
- `"available"`: Config successfully pushed, node reachable directly
- `"dirty"`: Config pushed via proxy server (less reliable)
- `"unavailable"`: Cannot reach node for config updates

**Pull Status** (Health monitoring):
- `""` (empty): Processing/initial state  
- `"available"`: Node responding to health checks
- `"unavailable"`: Node not responding (marked after 1 minute timeout)

### Health Check Implementation

```go
func (c *Coordinator) syncNodePullStatuses() error {
    needsSync := false
    for _, n := range c.d.Content.Nodes {
        // Mark as unavailable if no response for 1 minute
        if time.Now().Sub(time.UnixMilli(n.PulledAt)) > time.Minute && 
           n.PullStatus != database.NodeStatusUnavailable {
            n.PullStatus = database.NodeStatusUnavailable
            needsSync = true
        }
    }
    
    if needsSync {
        return c.d.Save()
    }
    return nil
}
```

## Error Handling and Fallbacks

### Proxy Communication

When direct communication fails, the system can route through a proxy:

```go
func (c *Coordinator) syncRemoteConfig(node *database.Node) {
    url := fmt.Sprintf("http://%s:%d/v1/configs", node.Host, node.HttpPort)
    proxy := c.d.Content.Settings.SingetServer  // Proxy server setting
    
    // Try direct connection first
    _, err := c.hc.Do(http.MethodPost, url, node.HttpToken, xc)
    if err == nil {
        node.PushStatus = database.NodeStatusAvailable
        return
    }
    
    // Fallback to proxy if configured
    if proxy != "" {
        _, err = c.hc.DoThrough(proxy, http.MethodPost, url, node.HttpToken, xc)
        if err == nil {
            node.PushStatus = database.NodeStatusDirty  // Marked as "dirty"
            return
        }
    }
    
    // Mark as unavailable if both fail
    node.PushStatus = database.NodeStatusUnavailable
}
```

### Load Balancing

When nodes become unavailable, the system automatically removes them from load balancers:

```go
// Only add healthy nodes to balancer
if node.PushStatus == database.NodeStatusAvailable {
    xc.FindBalancer("relay").Selector = append(
        xc.FindBalancer("relay").Selector,
        fmt.Sprintf("relay-%d", node.Id),
    )
}
```

## Configuration Synchronization

### Sync Triggers

Configurations are pushed to nodes when:
1. **Initial setup**: When node is first added
2. **User changes**: Add/remove/modify users
3. **Settings updates**: Port changes, policy updates
4. **Periodic sync**: Every 10 seconds for outdated configs
5. **Manual restart**: Admin triggers Xray restart

### Sync Process

```go
func (c *Coordinator) SyncConfigs() {
    // Update local Xray first
    if err := c.syncLocalConfig(); err != nil {
        c.l.Fatal("coordinator: cannot sync local configs", zap.Error(err))
    }
    
    // Push to all remote nodes
    c.syncRemoteConfigs()
}

func (c *Coordinator) syncRemoteConfigs() {
    for _, node := range c.d.Content.Nodes {
        go c.syncRemoteConfig(node)  // Parallel execution
    }
}
```

### Configuration Validation

Each generated configuration includes metadata for validation:

```go
xc.Metadata = &xray.Metadata{
    UpdatedAt: lastUpdate.Format(time.RFC3339),  // "2025-08-22T10:30:00Z"
    UpdatedBy: w.database.Content.Settings.Host, // "192.168.1.10"
}
```

## Security Considerations

### Authentication
- Each node requires unique HTTP token
- Tokens should be cryptographically random
- Rotate tokens periodically for security

### Network Security
- Use HTTPS when possible (configure TLS certificates)
- Restrict node API access via firewall
- Monitor for unauthorized access attempts

### Data Protection
- Node communications contain sensitive user data
- Ensure secure channels for API communication
- Log access and configuration changes

This comprehensive communication system ensures reliable, scalable management of distributed proxy nodes while maintaining high availability and security.
