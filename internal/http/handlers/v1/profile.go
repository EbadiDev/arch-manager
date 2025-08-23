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
	
	// Add transport-specific parameters
	if node.NetworkSettings.Transport == "ws" && node.NetworkSettings.Settings != nil {
		if path, exists := node.NetworkSettings.Settings["path"]; exists {
			params = append(params, fmt.Sprintf("path=%s", path))
		}
		if host, exists := node.NetworkSettings.Settings["host"]; exists {
			params = append(params, fmt.Sprintf("host=%s", host))
		}
	}
	
	return fmt.Sprintf("%s?%s#%s", baseURL, joinParams(params), node.ServerName)
}

func generateTrojanLink(node *database.Node, user *database.User, settings *database.Settings) string {
	// Trojan link format: trojan://password@host:port?params#name
	password := generateTrojanPassword(user)
	baseURL := fmt.Sprintf("trojan://%s@%s:%d", password, settings.Host, node.ListeningPort)
	
	params := []string{
		fmt.Sprintf("type=%s", node.NetworkSettings.Transport),
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
