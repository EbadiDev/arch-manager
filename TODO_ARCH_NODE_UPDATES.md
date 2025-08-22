# TODO: Arch-Node Package Updates & Transport Configuration Enhancement

## Current Status (August 22, 2025)

### ✅ Completed Phase 1 & 2 (Multi-Protocol Foundation)
- **Database Schema** - Complete Node struct with protocol fields
- **Protocol Factory** - Extensible switch pattern in writer.go
- **Web Interface** - Dynamic multi-protocol configuration forms
- **API Endpoints** - Protocol management and node config APIs
- **Authentication** - Fixed auth headers in web interface

### 🎯 Phase 3: External Package & Advanced Configuration

## 1. External Package Verification (Critical Priority)

### Check arch-node Package Methods
The current package `github.com/ebadidev/arch-node v0.0.0-20250821092106-159744a8dedc` needs verification for:

```go
// Required methods to check/implement:
func MakeVlessInbound(id int, port int, password string, network NetworkConfig, security interface{}) (interface{}, error)
func MakeVmessInbound(id int, port int, password string, encryption string, network NetworkConfig) (interface{}, error) 
func MakeTrojanInbound(id int, port int, password string, network NetworkConfig, security interface{}) (interface{}, error)
func MakeVlessOutbound(config OutboundConfig) (interface{}, error)
func MakeVmessOutbound(config OutboundConfig) (interface{}, error)
func MakeTrojanOutbound(config OutboundConfig) (interface{}, error)
```

### Action Items:
1. **Inspect arch-node source** - Check what methods exist
2. **Fork if needed** - Create extended version with missing methods
3. **Update dependency** - Point to extended package
4. **Test integration** - Verify all protocols work end-to-end

## 2. Transport Configuration Enhancement (In Progress)

### Current Limitation
- Web interface only allows transport type selection
- No way to configure transport-specific settings
- Missing JSON editor for advanced transport configurations

### Required Enhancement: JSON Transport Settings Editor

#### Updated Web Interface (`web/admin-node-config.html`)
```html
<!-- Enhanced Network Settings Section -->
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
        
        <!-- NEW: JSON Transport Settings Editor -->
        <div class="mt-3">
            <label class="form-label">Transport Settings (JSON)</label>
            <div class="d-flex justify-content-between align-items-center mb-2">
                <small class="text-muted">Configure transport-specific settings</small>
                <div>
                    <button type="button" class="btn btn-sm btn-outline-primary" id="formatJsonBtn">Format JSON</button>
                    <button type="button" class="btn btn-sm btn-outline-info" id="loadTemplateBtn">Load Template</button>
                </div>
            </div>
            <textarea 
                name="network_settings" 
                id="networkSettingsEditor" 
                class="form-control font-monospace" 
                rows="8" 
                placeholder='{"path": "/", "host": "example.com"}'
                style="resize: vertical; font-size: 0.9em;"
            ></textarea>
            <div class="form-text">
                <span id="jsonValidationStatus" class="text-muted">Valid JSON</span>
                <span class="ms-2">|</span>
                <a href="#" id="showExamplesLink" class="ms-2">Show Examples</a>
            </div>
        </div>
        
        <!-- JSON Examples Modal Trigger -->
        <div class="collapse mt-2" id="jsonExamples">
            <div class="card card-body bg-light">
                <h6>Transport Configuration Examples:</h6>
                <div id="transportExamples">
                    <!-- Examples populated by JavaScript -->
                </div>
            </div>
        </div>
    </div>
</div>
```

