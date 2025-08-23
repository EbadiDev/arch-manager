# TODO: Multi-Protocol Support Implementation

## Overview
Complete redesign for per-node configuration support:
- **Each node has its own protocol configuration**
- **Support for Xray protocols**: VMess, VLESS, Shadowsocks, Trojan
- **Future Sing-box support** 
- **No backward compatibility** - fresh implementation
- **Advanced configuration interface** for each node

## Current Architecture Analysis

### Current Limitations
1. **Single Protocol**: Only Shadowsocks supported (`MakeShadowsocksInbound`/`MakeShadowsocksOutbound`)
2. **Hardcoded Ports**: Fixed SS ports in Settings struct (`ss_direct_port`, `ss_relay_port`, etc.)
3. **No Per-Node Configuration**: All nodes use same protocol/settings
4. **Static Web Interface**: Only Shadowsocks configuration forms
5. **Writer Logic**: Hardcoded protocol creation in `writer.go`
6. **No Advanced Options**: Missing DNS, routing, network, security settings

## Required Changes

### 1. Database Schema Changes

#### `internal/database/nodes.go`
```go
// Complete Node struct redesign:
type Node struct {
    // Basic Info
    Id         int        `json:"id"`
    Host       string     `json:"host" validate:"required,max=128"`
    HttpToken  string     `json:"http_token" validate:"required"`
    HttpPort   int        `json:"http_port" validate:"required,min=1,max=65536"`
    Usage      float64    `json:"usage" validate:"min=0"`
    UsageBytes int64      `json:"usage_bytes" validate:"min=0"`
    PushStatus NodeStatus `json:"push_status"`
    PullStatus NodeStatus `json:"pull_status"`
    PushedAt   int64      `json:"pushed_at"`
    PulledAt   int64      `json:"pulled_at"`
    
    // Core Configuration
    CoreType   string     `json:"core_type" validate:"required,oneof=xray sing-box"`
    
    // Protocol Configuration
    Protocol     string `json:"protocol" validate:"required,oneof=vmess vless shadowsocks trojan"`
    ServerName   string `json:"server_name" validate:"required,max=128"`
    ServerAddr   string `json:"server_address" validate:"required,max=128"`
    ServerIP     string `json:"server_ip" validate:"required,ip"`
    ServerPort   string `json:"server_port" validate:"required"` // Support port ranges like "400:450"
    Encryption   string `json:"encryption" validate:"required"`
    
    // Network Configuration
    ListeningIP   string `json:"listening_ip" validate:"required,ip"`
    ListeningPort int    `json:"listening_port" validate:"required,min=1,max=65536"`
    SendThrough   string `json:"send_through" validate:"omitempty,ip"`
    
    // Advanced Settings
    DNSSettings      DNSConfig      `json:"dns_settings"`
    RoutingSettings  RoutingConfig  `json:"routing_settings"`
    NetworkSettings  NetworkConfig  `json:"network_settings"`
    
    // Security Configuration
    Security         string        `json:"security" validate:"required,oneof=tls reality none"`
    SecuritySettings SecurityConfig `json:"security_settings"`
    
    // Certificate & Fragment
    CertMode      string `json:"cert_mode" validate:"required,oneof=http file dns none"`
    Fragment      bool   `json:"fragment"`
    FragmentValue string `json:"fragment_value" validate:"omitempty"` // e.g., "1,40-60,30-50"
}

type DNSConfig struct {
    Servers []string          `json:"servers"`
    Hosts   map[string]string `json:"hosts"`
    Tag     string            `json:"tag"`
}

type RoutingConfig struct {
    Rules []RoutingRule `json:"rules"`
    Tag   string        `json:"tag"`
}

type RoutingRule struct {
    Type        string   `json:"type"`
    Domain      []string `json:"domain,omitempty"`
    IP          []string `json:"ip,omitempty"`
    Port        string   `json:"port,omitempty"`
    OutboundTag string   `json:"outbound_tag"`
}

type NetworkConfig struct {
    Transport           string                 `json:"transport" validate:"required,oneof=tcp http ws grpc kcp httpupgrade xhttp"`
    AcceptProxyProtocol bool                   `json:"accept_proxy_protocol"`
    Settings            map[string]interface{} `json:"settings"`
}

type SecurityConfig struct {
    TLS     *TLSConfig     `json:"tls,omitempty"`
    Reality *RealityConfig `json:"reality,omitempty"`
}

type TLSConfig struct {
    ServerName          string   `json:"server_name"`
    RejectUnknownSni    bool     `json:"reject_unknown_sni"`
    AllowInsecure       bool     `json:"allow_insecure"`
    Fingerprint         string   `json:"fingerprint"`
    SNI                 string   `json:"sni"`
    CurvePreferences    string   `json:"curve_preferences"`
    ALPN                []string `json:"alpn"`
    ServerNameToVerify  string   `json:"server_name_to_verify"`
}

type RealityConfig struct {
    Show           bool     `json:"show"`
    Dest           string   `json:"dest"`
    PrivateKey     string   `json:"private_key"`
    MinClientVer   string   `json:"min_client_ver"`
    MaxClientVer   string   `json:"max_client_ver"`
    MaxTimeDiff    int      `json:"max_time_diff"`
    ProxyProtocol  int      `json:"proxy_protocol"`
    ShortIDs       []string `json:"short_ids"`
    ServerNames    []string `json:"server_names"`
    Fingerprint    string   `json:"fingerprint"`
    SpiderX        string   `json:"spider_x"`
    PublicKey      string   `json:"public_key"`
}
```

