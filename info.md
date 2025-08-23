# Protocol & Transport Compatibility

# Xray Core
## Transport Support Table

| Protocol | TCP | TCP+HTTP | WebSocket | HTTP Upgrade | gRPC | KCP | XHTTP |
|----------|-----|----------|-----------|--------------|------|-----|-------|
| **VMess** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| **VLESS** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Trojan** | ✅ | ❌ | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Shadowsocks** | ❌ | ❌ | ❌* | ❌* | ❌ | ❌* | ❌ |

*\*Shadowsocks transport variants require plugins*

## Security Support Table

| Protocol | None | TLS | Reality |
|----------|------|-----|---------|
| **VMess** | ✅ | ✅ | ❌ |
| **VLESS** | ✅ | ✅ | ✅ |
| **Trojan** | ❌ | ✅ (Required) | ❌ |
| **Shadowsocks** | ✅ | ❌* | ❌ |

*\*Shadowsocks TLS requires plugins*

## Protocol-Specific Details

### VMess
- **Transports**: TCP, TCP+HTTP, WebSocket, HTTP Upgrade, gRPC, KCP
- **Security**: None, TLS
- **Limitations**: No Reality, No XHTTP, No XTLS

### VLESS  
- **Transports**: TCP, TCP+HTTP, WebSocket, HTTP Upgrade, gRPC, KCP, XHTTP
- **Security**: None, TLS, Reality
- **Features**: Full feature support, XTLS, Reality anti-detection

### Trojan
- **Transports**: TCP, TCP+HTTP, WebSocket, HTTP Upgrade, gRPC, KCP
- **Security**: TLS (Mandatory)
- **Features**: Password-based, HTTPS mimicking

### Shadowsocks
- **Transports**: TCP (native), UDP
- **Security**: Built-in encryption
- **Plugins**: v2ray-plugin, obfs, kcptun (for additional transports)
- **Methods**: AEAD ciphers, 2022 edition support