#### Enhanced JavaScript (`web/assets/js/node-config.js`)
```javascript
// JSON Editor Enhancement
class TransportConfigEditor {
    constructor() {
        this.editor = document.getElementById('networkSettingsEditor');
        this.statusElement = document.getElementById('jsonValidationStatus');
        this.templates = this.getTransportTemplates();
        this.initializeEditor();
    }
    
    initializeEditor() {
        // Real-time JSON validation
        this.editor.addEventListener('input', () => this.validateJSON());
        
        // Format JSON button
        document.getElementById('formatJsonBtn').addEventListener('click', () => this.formatJSON());
        
        // Load template button  
        document.getElementById('loadTemplateBtn').addEventListener('click', () => this.showTemplateModal());
        
        // Transport change handler
        document.getElementById('transportSelect').addEventListener('change', (e) => {
            this.loadTemplate(e.target.value);
        });
        
        // Examples toggle
        document.getElementById('showExamplesLink').addEventListener('click', (e) => {
            e.preventDefault();
            this.toggleExamples();
        });
    }
    
    validateJSON() {
        try {
            const value = this.editor.value.trim();
            if (value === '') {
                this.setStatus('Empty (will use defaults)', 'text-muted');
                return true;
            }
            
            JSON.parse(value);
            this.setStatus('✓ Valid JSON', 'text-success');
            return true;
        } catch (error) {
            this.setStatus('✗ Invalid JSON: ' + error.message, 'text-danger');
            return false;
        }
    }
    
    formatJSON() {
        try {
            const value = this.editor.value.trim();
            if (value === '') return;
            
            const parsed = JSON.parse(value);
            const formatted = JSON.stringify(parsed, null, 2);
            this.editor.value = formatted;
            this.setStatus('✓ Formatted', 'text-success');
        } catch (error) {
            this.setStatus('✗ Cannot format invalid JSON', 'text-danger');
        }
    }
    
    loadTemplate(transport) {
        const template = this.templates[transport];
        if (template) {
            this.editor.value = JSON.stringify(template, null, 2);
            this.validateJSON();
        }
    }
    
    getTransportTemplates() {
        return {
            tcp: {
                header: {
                    type: "none"
                }
            },
            http: {
                path: "/",
                host: ["example.com"],
                method: "GET",
                headers: {
                    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
                }
            },
            ws: {
                path: "/",
                headers: {
                    "Host": "example.com"
                }
            },
            grpc: {
                serviceName: "GunService",
                multiMode: false
            },
            kcp: {
                mtu: 1350,
                tti: 50,
                uplinkCapacity: 5,
                downlinkCapacity: 20,
                congestion: false,
                readBufferSize: 2,
                writeBufferSize: 2,
                header: {
                    type: "none"
                }
            },
            httpupgrade: {
                path: "/",
                host: "example.com"
            },
            xhttp: {
                path: "/",
                host: ["example.com"]
            }
        };
    }
    
    setStatus(message, className) {
        this.statusElement.textContent = message;
        this.statusElement.className = className;
    }
    
    toggleExamples() {
        const examples = document.getElementById('jsonExamples');
        const collapse = new bootstrap.Collapse(examples);
        
        if (!examples.querySelector('.populated')) {
            this.populateExamples();
        }
    }
    
    populateExamples() {
        const container = document.getElementById('transportExamples');
        const examples = this.templates;
        
        Object.keys(examples).forEach(transport => {
            const example = document.createElement('div');
            example.className = 'mb-2';
            example.innerHTML = `
                <strong>${transport.toUpperCase()}:</strong>
                <pre class="bg-white p-2 border rounded mt-1" style="font-size: 0.8em;"><code>${JSON.stringify(examples[transport], null, 2)}</code></pre>
            `;
            container.appendChild(example);
        });
        
        container.classList.add('populated');
    }
}
```

## 3. Complete Writer Logic Implementation

### Update `internal/writer/writer.go`
Once external package is verified/updated:

```go
func (w *Writer) makeProtocolInbound(node *database.Node, port int, password string) (interface{}, error) {
    switch node.Protocol {
    case "shadowsocks":
        return xc.MakeShadowsocksInbound(node.Id, node.ListeningPort, password, node.Encryption)
        
    case "vless":
        var securityConfig interface{}
        if node.Security == "reality" && node.SecuritySettings.Reality != nil {
            securityConfig = node.SecuritySettings.Reality
        } else if node.Security == "tls" && node.SecuritySettings.TLS != nil {
            securityConfig = node.SecuritySettings.TLS
        }
        return xc.MakeVlessInbound(node.Id, node.ListeningPort, password, node.NetworkSettings, securityConfig)
        
    case "vmess":
        return xc.MakeVmessInbound(node.Id, node.ListeningPort, password, node.Encryption, node.NetworkSettings)
        
    case "trojan":
        return xc.MakeTrojanInbound(node.Id, node.ListeningPort, password, node.NetworkSettings, node.SecuritySettings)
        
    default:
        return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
    }
}
```

