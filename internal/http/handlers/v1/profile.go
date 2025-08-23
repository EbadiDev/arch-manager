package v1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/ebadidev/arch-manager/internal/coordinator"
	"github.com/ebadidev/arch-manager/internal/database"
	"github.com/labstack/echo/v4"
)

type ProfileResponse struct {
	User        database.User           `json:"user"`
	Connections []ConnectionInfo        `json:"connections"`
}

type ConnectionInfo struct {
	Type        string `json:"type"`        // "direct", "relay", "reverse", "remote"
	Protocol    string `json:"protocol"`    // "shadowsocks", "vmess", "vless", "trojan"
	Transport   string `json:"transport"`   // "tcp", "ws", "grpc", "http", etc.
	Name        string `json:"name"`        // Display name
	Link        string `json:"link"`        // Connection URL
	Port        int    `json:"port"`        // Connection port
}

func ProfileShow(d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var user *database.User
		for _, u := range d.Content.Users {
			if u.Identity == c.QueryParam("u") {
				user = u
			}
		}
		if user == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		r := ProfileResponse{User: *user}
		r.User.Usage = r.User.Usage * d.Content.Settings.TrafficRatio
		r.User.Quota = r.User.Quota * d.Content.Settings.TrafficRatio

		// Generate connection info based on actual configurations
		r.Connections = generateConnectionInfo(d, user)

		return c.JSON(http.StatusOK, r)
	}
}

func generateConnectionInfo(d *database.Database, user *database.User) []ConnectionInfo {
	var connections []ConnectionInfo
	s := d.Content.Settings

	// Only generate connections from actual configured nodes
	// Internal Shadowsocks connections are hidden from users
	for _, node := range d.Content.Nodes {
		if node.Protocol == "" || node.ServerPort == "" {
			continue
		}
		
		// Skip internal Shadowsocks connections (used for manager-node communication)
		// Users should only see their configured client-facing protocols
		if isInternalConnection(node) {
			continue
		}

		// Create connection info based on node configuration
		connInfo := ConnectionInfo{
			Type:      "remote",
			Protocol:  node.Protocol,
			Transport: node.NetworkSettings.Transport,
			Port:      node.ListeningPort,
		}

		// Set display name
		if node.NetworkSettings.Transport != "" && node.NetworkSettings.Transport != "tcp" {
			connInfo.Name = fmt.Sprintf("%s (%s)", 
				formatProtocolName(node.Protocol), 
				formatTransportName(node.NetworkSettings.Transport))
		} else {
			connInfo.Name = formatProtocolName(node.Protocol)
		}

		// Generate appropriate connection link
		switch node.Protocol {
		case "vmess":
			connInfo.Link = generateVMessLink(node, user, s)
		case "vless":
			connInfo.Link = generateVLESSLink(node, user, s)
		case "trojan":
			connInfo.Link = generateTrojanLink(node, user, s)
		case "shadowsocks":
			connInfo.Link = generateShadowsocksLink(node, user, s)
		}

		if connInfo.Link != "" {
			connections = append(connections, connInfo)
		}
	}

	return connections
}

// isInternalConnection checks if a node connection is for internal manager-node communication
func isInternalConnection(node *database.Node) bool {
	// Internal connections are typically:
	// 1. Shadowsocks protocol used purely for infrastructure
	// 2. Connections with no user-facing server configuration
	// 3. Nodes that are only used for relay/reverse tunneling
	
	// For now, we show all properly configured nodes to users
	// Internal Shadowsocks inbounds are created in writer.go but not stored as user nodes
	return false
}

func formatProtocolName(protocol string) string {
	switch protocol {
	case "vmess":
		return "VMess"
	case "vless":
		return "VLESS"
	case "trojan":
		return "Trojan"
	case "shadowsocks":
		return "Shadowsocks"
	default:
		return protocol
	}
}

func formatTransportName(transport string) string {
	switch transport {
	case "ws":
		return "WebSocket"
	case "grpc":
		return "gRPC"
	case "http":
		return "HTTP"
	case "tcp":
		return "TCP"
	default:
		return transport
	}
}

