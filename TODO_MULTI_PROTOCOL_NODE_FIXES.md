# TODO: Multi-Protocol Support - Arch-Node Fixes

## 🔍 Issue Summary
After implementing multi-protocol support in arch-manager, arch-node fails to start with VMess configurations due to validation errors. The node expects Shadowsocks-specific fields (`Method`) for all protocols, but VMess uses different field structures.

**Error**: `Key: 'Config.Outbounds[1].Settings.Servers[0].Method' Error:Field validation for 'Method' failed on the 'required' tag`

## ✅ Completed Changes in Arch-Manager

### 1. Protocol-Aware Configuration Generation
- **File**: `internal/writer/writer.go`
- **Change**: Replaced hardcoded `MakeShadowsocksInbound()` with `makeProtocolInbound()`
- **Impact**: Now generates protocol-specific configurations (VMess, VLESS, Trojan, Shadowsocks)
- **Line**: ~207 in `LocalConfig()` function

### 2. UUID Generation for VMess/VLESS
- **File**: `internal/writer/writer.go`
- **Change**: Added proper UUID generation for VMess/VLESS vs Shadowsocks keys
- **Impact**: VMess configurations now use valid UUIDs instead of Shadowsocks-style keys

### 3. API Endpoints for Node Configuration
- **File**: `internal/http/handlers/v1/protocols.go`
- **Change**: Added `NodeConfigGet()`, `NodeConfigUpdate()`, `NodeConfigCreate()`
- **Impact**: Proper API endpoints for node configuration management

### 4. Connection Field Preservation
- **File**: `internal/http/handlers/v1/protocols.go`
- **Change**: Fixed `NodeConfigUpdate()` to preserve `host`, `http_token`, `http_port`
- **Impact**: Protocol changes no longer delete connection settings

### 5. Frontend Protocol Support
- **File**: `web/assets/js/node-config/form.js`, `web/assets/js/node-config/utils.js`
- **Change**: Added authentication headers and protocol mapping
- **Impact**: UI properly handles protocol switching with correct encryption options

## 🚨 Required Changes in Arch-Node

### Priority 1: Critical Configuration Validation

#### 1.1 Fix Server Struct Validation
- **File**: `pkg/xray/config.go`
- **Current Issue**: 
  ```go
  type Server struct {
      Method string `json:"method" validate:"required"` // ❌ Shadowsocks-only
  }
  ```
- **Required Fix**:
  ```go
  type Server struct {
      Method   string `json:"method,omitempty"`   // Only for Shadowsocks
      Security string `json:"security,omitempty"` // For VMess
      AlterId  int    `json:"alterId,omitempty"`  // For VMess (legacy)
      Level    int    `json:"level,omitempty"`    // For VMess/VLESS
      // Add other protocol-specific fields as needed
  }
  ```
- **Status**: ✅ **COMPLETED** - Validation error fixed, arch-node now starts

#### 1.2 New Issue: VMess Outbound Structure Incorrect
- **Error**: `Failed to start: main: failed to load config files: [storage/app/xray.json] > infra/conf: failed to build outbound config with tag internal > infra/conf: failed to build outbound handler for protocol vmess > infra/conf: 0 VMess receiver configured`
- **Root Cause**: Arch-node package's `MakeVmessOutbound` generates wrong structure
- **Current Output**:
  ```json
  {
    "settings": {
      "servers": [{"address": "...", "id": "...", "method": "auto"}]
    }
  }
  ```
- **Expected VMess Structure**:
  ```json
  {
    "settings": {
      "vnext": [
        {
          "address": "...",
          "port": 123,
          "users": [{"id": "...", "security": "auto"}]
        }
      ]
    }
  }
  ```
- **Location**: Arch-node package `MakeVmessOutbound` function bug
- **Workaround**: Use Shadowsocks for internal communication, VMess only for client connections
- **Status**: ❌ **PENDING** - Requires arch-node package fix

#### 1.3 Previous Issue: Shadowsocks Outbound Password Missing
- **Error**: `Failed to start: main: failed to load config files: [storage/app/xray.json] > infra/conf: failed to build outbound config with tag internal > infra/conf: failed to build outbound handler for protocol shadowsocks > infra/conf: Shadowsocks password is not specified.`
- **Root Cause**: Arch-manager is sending Shadowsocks outbound config to arch-node but missing password field
- **Location**: `internal/writer/writer.go` in `RemoteConfig()` function
- **Issue**: The `internal` outbound is using Shadowsocks protocol but missing password field
- **Status**: ❌ **PENDING INVESTIGATION**