## 4. Remove Legacy Shadowsocks Settings

### Update `web/admin-system.html`
Remove hardcoded SS port configurations:
- SS Direct Port field
- SS Relay Port field  
- SS Ports section entirely

### Keep Only General Settings:
- Admin Password
- Traffic Ratio
- Singet Server
- Reset Policy

## 5. Advanced Features Implementation

### Reality Key Generation
```go
// internal/http/handlers/v1/protocols.go
func GenerateRealityKeys() echo.HandlerFunc {
    return func(c echo.Context) error {
        privateKey, publicKey, err := generateX25519KeyPair()
        if err != nil {
            return c.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to generate keys"
            })
        }
        
        return c.JSON(http.StatusOK, map[string]string{
            "private_key": privateKey,
            "public_key":  publicKey,
        })
    }
}
```

### Enhanced Validation
- JSON schema validation for transport settings
- Protocol-specific field validation
- Real-time form validation feedback

## 6. Testing & Quality Assurance

### Unit Tests Needed
- [ ] Transport JSON parsing and validation
- [ ] Protocol factory method testing
- [ ] Configuration generation per protocol
- [ ] Reality key generation
- [ ] Form validation edge cases

### Integration Tests
- [ ] End-to-end node creation with JSON transport configs
- [ ] Multi-protocol configuration push/pull
- [ ] Web interface JSON editor functionality
- [ ] API endpoint validation

### Manual Testing Scenarios
1. **JSON Transport Editing** - Test all transport types with complex configs
2. **JSON Validation** - Test invalid JSON handling and error messages
3. **Template Loading** - Verify templates work for each transport type
4. **Format Function** - Test JSON beautification
5. **Form Integration** - Ensure JSON settings integrate with form submission

## Implementation Priority

### Immediate (Next Sprint)
1. **✅ JSON Transport Editor** - Enhanced web interface with JSON editing
2. **⏳ External Package Check** - Verify arch-node protocol methods
3. **⏳ Legacy Settings Cleanup** - Remove SS-specific hardcoded settings

### Short Term (1-2 weeks)
4. **Protocol Method Implementation** - Complete writer.go protocol factory
5. **Reality Key Generation** - Add cryptographic key generation
6. **Comprehensive Testing** - Unit and integration test coverage

### Medium Term (1 month)
7. **Performance Optimization** - Efficient config generation
8. **Documentation** - API documentation and user guides
9. **Sing-box Preparation** - Foundation for future Sing-box support

## Success Criteria

### Enhanced Transport Configuration ✅
- [x] JSON editor with syntax highlighting
- [x] Real-time validation and formatting
- [x] Transport-specific templates
- [x] Example configurations
- [x] Integration with form submission

### Multi-Protocol Support ⏳
- [ ] All protocols generate correct Xray configs
- [ ] Transport settings properly applied
- [ ] Security configurations work end-to-end
- [ ] Reality key generation functional

### Code Quality ⏳
- [ ] Clean protocol abstraction
- [ ] Comprehensive error handling  
- [ ] Full test coverage
- [ ] Performance optimized

## Notes

### Current Blockers
1. **External Package** - Need to verify/extend arch-node methods
2. **Legacy Cleanup** - Remove hardcoded SS settings from admin interface

### Development Approach
1. **Transport Enhancement** - Implement JSON editor first (immediate value)
2. **External Package** - Investigate and extend as needed
3. **Integration** - Connect all components end-to-end
4. **Polish** - Add advanced features and testing

### Future Considerations
- **Sing-box Support** - Design allows easy addition
- **Custom Protocols** - Extensible factory pattern ready
- **Advanced Xray Features** - Framework supports future additions