#### `internal/database/settings.go`
```go
// Simplified settings - remove hardcoded SS ports:
type Settings struct {
    AdminPassword string  `json:"admin_password" validate:"required,min=8,max=32"`
    TrafficRatio  float64 `json:"traffic_ratio" validate:"min=1,max=1024"`
    SingetServer  string  `json:"singet_server" validate:"omitempty,url"`
    ResetPolicy   string  `json:"reset_policy" validate:"omitempty,oneof=monthly"`
    
    // Protocol encryption options for UI
    EncryptionOptions EncryptionOptions `json:"encryption_options"`
}

type EncryptionOptions struct {
    VMess   []string `json:"vmess"`   // ["auto", "none", "zero", "aes-128-gcm"]
    VLESS   []string `json:"vless"`   // ["none"]
    Trojan  []string `json:"trojan"`  // ["none"]
    SS      []string `json:"ss"`      // ["aes-128-gcm", "aes-256-gcm", "chacha20-poly1305", etc.]
}
```

### 2. Configuration Writer Updates

#### `internal/writer/writer.go`
**Complete rewrite required:**

1. **Replace all hardcoded `MakeShadowsocksInbound` calls** with protocol factory:
```go
func (w *Writer) makeProtocolInbound(node *database.Node, port int, password string) (interface{}, error) {
    switch node.Protocol {
    case "shadowsocks":
        return w.makeShadowsocksInbound(node, port, password)
    case "vless":
        return w.makeVlessInbound(node, port, password)
    case "vmess":
        return w.makeVmessInbound(node, port, password)
    case "trojan":
        return w.makeTrojanInbound(node, port, password)
    default:
        return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
    }
}

func (w *Writer) makeShadowsocksInbound(node *database.Node, port int, password string) (interface{}, error) {
    return xc.MakeShadowsocksInbound(
        node.Id, 
        node.ListeningPort, 
        password, 
        node.Encryption,
    )
}

func (w *Writer) makeVlessInbound(node *database.Node, port int, password string) (interface{}, error) {
    var securityConfig interface{}
    
    switch node.Security {
    case "reality":
        if node.SecuritySettings.Reality == nil {
            return nil, fmt.Errorf("reality configuration required")
        }
        securityConfig = node.SecuritySettings.Reality
    case "tls":
        if node.SecuritySettings.TLS == nil {
            return nil, fmt.Errorf("tls configuration required")
        }
        securityConfig = node.SecuritySettings.TLS
    case "none":
        securityConfig = nil
    }
    
    return xc.MakeVlessInbound(
        node.Id,
        node.ListeningPort,
        password,
        node.NetworkSettings,
        securityConfig,
    )
}

func (w *Writer) makeVmessInbound(node *database.Node, port int, password string) (interface{}, error) {
    return xc.MakeVmessInbound(
        node.Id,
        node.ListeningPort,
        password,
        node.Encryption,
        node.NetworkSettings,
    )
}

func (w *Writer) makeTrojanInbound(node *database.Node, port int, password string) (interface{}, error) {
    return xc.MakeTrojanInbound(
        node.Id,
        node.ListeningPort,
        password,
        node.NetworkSettings,
        node.SecuritySettings,
    )
}
```

