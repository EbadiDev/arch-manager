# TODO: Arch-Node Package Updates for Multi-Protocol Support

This document outlines the required updates to the `github.com/ebadidev/arch-node` package to support the new multi-protocol configuration system implemented in ArchNet Manager.

## Overview

The manager has been updated to support comprehensive multi-protocol configuration (VMess, VLESS, Shadowsocks, Trojan) with advanced features like Reality, TLS, and custom transport settings. The arch-node package needs corresponding updates to handle these configurations.

## Required Changes

### 1. Missing Protocol Factory Methods in `pkg/xray/config.go`

**Current Status**: Only `MakeShadowsocksInbound` and `MakeShadowsocksOutbound` exist.

**Required Methods**:

```go
// Add these methods to the Config struct in pkg/xray/config.go

func (c *Config) MakeVlessInbound(tag, uuid string, port int, decryption string, clients []*VlessClient) *Inbound {
    return &Inbound{
        Tag:      tag,
        Protocol: "vless",
        Listen:   "0.0.0.0",
        Port:     port,
        Settings: &InboundSettings{
            Clients:    clients,
            Decryption: decryption,
        },
    }
}

func (c *Config) MakeVmessInbound(tag, uuid string, port int, alterId int, clients []*VmessClient) *Inbound {
    return &Inbound{
        Tag:      tag,
        Protocol: "vmess",
        Listen:   "0.0.0.0",
        Port:     port,
        Settings: &InboundSettings{
            Clients: clients,
        },
    }
}

func (c *Config) MakeTrojanInbound(tag string, port int, clients []*TrojanClient) *Inbound {
    return &Inbound{
        Tag:      tag,
        Protocol: "trojan",
        Listen:   "0.0.0.0",
        Port:     port,
        Settings: &InboundSettings{
            Clients: clients,
        },
    }
}
```

### 2. Extended Client Types

**Current**: Only basic `Client` struct exists with Password, Method, Email fields.

**Required**: Protocol-specific client structures:

```go
// Add these types to pkg/xray/config.go

type VlessClient struct {
    ID    string `json:"id" validate:"required"`
    Level int    `json:"level,omitempty"`
    Email string `json:"email,omitempty"`
}

type VmessClient struct {
    ID      string `json:"id" validate:"required"`
    AlterID int    `json:"alterId,omitempty"`
    Level   int    `json:"level,omitempty"`
    Email   string `json:"email,omitempty"`
}

type TrojanClient struct {
    Password string `json:"password" validate:"required"`
    Level    int    `json:"level,omitempty"`
    Email    string `json:"email,omitempty"`
}
```

### 3. Enhanced InboundSettings Structure

**Current**: Limited to Address, Clients, Network, Method, Password fields.

**Required**: Extended support for all protocol settings:

```go
// Update InboundSettings in pkg/xray/config.go
type InboundSettings struct {
    // Existing fields
    Address  string    `json:"address,omitempty"`
    Clients  []*Client `json:"clients,omitempty" validate:"omitempty,dive"`
    Network  string    `json:"network,omitempty"`
    Method   string    `json:"method,omitempty"`
    Password string    `json:"password,omitempty"`
    
    // New protocol-specific fields
    Decryption    string          `json:"decryption,omitempty"`     // VLESS
    VlessClients  []*VlessClient  `json:"vlessClients,omitempty"`
    VmessClients  []*VmessClient  `json:"vmessClients,omitempty"`
    TrojanClients []*TrojanClient `json:"trojanClients,omitempty"`
}
```

### 4. Stream Settings Enhancements

**Current**: Basic StreamSettings with only Network field.

**Required**: Comprehensive stream settings for TLS, Reality, and transport:

```go
// Update StreamSettings in pkg/xray/config.go
type StreamSettings struct {
    Network    string          `json:"network" validate:"required"`
    Security   string          `json:"security,omitempty"`
    TLSSettings    *TLSSettings    `json:"tlsSettings,omitempty"`
    RealitySettings *RealitySettings `json:"realitySettings,omitempty"`
    TCPSettings    *TCPSettings    `json:"tcpSettings,omitempty"`
    WSSettings     *WSSettings     `json:"wsSettings,omitempty"`
    GRPCSettings   *GRPCSettings   `json:"grpcSettings,omitempty"`
}

type TLSSettings struct {
    ServerName   string          `json:"serverName,omitempty"`
    Certificates []*Certificate  `json:"certificates,omitempty"`
    ALPN         []string        `json:"alpn,omitempty"`
}

type RealitySettings struct {
    Show         bool     `json:"show"`
    Dest         string   `json:"dest"`
    ServerNames  []string `json:"serverNames"`
    PrivateKey   string   `json:"privateKey"`
    ShortIds     []string `json:"shortIds"`
}

type TCPSettings struct {
    Header map[string]interface{} `json:"header,omitempty"`
}

type WSSettings struct {
    Path    string            `json:"path,omitempty"`
    Headers map[string]string `json:"headers,omitempty"`
}

type GRPCSettings struct {
    ServiceName string `json:"serviceName,omitempty"`
}

type Certificate struct {
    CertificateFile string `json:"certificateFile"`
    KeyFile         string `json:"keyFile"`
}
```

### 5. DNS Structure Field Updates

**Current**: Simple DNS struct with only Servers field.

**Issue**: Manager writer expects `Hosts` field, but arch-node DNS struct doesn't have it.

**Required**: Add missing DNS fields:

```go
// Update DNS struct in pkg/xray/config.go
type DNS struct {
    Servers []string                 `json:"servers" validate:"required"`
    Hosts   map[string]interface{}   `json:"hosts,omitempty"`
    Tag     string                   `json:"tag,omitempty"`
}
```