func generateVMessLink(node *database.Node, user *database.User, settings *database.Settings) string {
	// VMess link format: vmess://base64(json_config)
	config := map[string]interface{}{
		"v":    "2",
		"ps":   node.ServerName,
		"add":  settings.Host,
		"port": node.ListeningPort,
		"id":   generateUserUUID(user), // Generate UUID based on user
		"aid":  "0",
		"scy":  node.Encryption,
		"net":  node.NetworkSettings.Transport,
		"type": "none",
		"host": "",
		"path": "",
		"tls":  "",
	}

	// Set security (TLS) field based on node security configuration
	switch node.Security {
	case "tls":
		config["tls"] = "tls"
		// Add TLS-specific settings if available
		if node.SecuritySettings.TLS != nil {
			if node.SecuritySettings.TLS.SNI != "" {
				config["sni"] = node.SecuritySettings.TLS.SNI
			} else if node.SecuritySettings.TLS.ServerName != "" {
				config["sni"] = node.SecuritySettings.TLS.ServerName // fallback to server_name
			}
			
			// Add fingerprint if configured
			if node.SecuritySettings.TLS.Fingerprint != "" {
				config["fp"] = node.SecuritySettings.TLS.Fingerprint
			}
			
			// Add ALPN if configured
			if len(node.SecuritySettings.TLS.ALPN) > 0 {
				// Join ALPN protocols with comma
				alpnStr := ""
				for i, protocol := range node.SecuritySettings.TLS.ALPN {
					if i > 0 {
						alpnStr += ","
					}
					alpnStr += protocol
				}
				config["alpn"] = alpnStr
			}
		}
	case "none":
		config["tls"] = ""
	default:
		// VMess doesn't support Reality or other security types
		// Default to no TLS for unsupported security types
		config["tls"] = ""
	}

	// Add transport-specific settings
	if node.NetworkSettings.Settings != nil {
		switch node.NetworkSettings.Transport {
		case "ws":
			// WebSocket transport settings
			if path, exists := node.NetworkSettings.Settings["path"]; exists {
				config["path"] = path
			}
			if host, exists := node.NetworkSettings.Settings["host"]; exists {
				config["host"] = host
			}
		case "kcp":
			// KCP transport settings
			if header, exists := node.NetworkSettings.Settings["header"]; exists {
				if headerMap, ok := header.(map[string]interface{}); ok {
					if headerType, exists := headerMap["type"]; exists {
						config["type"] = headerType // KCP header type (none, srtp, utp, etc.)
					}
					if domain, exists := headerMap["domain"]; exists {
						config["host"] = domain // KCP domain goes in host field
					}
				}
			}
			if seed, exists := node.NetworkSettings.Settings["seed"]; exists {
				config["path"] = seed // KCP seed goes in path field
			}
		case "grpc":
			// gRPC transport settings
			if serviceName, exists := node.NetworkSettings.Settings["serviceName"]; exists {
				config["path"] = serviceName // gRPC service name goes in path
			}
			if authority, exists := node.NetworkSettings.Settings["authority"]; exists {
				config["host"] = authority // gRPC authority goes in host
			}
		case "http":
			// HTTP transport settings
			config["net"] = "tcp" // HTTP transport uses TCP network
			config["type"] = "http" // Set header type to http
			if path, exists := node.NetworkSettings.Settings["path"]; exists {
				config["path"] = path
			}
			if host, exists := node.NetworkSettings.Settings["host"]; exists {
				if hosts, ok := host.([]interface{}); ok && len(hosts) > 0 {
					// Use first host from array
					config["host"] = hosts[0]
				} else {
					config["host"] = host
				}
			}
		}
	}

	// Convert to JSON and encode
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		return ""
	}
	
	return "vmess://" + base64.StdEncoding.EncodeToString(jsonBytes)
}