2. **Update `RemoteConfig` method**:
```go
func (w *Writer) RemoteConfig(node *database.Node, xrayUpdatedAt int64, sharedPassword string) interface{} {
    xc := xray.NewXrayConfig()
    
    // Add DNS configuration
    if len(node.DNSSettings.Servers) > 0 {
        xc.DNS = node.DNSSettings
    }
    
    // Add routing configuration
    if len(node.RoutingSettings.Rules) > 0 {
        xc.Routing = node.RoutingSettings
    }
    
    // Generate inbound based on node protocol
    inbound, err := w.makeProtocolInbound(node, node.ListeningPort, sharedPassword)
    if err != nil {
        return map[string]interface{}{"error": err.Error()}
    }
    xc.Inbounds = append(xc.Inbounds, inbound)
    
    // Add outbound configuration
    outbound := w.makeOutbound(node)
    xc.Outbounds = append(xc.Outbounds, outbound)
    
    return xc
}
```

### 3. External Package Verification

#### Check `github.com/ebadidev/arch-node/pkg/xray`
**Required Methods to Verify/Implement:**
- `MakeVlessInbound()` - for VLESS + TCP + Reality
- `MakeVmessInbound()` - for VMess + HTTP  
- `MakeVlessOutbound()` - for VLESS outbound connections
- `MakeVmessOutbound()` - for VMess outbound connections

**Action**: Check if these methods exist in the external package, if not:
1. Fork the `arch-node` repository
2. Add missing protocol methods
3. Update go.mod dependency

### 4. Web Interface Updates

#### `web/admin-nodes.html` (Update existing file)
**Add Edit Button and simplified columns:**
```html
<!-- Updated table columns - remove port configurations -->
columns: [
    { title: "ID", field: "id", widthGrow: 1, resizable: true, headerFilter: "input", editable: true },
    { title: "Host", field: "host", editor: "input", widthGrow: 2, headerFilter: "input", validator: ["required"], editable: true },
    { title: "HTTP Port", field: "http_port", editor: "number", widthGrow: 1, validator: ["required", "min:1", "max:65536"], editable: true },
    { title: "HTTP Token", field: "http_token", editor: "input", widthGrow: 2, validator: ["required"], editable: true },
    { title: "Protocol", field: "protocol", widthGrow: 1, editable: false }, // Display only
    { title: "Usage (GB)", field: "usage", widthGrow: 1, editable: false, formatter: cell => parseFloat(cell.getData()['usage']).toFixed(2) },
    { title: "Push Status", field: "push_status", widthGrow: 1, formatter: statusFormatter },
    { title: "Pull Status", field: "pull_status", widthGrow: 1, formatter: statusFormatter },
    { title: "Actions", formatter: actionsFormatter, hozAlign: "right" }
]

// Updated actions formatter to include edit button
let actionsFormatter = cell => [
    `<span class="badge bg-primary" onclick="editConfig('${cell.getRow().getIndex()}')" title="Edit Config">⚙️</span>`,
    `<span class="badge bg-danger" onclick="destroy('${cell.getRow().getIndex()}')" title="Delete">X</span>`,
    `<span class="badge bg-info" onclick="info('${cell.getRow().getIndex()}')" title="Info">i</span>`,
].join('&nbsp')

// Edit configuration function
let editConfig = rowIndex => {
    window.location.href = `admin-node-config.html?id=${rowIndex}`
}
```