#### 1.3 Protocol-Aware Validation Logic
- **File**: `pkg/xray/config.go`
- **Task**: Implement conditional validation based on protocol type
- **Details**:
  - Shadowsocks: Require `Method` field
  - VMess: Require `Security`, optional `AlterId`
  - VLESS: Require `Security` for TLS/Reality
  - Trojan: Require `Password` field
- **Status**: ❌ **PENDING**

### Priority 2: Configuration Processing

#### 2.1 Update Configuration Loading
- **File**: `pkg/xray/xray.go`
- **Function**: `loadConfig()`
- **Task**: Handle different protocol structures during config loading
- **Details**:
  - Add protocol detection logic
  - Handle VMess-specific field validation
  - Maintain backward compatibility with Shadowsocks
- **Status**: ❌ **PENDING**

#### 2.2 Outbound Configuration Processing
- **File**: `pkg/xray/xray.go`
- **Task**: Update outbound server configuration for multi-protocol support
- **Details**:
  - Protocol-specific field mapping
  - Graceful handling of missing protocol-specific fields
  - Error messages with protocol context
- **Status**: ❌ **PENDING**

### Priority 3: Application Layer

#### 3.1 Startup Validation
- **File**: `internal/app/app.go`
- **Function**: `Start()`
- **Task**: Add protocol validation before Xray initialization
- **Details**:
  - Validate configuration completeness per protocol
  - Provide clear error messages for missing fields
  - Graceful fallback strategies
- **Status**: ❌ **PENDING**

#### 3.2 Configuration Sync Robustness
- **File**: Configuration receiver logic
- **Task**: Handle arch-manager protocol configuration updates
- **Details**:
  - Validate received configurations before applying
  - Handle protocol migrations (Shadowsocks → VMess)
  - Log protocol changes for debugging
- **Status**: ❌ **PENDING**

## 🛠 Immediate Actions

### Quick Fix (For Testing)
1. **Remove validation requirement** from `Method` field in `pkg/xray/config.go`:
   ```go
   // Change from:
   Method string `json:"method" validate:"required"`
   // To:
   Method string `json:"method,omitempty"`
   ```
2. **Test VMess configuration loading** without validation errors

### Development Approach
1. **Phase 1**: Remove blocking validation (immediate)
2. **Phase 2**: Implement protocol-aware validation (short-term)
3. **Phase 3**: Add comprehensive multi-protocol support (medium-term)
4. **Phase 4**: Optimize and test all protocol combinations (long-term)

## 🔄 Testing Strategy

### Test Cases to Implement
- [ ] Shadowsocks configuration (backward compatibility)
- [ ] VMess configuration with auto encryption
- [ ] VMess configuration with specific encryption
- [ ] VLESS configuration with TLS
- [ ] VLESS configuration with Reality
- [ ] Trojan configuration
- [ ] Protocol switching scenarios
- [ ] Invalid configuration handling

### Validation Points
- [ ] Configuration parsing without validation errors
- [ ] Xray process starts successfully with each protocol
- [ ] Network connectivity works for each protocol
- [ ] Manager-to-node sync works for all protocols
- [ ] Protocol changes propagate correctly

## 📋 Dependencies

### Arch-Node Dependencies
- Review `github.com/ebadidev/arch-node` package version compatibility
- Ensure Xray-core supports all target protocols
- Validate struct tags and validation library compatibility

### Integration Points
- Manager-to-node configuration sync mechanism
- Xray configuration file generation and validation
- Protocol-specific credential management (UUIDs vs keys)

## 🎯 Success Criteria

- [ ] Arch-node starts successfully with VMess configurations from arch-manager
- [ ] All protocols (Shadowsocks, VMess, VLESS, Trojan) work end-to-end
- [ ] Protocol switching in manager UI reflects correctly in node configurations
- [ ] Backward compatibility maintained for existing Shadowsocks setups
- [ ] Clear error messages for configuration issues
- [ ] Comprehensive test coverage for all protocol scenarios

---

**Last Updated**: August 22, 2025  
**Priority**: **HIGH** - Blocking multi-protocol functionality  
**Estimated Effort**: 2-3 days for full implementation  
**Dependencies**: arch-node codebase access and testing environment