func generateVLESSLink(node *database.Node, user *database.User, settings *database.Settings) string {
	// VLESS link format: vless://uuid@host:port?params#name
	uuid := generateUserUUID(user)
	baseURL := fmt.Sprintf("vless://%s@%s:%d", uuid, settings.Host, node.ListeningPort)
	
	params := []string{
		"encryption=none",
		fmt.Sprintf("type=%s", node.NetworkSettings.Transport),
	}
	
	// Add security settings
	switch node.Security {
	case "tls":
		params = append(params, "security=tls")
		
		if node.SecuritySettings.TLS != nil {
			// Add SNI
			if node.SecuritySettings.TLS.SNI != "" {
				params = append(params, fmt.Sprintf("sni=%s", node.SecuritySettings.TLS.SNI))
			} else if node.SecuritySettings.TLS.ServerName != "" {
				params = append(params, fmt.Sprintf("sni=%s", node.SecuritySettings.TLS.ServerName))
			}
			
			// Add fingerprint
			if node.SecuritySettings.TLS.Fingerprint != "" {
				params = append(params, fmt.Sprintf("fp=%s", node.SecuritySettings.TLS.Fingerprint))
			}
			
			// Add ALPN
			if len(node.SecuritySettings.TLS.ALPN) > 0 {
				alpnStr := ""
				for i, protocol := range node.SecuritySettings.TLS.ALPN {
					if i > 0 {
						alpnStr += ","
					}
					alpnStr += protocol
				}
				params = append(params, fmt.Sprintf("alpn=%s", alpnStr))
			}
			
			// Add allowInsecure
			if node.SecuritySettings.TLS.AllowInsecure {
				params = append(params, "allowInsecure=1")
			}
		}
	case "reality":
		params = append(params, "security=reality")
		
		if node.SecuritySettings.Reality != nil {
			// Add Reality fingerprint (most commonly configured)
			if node.SecuritySettings.Reality.Fingerprint != "" {
				params = append(params, fmt.Sprintf("fp=%s", node.SecuritySettings.Reality.Fingerprint))
			}
			
			// Add Reality server names (SNI)
			if len(node.SecuritySettings.Reality.ServerNames) > 0 {
				params = append(params, fmt.Sprintf("sni=%s", node.SecuritySettings.Reality.ServerNames[0]))
			}
			
			// Add Reality public key (critical for connection)
			if node.SecuritySettings.Reality.PublicKey != "" {
				params = append(params, fmt.Sprintf("pbk=%s", node.SecuritySettings.Reality.PublicKey))
			}
			
			// Add Reality short ID (use first one if multiple)
			if len(node.SecuritySettings.Reality.ShortIDs) > 0 {
				params = append(params, fmt.Sprintf("sid=%s", node.SecuritySettings.Reality.ShortIDs[0]))
			}
			
			// Add Reality spider X (spx) if configured
			if node.SecuritySettings.Reality.SpiderX != "" {
				params = append(params, fmt.Sprintf("spx=%s", node.SecuritySettings.Reality.SpiderX))
			}
		}
	case "none":
		// No security parameters needed
	default:
		// Invalid security type - should not happen due to validation
		// But handle gracefully by defaulting to no security
	}
	
	// Add transport-specific parameters
	if node.NetworkSettings.Settings != nil {
		switch node.NetworkSettings.Transport {
		case "ws":
			if path, exists := node.NetworkSettings.Settings["path"]; exists {
				params = append(params, fmt.Sprintf("path=%s", path))
			}
			if host, exists := node.NetworkSettings.Settings["host"]; exists {
				params = append(params, fmt.Sprintf("host=%s", host))
			}
		case "grpc":
			if serviceName, exists := node.NetworkSettings.Settings["serviceName"]; exists {
				params = append(params, fmt.Sprintf("serviceName=%s", serviceName))
			}
			if authority, exists := node.NetworkSettings.Settings["authority"]; exists {
				params = append(params, fmt.Sprintf("authority=%s", authority))
			}
		case "http":
			// HTTP+TCP transport settings
			if path, exists := node.NetworkSettings.Settings["path"]; exists {
				params = append(params, fmt.Sprintf("path=%s", path))
			}
			if host, exists := node.NetworkSettings.Settings["host"]; exists {
				if hosts, ok := host.([]interface{}); ok && len(hosts) > 0 {
					params = append(params, fmt.Sprintf("host=%s", hosts[0]))
				} else {
					params = append(params, fmt.Sprintf("host=%s", host))
				}
			}
		case "xhttp":
			// XHTTP transport settings
			if path, exists := node.NetworkSettings.Settings["path"]; exists {
				params = append(params, fmt.Sprintf("path=%s", path))
			}
			if host, exists := node.NetworkSettings.Settings["host"]; exists {
				if hosts, ok := host.([]interface{}); ok && len(hosts) > 0 {
					params = append(params, fmt.Sprintf("host=%s", hosts[0]))
				} else {
					params = append(params, fmt.Sprintf("host=%s", host))
				}
			}
			// XHTTP-specific parameters
			if mode, exists := node.NetworkSettings.Settings["mode"]; exists {
				params = append(params, fmt.Sprintf("mode=%s", mode))
			}
			if customHost, exists := node.NetworkSettings.Settings["custom_host"]; exists {
				params = append(params, fmt.Sprintf("custom_host=%s", customHost))
			}
			if noGRPCHeader, exists := node.NetworkSettings.Settings["noGRPCHeader"]; exists {
				if noGRPC, ok := noGRPCHeader.(bool); ok && noGRPC {
					params = append(params, "noGRPCHeader=true")
				}
			}
			if noSSEHeader, exists := node.NetworkSettings.Settings["noSSEHeader"]; exists {
				if noSSE, ok := noSSEHeader.(bool); ok && noSSE {
					params = append(params, "noSSEHeader=true")
				}
			}
		case "kcp":
			// KCP transport settings
			if header, exists := node.NetworkSettings.Settings["header"]; exists {
				if headerMap, ok := header.(map[string]interface{}); ok {
					if headerType, exists := headerMap["type"]; exists {
						params = append(params, fmt.Sprintf("headerType=%s", headerType))
					}
					if domain, exists := headerMap["domain"]; exists {
						params = append(params, fmt.Sprintf("host=%s", domain))
					}
				}
			}
			if seed, exists := node.NetworkSettings.Settings["seed"]; exists {
				params = append(params, fmt.Sprintf("seed=%s", seed))
			}
		case "tcp":
			// TCP transport with optional HTTP header
			if header, exists := node.NetworkSettings.Settings["header"]; exists {
				if headerMap, ok := header.(map[string]interface{}); ok {
					if headerType, exists := headerMap["type"]; exists && headerType == "http" {
						// TCP with HTTP header obfuscation
						if path, exists := headerMap["path"]; exists {
							params = append(params, fmt.Sprintf("path=%s", path))
						}
						if host, exists := headerMap["host"]; exists {
							if hosts, ok := host.([]interface{}); ok && len(hosts) > 0 {
								params = append(params, fmt.Sprintf("host=%s", hosts[0]))
							} else {
								params = append(params, fmt.Sprintf("host=%s", host))
							}
						}
					}
				}
			}
		}
	}
	
	return fmt.Sprintf("%s?%s#%s", baseURL, joinParams(params), node.ServerName)
}