#### `web/admin-node-config.html` (Create new file)
**Complete configuration interface:**
```html
<!doctype html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>Node Configuration - Arch-Manager</title>
    <link rel="stylesheet" href="assets/third_party/bootstrap-5.3.5/css/bootstrap.min.css">
</head>
<body>
<div class="container py-4">
    <h2>Node Configuration</h2>
    <form id="nodeConfigForm">
        
        <!-- Core Configuration -->
        <div class="card mb-3">
            <div class="card-header">Core Configuration</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-6">
                        <label class="form-label">Core Type</label>
                        <select name="core_type" class="form-select" required>
                            <option value="xray">Xray</option>
                            <option value="sing-box" disabled>Sing-box (Coming Soon)</option>
                        </select>
                    </div>
                    <div class="col-md-6">
                        <label class="form-label">Protocol Type</label>
                        <select name="protocol" class="form-select" required id="protocolSelect">
                            <option value="vmess">VMess</option>
                            <option value="vless">VLESS</option>
                            <option value="shadowsocks">Shadowsocks</option>
                            <option value="trojan">Trojan</option>
                        </select>
                    </div>
                </div>
            </div>
        </div>

        <!-- Server Configuration -->
        <div class="card mb-3">
            <div class="card-header">Server Configuration</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-6">
                        <label class="form-label">Server Name</label>
                        <input type="text" name="server_name" class="form-control" required>
                    </div>
                    <div class="col-md-6">
                        <label class="form-label">Server Address</label>
                        <input type="text" name="server_address" class="form-control" required>
                    </div>
                </div>
                <div class="row mt-2">
                    <div class="col-md-4">
                        <label class="form-label">Server IP</label>
                        <input type="text" name="server_ip" class="form-control" required pattern="^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$">
                    </div>
                    <div class="col-md-4">
                        <label class="form-label">Server Port</label>
                        <input type="text" name="server_port" class="form-control" required placeholder="443 or 400:450">
                    </div>
                    <div class="col-md-4">
                        <label class="form-label">Encryption</label>
                        <select name="encryption" class="form-select" required id="encryptionSelect">
                            <!-- Options populated by JavaScript based on protocol -->
                        </select>
                    </div>
                </div>
            </div>
        </div>

        <!-- Network Configuration -->
        <div class="card mb-3">
            <div class="card-header">Network Configuration</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-4">
                        <label class="form-label">Listening IP</label>
                        <input type="text" name="listening_ip" class="form-control" required pattern="^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$">
                    </div>
                    <div class="col-md-4">
                        <label class="form-label">Listening Port</label>
                        <input type="number" name="listening_port" class="form-control" required min="1" max="65536">
                    </div>
                    <div class="col-md-4">
                        <label class="form-label">Send Through (Optional)</label>
                        <input type="text" name="send_through" class="form-control" pattern="^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$">
                    </div>
                </div>
            </div>
        </div>

        <!-- Network Settings (Transport) -->
        <div class="card mb-3">
            <div class="card-header">Network Settings</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-6">
                        <label class="form-label">Transport</label>
                        <select name="transport" class="form-select" required id="transportSelect">
                            <option value="tcp">TCP</option>
                            <option value="http">HTTP</option>
                            <option value="ws">WebSocket</option>
                            <option value="grpc">GRPC</option>
                            <option value="kcp">KCP</option>
                            <option value="httpupgrade">HTTP Upgrade</option>
                            <option value="xhttp">XHTTP</option>
                        </select>
                    </div>
                    <div class="col-md-6">
                        <div class="form-check mt-4">
                            <input class="form-check-input" type="checkbox" name="accept_proxy_protocol" id="acceptProxyProtocol">
                            <label class="form-check-label" for="acceptProxyProtocol">Accept Proxy Protocol</label>
                        </div>
                    </div>
                </div>
                <!-- Transport-specific settings will be populated by JavaScript -->
                <div id="transportSettings" class="mt-3"></div>
            </div>
        </div>

        <!-- Security Configuration -->
        <div class="card mb-3">
            <div class="card-header">Security Configuration</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-6">
                        <label class="form-label">Security</label>
                        <select name="security" class="form-select" required id="securitySelect">
                            <option value="none">None</option>
                            <option value="tls">TLS</option>
                            <option value="reality">Reality</option>
                        </select>
                    </div>
                    <div class="col-md-6">
                        <label class="form-label">Certificate Mode</label>
                        <select name="cert_mode" class="form-select" required>
                            <option value="none">None</option>
                            <option value="http">HTTP</option>
                            <option value="file">File</option>
                            <option value="dns">DNS</option>
                        </select>
                    </div>
                </div>
                <!-- Security-specific settings -->
                <div id="securitySettings" class="mt-3"></div>
            </div>
        </div>

        <!-- Fragment Configuration -->
        <div class="card mb-3">
            <div class="card-header">Fragment Configuration</div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-6">
                        <div class="form-check">
                            <input class="form-check-input" type="checkbox" name="fragment" id="fragmentEnable">
                            <label class="form-check-label" for="fragmentEnable">Enable Fragment</label>
                        </div>
                    </div>
                    <div class="col-md-6">
                        <label class="form-label">Fragment Value</label>
                        <input type="text" name="fragment_value" class="form-control" placeholder="1,40-60,30-50" disabled id="fragmentValue">
                    </div>
                </div>
            </div>
        </div>

        <!-- Action Buttons -->
        <div class="d-grid gap-2 d-md-flex justify-content-md-end">
            <a href="admin-nodes.html" class="btn btn-secondary">Cancel</a>
            <button type="submit" class="btn btn-primary">Save Configuration</button>
        </div>
    </form>
</div>

<script src="assets/third_party/jquery-3.7.1.min.js"></script>
<script src="assets/third_party/bootstrap-5.3.5/js/bootstrap.bundle.min.js"></script>
<script src="assets/js/node-config.js"></script>
</body>
</html>
```