### 6. Routing Rule Structure Updates

**Current**: Rule struct has InboundTag, OutboundTag, BalancerTag, Domain fields.

**Issue**: Manager writer expects `Type`, `IP`, `Port` fields that don't exist.

**Required**: Extend Rule struct:

```go
// Update Rule struct in pkg/xray/config.go
type Rule struct {
    // Existing fields
    InboundTag  []string `json:"inboundTag,omitempty"`
    OutboundTag string   `json:"outboundTag,omitempty"`
    BalancerTag string   `json:"balancerTag,omitempty"`
    Domain      []string `json:"domain,omitempty"`
    
    // Missing fields expected by manager
    Type string   `json:"type,omitempty"`
    IP   []string `json:"ip,omitempty"`
    Port string   `json:"port,omitempty"`
}
```

### 7. Outbound Settings Enhancements

**Current**: Basic OutboundSettings for Shadowsocks only.

**Required**: Support for all protocol outbounds:

```go
// Update OutboundSettings in pkg/xray/config.go
type OutboundSettings struct {
    // Existing Shadowsocks support
    Servers []*OutboundServer `json:"servers,omitempty" validate:"omitempty,dive"`
    
    // VMess/VLESS/Trojan support
    VlessSettings  *VlessOutboundSettings  `json:"vlessSettings,omitempty"`
    VmessSettings  *VmessOutboundSettings  `json:"vmessSettings,omitempty"`
    TrojanSettings *TrojanOutboundSettings `json:"trojanSettings,omitempty"`
}

type VlessOutboundSettings struct {
    Users []*VlessUser `json:"users"`
}

type VmessOutboundSettings struct {
    Users []*VmessUser `json:"users"`
}

type TrojanOutboundSettings struct {
    Servers []*TrojanServer `json:"servers"`
}

type VlessUser struct {
    ID         string `json:"id"`
    Encryption string `json:"encryption"`
    Level      int    `json:"level,omitempty"`
}

type VmessUser struct {
    ID      string `json:"id"`
    AlterID int    `json:"alterId"`
    Level   int    `json:"level,omitempty"`
}

type TrojanServer struct {
    Address  string `json:"address"`
    Port     int    `json:"port"`
    Password string `json:"password"`
    Level    int    `json:"level,omitempty"`
}
```

### 8. Add Missing Outbound Factory Methods

**Required**: Complete the outbound factory pattern:

```go
// Add these methods to Config struct in pkg/xray/config.go

func (c *Config) MakeVlessOutbound(tag, address, uuid string, port int) *Outbound {
    return &Outbound{
        Tag:      tag,
        Protocol: "vless",
        Settings: &OutboundSettings{
            VlessSettings: &VlessOutboundSettings{
                Users: []*VlessUser{
                    {
                        ID:         uuid,
                        Encryption: "none",
                    },
                },
            },
        },
        StreamSettings: &StreamSettings{
            Network: "tcp",
        },
    }
}

func (c *Config) MakeVmessOutbound(tag, address, uuid string, port, alterId int) *Outbound {
    return &Outbound{
        Tag:      tag,
        Protocol: "vmess",
        Settings: &OutboundSettings{
            VmessSettings: &VmessOutboundSettings{
                Users: []*VmessUser{
                    {
                        ID:      uuid,
                        AlterID: alterId,
                    },
                },
            },
        },
        StreamSettings: &StreamSettings{
            Network: "tcp",
        },
    }
}

func (c *Config) MakeTrojanOutbound(tag, address, password string, port int) *Outbound {
    return &Outbound{
        Tag:      tag,
        Protocol: "trojan",
        Settings: &OutboundSettings{
            TrojanSettings: &TrojanOutboundSettings{
                Servers: []*TrojanServer{
                    {
                        Address:  address,
                        Port:     port,
                        Password: password,
                    },
                },
            },
        },
        StreamSettings: &StreamSettings{
            Network: "tcp",
        },
    }
}
```

## Implementation Priority

### Phase 1: Critical (Blocks Manager)
1. Add missing protocol factory methods (`MakeVlessInbound`, `MakeVmessInbound`, `MakeTrojanInbound`)
2. Fix DNS struct (`Hosts` field)
3. Fix Rule struct (`Type`, `IP`, `Port` fields)

### Phase 2: Protocol Support
1. Add protocol-specific client types
2. Extend InboundSettings structure
3. Add outbound factory methods

### Phase 3: Advanced Features
1. Implement comprehensive StreamSettings
2. Add TLS/Reality support structures
3. Complete transport settings (TCP, WebSocket, gRPC)

## Testing Requirements

After implementing these changes, test with:

1. **Basic Protocol Tests**: Ensure each protocol (VMess, VLESS, Shadowsocks, Trojan) can be configured
2. **Manager Integration**: Verify manager can successfully send configurations to arch-node
3. **Xray Compatibility**: Confirm generated configurations work with actual Xray-core
4. **Backward Compatibility**: Ensure existing Shadowsocks configurations continue working

## Migration Notes

- All changes should be backward compatible
- Existing `Client` struct should remain for Shadowsocks support
- New fields should be optional (`omitempty` JSON tags)
- Consider version field in config for future migrations

## Estimated Impact

- **Files to Modify**: Primarily `pkg/xray/config.go`
- **Breaking Changes**: None (additive changes only)
- **Testing Scope**: Full protocol support testing required
- **Documentation**: Update xray integration docs with new protocol examples

---

**Note**: This TODO represents the minimum viable changes needed for the manager's multi-protocol support to function with arch-node. Additional enhancements may be needed for full feature parity with Xray-core capabilities.