func generateTrojanLink(node *database.Node, user *database.User, settings *database.Settings) string {
	// Trojan requires TLS security - validate using tagged switch
	switch node.Security {
	case "tls":
		// Valid - continue with link generation
	default:
		// Invalid security for Trojan - return empty link
		return ""
	}
	
	// Trojan link format: trojan://password@host:port?params#name
	password := generateTrojanPassword(user)
	baseURL := fmt.Sprintf("trojan://%s@%s:%d", password, settings.Host, node.ListeningPort)
	
	params := []string{
		fmt.Sprintf("type=%s", node.NetworkSettings.Transport),
		"security=tls", // Always TLS for Trojan
	}
	
	// Add TLS settings (required for Trojan)
	if node.SecuritySettings.TLS != nil {
		// Add SNI
		if node.SecuritySettings.TLS.SNI != "" {
			params = append(params, fmt.Sprintf("sni=%s", node.SecuritySettings.TLS.SNI))
		} else if node.SecuritySettings.TLS.ServerName != "" {
			params = append(params, fmt.Sprintf("sni=%s", node.SecuritySettings.TLS.ServerName))
		}
		
		// Add fingerprint
		if node.SecuritySettings.TLS.Fingerprint != "" {
			params = append(params, fmt.Sprintf("fp=%s", node.SecuritySettings.TLS.Fingerprint))
		}
		
		// Add ALPN
		if len(node.SecuritySettings.TLS.ALPN) > 0 {
			alpnStr := ""
			for i, protocol := range node.SecuritySettings.TLS.ALPN {
				if i > 0 {
					alpnStr += ","
				}
				alpnStr += protocol
			}
			params = append(params, fmt.Sprintf("alpn=%s", alpnStr))
		}
		
		// Add allowInsecure
		if node.SecuritySettings.TLS.AllowInsecure {
			params = append(params, "allowInsecure=1")
		}
	}
	
	// Add transport-specific parameters
	if node.NetworkSettings.Settings != nil {
		switch node.NetworkSettings.Transport {
		case "ws":
			if path, exists := node.NetworkSettings.Settings["path"]; exists {
				params = append(params, fmt.Sprintf("path=%s", path))
			}
			if host, exists := node.NetworkSettings.Settings["host"]; exists {
				params = append(params, fmt.Sprintf("host=%s", host))
			}
		case "grpc":
			if serviceName, exists := node.NetworkSettings.Settings["serviceName"]; exists {
				params = append(params, fmt.Sprintf("serviceName=%s", serviceName))
			}
			if authority, exists := node.NetworkSettings.Settings["authority"]; exists {
				params = append(params, fmt.Sprintf("authority=%s", authority))
			}
		case "http":
			// HTTP+TCP transport settings
			if path, exists := node.NetworkSettings.Settings["path"]; exists {
				params = append(params, fmt.Sprintf("path=%s", path))
			}
			if host, exists := node.NetworkSettings.Settings["host"]; exists {
				if hosts, ok := host.([]interface{}); ok && len(hosts) > 0 {
					params = append(params, fmt.Sprintf("host=%s", hosts[0]))
				} else {
					params = append(params, fmt.Sprintf("host=%s", host))
				}
			}
		case "kcp":
			// KCP transport settings
			if header, exists := node.NetworkSettings.Settings["header"]; exists {
				if headerMap, ok := header.(map[string]interface{}); ok {
					if headerType, exists := headerMap["type"]; exists {
						params = append(params, fmt.Sprintf("headerType=%s", headerType))
					}
					if domain, exists := headerMap["domain"]; exists {
						params = append(params, fmt.Sprintf("host=%s", domain))
					}
				}
			}
			if seed, exists := node.NetworkSettings.Settings["seed"]; exists {
				params = append(params, fmt.Sprintf("seed=%s", seed))
			}
		case "tcp":
			// TCP transport with optional HTTP header
			if header, exists := node.NetworkSettings.Settings["header"]; exists {
				if headerMap, ok := header.(map[string]interface{}); ok {
					if headerType, exists := headerMap["type"]; exists && headerType == "http" {
						// TCP with HTTP header obfuscation
						if path, exists := headerMap["path"]; exists {
							params = append(params, fmt.Sprintf("path=%s", path))
						}
						if host, exists := headerMap["host"]; exists {
							if hosts, ok := host.([]interface{}); ok && len(hosts) > 0 {
								params = append(params, fmt.Sprintf("host=%s", hosts[0]))
							} else {
								params = append(params, fmt.Sprintf("host=%s", host))
							}
						}
					}
				}
			}
		}
	}
	
	return fmt.Sprintf("%s?%s#%s", baseURL, joinParams(params), node.ServerName)
}