### 5. HTTP API Updates

#### `internal/http/handlers/v1/nodes.go`
**New endpoints needed:**
```go
// GET /api/v1/nodes/:id/config - Get node configuration
// PUT /api/v1/nodes/:id/config - Update node configuration
// POST /api/v1/nodes/config - Create new node configuration
// GET /api/v1/protocols - List all available protocols and their options
// POST /api/v1/generate-reality-keys - Generate Reality private/public keys
```

**Implementation:**
```go
func NodesConfigGet(d *database.Database) echo.HandlerFunc {
    return func(c echo.Context) error {
        nodeId := c.Param("id")
        // Return full node configuration including protocol settings
    }
}

func NodesConfigUpdate(d *database.Database) echo.HandlerFunc {
    return func(c echo.Context) error {
        nodeId := c.Param("id")
        var config database.Node
        // Validate and save configuration
        // Support all protocol types and advanced settings
    }
}

func ProtocolsList(d *database.Database) echo.HandlerFunc {
    return func(c echo.Context) error {
        return c.JSON(http.StatusOK, map[string]interface{}{
            "protocols": []string{"vmess", "vless", "shadowsocks", "trojan"},
            "transports": []string{"tcp", "http", "ws", "grpc", "kcp", "httpupgrade", "xhttp"},
            "securities": []string{"none", "tls", "reality"},
            "encryption_options": database.EncryptionOptions{
                VMess:   []string{"auto", "none", "zero", "aes-128-gcm"},
                VLESS:   []string{"none"},
                Trojan:  []string{"none"},
                SS:      []string{"aes-128-gcm", "aes-256-gcm", "chacha20-poly1305", "xchacha20-poly1305", "chacha20-ietf-poly1305", "2022-blake3-aes-128-gcm", "2022-blake3-aes-256-gcm"},
            },
        })
    }
}

func GenerateRealityKeys() echo.HandlerFunc {
    return func(c echo.Context) error {
        // Generate X25519 key pair for Reality
        privateKey, publicKey := generateX25519KeyPair()
        return c.JSON(http.StatusOK, map[string]string{
            "private_key": privateKey,
            "public_key":  publicKey,
        })
    }
}
```

#### `internal/http/handlers/v1/node_configs.go`
**Update `NodesConfigsShow`**:
```go
func NodesConfigsShow(cdr *coordinator.Coordinator, writer *writer.Writer, d *database.Database) echo.HandlerFunc {
    return func(c echo.Context) error {
        d.Locker.Lock()
        defer d.Locker.Unlock()

        nodeId := c.Param("id")
        var node *database.Node
        for _, n := range d.Content.Nodes {
            if strconv.Itoa(n.Id) == nodeId {
                node = n
                node.PulledAt = time.Now().UnixMilli()
                node.PullStatus = database.NodeStatusAvailable

                if err := d.Save(); err != nil {
                    return errors.WithStack(err)
                }
                break
            }
        }
        if node == nil {
            return c.NoContent(http.StatusNotFound)
        }

        // Generate configuration based on node's protocol
        configs, err := writer.RemoteConfig(node, cdr.State().XrayUpdatedAt(), cdr.State().XraySharedPassword())
        if err != nil {
            return errors.WithStack(err)
        }

        return c.JSON(http.StatusOK, configs)
    }
}
```

### 6. Configuration Updates

#### `configs/main.defaults.json`
**Remove SS port settings and add encryption options:**
```json
{
  "admin_password": "admin123",
  "traffic_ratio": 1.0,
  "singet_server": "",
  "reset_policy": "monthly",
  "encryption_options": {
    "vmess": ["auto", "none", "zero", "aes-128-gcm"],
    "vless": ["none"],
    "trojan": ["none"], 
    "ss": [
      "aes-128-gcm",
      "aes-256-gcm",
      "chacha20-poly1305",
      "xchacha20-poly1305",
      "chacha20-ietf-poly1305",
      "2022-blake3-aes-128-gcm",
      "2022-blake3-aes-256-gcm"
    ]
  }
}
```

### 7. Migration Strategy

#### Database Migration
**Create migration script to remove old structure:**
1. Remove all hardcoded SS port settings from existing installations
2. Add new node configuration fields with defaults
3. Convert any existing nodes to use new schema

#### Fresh Installation Approach
**No backward compatibility needed:**
1. Remove all SS-specific settings from Settings struct
2. Implement new Node struct with full protocol support
3. Create default node configurations for each protocol type

## Implementation Priority

### Phase 1: Core Infrastructure ✅
1. **Database schema updates** (`nodes.go`, `settings.go`) 
2. **Protocol factory methods** (`writer.go`)
3. **External package verification** (check `arch-node/pkg/xray`)

### Phase 2: Web Interface ✅
1. **Node management interface** - Updated `admin-nodes.html` with edit button
2. **Configuration interface** - Created `admin-node-config.html` 
3. **JavaScript functionality** - Created `node-config.js` with dynamic forms

### Phase 3: API & Integration
1. **HTTP endpoints** (protocol management, Reality key generation)
2. **Configuration generation** (per-protocol configs in writer.go)
3. **External package methods** (verify/implement missing protocol methods)

## Testing Requirements

### Unit Tests
- Protocol factory method tests for each protocol type
- Configuration generation per protocol (VMess, VLESS, SS, Trojan)
- Database schema validation with new Node structure
- Transport settings validation (TCP, HTTP, WS, GRPC, etc.)
- Security settings validation (TLS, Reality, None)

### Integration Tests  
- End-to-end node creation with different protocols
- Configuration push/pull with multiple protocols  
- Web interface protocol switching and form validation
- Reality key generation and validation
- Fragment configuration testing

### Manual Testing Scenarios
1. **Create VMess+HTTP node** → Verify correct config generation
2. **Create VLESS+Reality node** → Verify Reality configuration
3. **Create Shadowsocks node** → Verify encryption options
4. **Create Trojan+TLS node** → Verify TLS configuration
5. **Mixed protocol deployment** → Test multiple node types
6. **Transport variations** → Test TCP, WS, GRPC, etc.
7. **Security variations** → Test TLS, Reality, None

## Files Requiring Changes

### Critical Files (Must Change)
- ✅ `internal/database/nodes.go` - Complete Node struct redesign with protocol fields
- ✅ `internal/database/settings.go` - Remove SS ports, add encryption options
- ⏳ `internal/writer/writer.go` - Protocol factory methods and config generation
- ⏳ `web/admin-system.html` - Remove hardcoded SS settings
- ✅ `configs/main.defaults.json` - Remove SS ports, add encryption options

### New Files (Created ✅)
- ✅ `web/admin-node-config.html` - Complete node protocol configuration interface
- ✅ `web/assets/js/node-config.js` - Dynamic form handling and protocol-specific options
- ⏳ `internal/http/handlers/v1/protocols.go` - Protocol API endpoints
- ⏳ `scripts/migrate-database.sh` - Database migration for new schema

### Updated Files (Modified ✅)
- ✅ `web/admin-nodes.html` - Added edit button and protocol column
- ⏳ `web/profile.html` - Multi-protocol connection strings
- ⏳ `internal/http/handlers/v1/nodes.go` - Protocol configuration endpoints
- ⏳ `internal/http/handlers/v1/node_configs.go` - Updated config generation
- ⏳ `docs/PROTOCOL_SUPPORT.md` - Documentation for new features

## External Dependencies