func generateShadowsocksLink(node *database.Node, user *database.User, settings *database.Settings) string {
	// Shadowsocks link format: ss://base64(method:password)@host:port#name
	// Use the node's encryption method, not the user's method
	auth := base64.StdEncoding.EncodeToString([]byte(node.Encryption + ":" + user.ShadowsocksPassword))
	return fmt.Sprintf("ss://%s@%s:%d#%s", auth, settings.Host, node.ListeningPort, node.ServerName)
}

func generateUserUUID(user *database.User) string {
	// Generate deterministic UUID based on user identity
	// This ensures the same user always gets the same UUID
	return fmt.Sprintf("%s-0000-0000-0000-000000000000", user.Identity[:8])
}

func generateTrojanPassword(user *database.User) string {
	// Generate Trojan password based on user identity
	return user.Identity
}

func joinParams(params []string) string {
	result := ""
	for i, param := range params {
		if i > 0 {
			result += "&"
		}
		result += param
	}
	return result
}

func ProfileRegenerate(coordinator *coordinator.Coordinator, d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		d.Locker.Lock()
		defer d.Locker.Unlock()

		var user *database.User
		for _, u := range d.Content.Users {
			if u.Identity == c.QueryParam("u") {
				user = u
			}
		}
		if user == nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Not found.",
			})
		}

		user.ShadowsocksPassword = d.GenerateUserPassword()

		if err := d.Save(); err != nil {
			return errors.WithStack(err)
		}

		go coordinator.SyncConfigs()

		return c.JSON(http.StatusOK, user)
	}
}