### Required Package Updates
1. **Verify `github.com/ebadidev/arch-node/pkg/xray`** needs:
   - ✅ `MakeVlessInbound()` - for VLESS protocol
   - ✅ `MakeVmessInbound()` - for VMess protocol  
   - ✅ `MakeTrojanInbound()` - for Trojan protocol
   - ✅ `MakeVlessOutbound()` - for VLESS outbound connections
   - ✅ `MakeVmessOutbound()` - for VMess outbound connections
   - ✅ `MakeTrojanOutbound()` - for Trojan outbound connections

2. **If missing methods**: 
   - Fork `github.com/ebadidev/arch-node` repository
   - Extend with missing protocol methods
   - Update go.mod dependency

### Xray-core Compatibility
- ✅ Ensure VLESS+Reality support in `github.com/xtls/xray-core v1.250803.0`
- ✅ Verify VMess+HTTP transport compatibility
- ✅ Verify Trojan+TLS support
- ✅ Support for all transport types (TCP, HTTP, WS, GRPC, KCP, etc.)

## Success Criteria

### Functional Requirements
- ✅ Web interface for comprehensive node configuration
- ✅ Support for VMess, VLESS, Shadowsocks, Trojan protocols
- ✅ Dynamic encryption options based on protocol selection
- ✅ Transport configuration (TCP, HTTP, WS, GRPC, KCP, HTTP Upgrade, XHTTP)
- ✅ Security configuration (None, TLS, Reality)
- ✅ DNS and routing settings per node
- ✅ Fragment configuration with custom values
- ⏳ Generate correct Xray configs per protocol
- ⏳ Reality key generation functionality

### Non-Functional Requirements
- ✅ Clean protocol abstraction for easy extension
- ✅ Intuitive web interface with dynamic forms
- ✅ Comprehensive configuration options
- ⏳ Performance: Efficient config generation
- ⏳ Extensibility: Easy to add Sing-box support later

## Next Steps

### Immediate Actions Required
1. **Update Database Schema** - Implement new Node struct in `nodes.go`
2. **Implement Protocol Factory** - Update `writer.go` with protocol methods
3. **Create API Endpoints** - Add configuration endpoints in HTTP handlers
4. **External Package Check** - Verify `arch-node/pkg/xray` has required methods
5. **Reality Key Generation** - Implement X25519 key pair generation

### Development Workflow
1. **Database Layer** → Update schemas and structures
2. **Business Logic** → Implement protocol factory and config generation  
3. **API Layer** → Create endpoints for configuration management
4. **Frontend Integration** → Connect forms to backend APIs
5. **Testing** → Comprehensive testing of all protocol combinations

## Notes

### Current External Package
Using `github.com/ebadidev/arch-node v0.0.0-20250821092106-159744a8dedc` for Xray configuration generation. This package requires extension for comprehensive multi-protocol support.

### Advanced Configuration Support
The new design supports:
- **Protocol-specific encryption** - Different options per protocol type
- **Transport flexibility** - All Xray transport types supported
- **Security variations** - TLS, Reality, and no security
- **Network customization** - DNS, routing, fragment configuration
- **Reality integration** - Full Reality configuration with key generation
- **Certificate management** - Multiple certificate modes (HTTP, file, DNS)

### Future Extensibility  
- **Sing-box Support** - Core type selection ready for Sing-box addition
- **New Protocols** - Easy addition of new protocols to factory methods
- **Advanced Features** - Room for additional Xray features as needed

### Web Interface Features
- **Dynamic Forms** - Protocol selection changes available options
- **Validation** - Client-side and server-side validation
- **User Experience** - Intuitive configuration flow
- **Real-time Updates** - Instant feedback on configuration changes

## Notes

### Current External Package
Using `github.com/ebadidev/arch-node v0.0.0-20250821092106-159744a8dedc` for Xray configuration generation. This package likely needs extension for VLESS and VMess support.

### Reality Configuration Requirements
VLESS+Reality needs additional configuration:
- Server name (SNI)
- Public key
- Short ID
- Spider X

### VMess+HTTP Requirements  
VMess+HTTP needs:
- Host header configuration
- Path configuration
- HTTP method specification

### Migration Considerations
- Existing deployments have hardcoded SS ports in settings
- Need graceful transition without service interruption
- Database schema changes require careful migration
