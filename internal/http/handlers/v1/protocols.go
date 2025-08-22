package v1

import (
	"net/http"
	"strconv"

	"github.com/ebadidev/arch-manager/internal/database"
	"github.com/labstack/echo/v4"
)

// ProtocolsList returns available protocols and their configuration options
func ProtocolsList(d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"protocols": []string{"shadowsocks", "vmess", "vless", "trojan"},
			"transports": []string{"tcp", "http", "ws", "grpc", "kcp", "httpupgrade", "xhttp"},
			"securities": []string{"none", "tls", "reality"},
			"core_types": []string{"xray"},
			"cert_modes": []string{"http", "file", "dns", "none"},
			"encryption_options": d.Content.Settings.EncryptionOptions,
		})
	}
}

// NodeConfigGet returns configuration for a specific node
func NodeConfigGet(d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		nodeId := c.Param("id")
		
		d.Locker.Lock()
		defer d.Locker.Unlock()
		
		for _, node := range d.Content.Nodes {
			if node.Id == parseNodeId(nodeId) {
				return c.JSON(http.StatusOK, node)
			}
		}
		
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Node not found",
		})
	}
}

// NodeConfigUpdate updates configuration for a specific node
func NodeConfigUpdate(d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		nodeId := c.Param("id")
		
		var config database.Node
		if err := c.Bind(&config); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		
		d.Locker.Lock()
		defer d.Locker.Unlock()
		
		for i, node := range d.Content.Nodes {
			if node.Id == parseNodeId(nodeId) {
				// Preserve existing fields
				config.Id = node.Id
				config.Usage = node.Usage
				config.UsageBytes = node.UsageBytes
				config.PushStatus = node.PushStatus
				config.PullStatus = node.PullStatus
				config.PushedAt = node.PushedAt
				config.PulledAt = node.PulledAt
				
				d.Content.Nodes[i] = &config
				
				if err := d.Save(); err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{
						"message": "Failed to save configuration",
					})
				}
				
				return c.JSON(http.StatusOK, config)
			}
		}
		
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Node not found",
		})
	}
}

// Helper function to parse node ID
func parseNodeId(nodeIdStr string) int {
	if id, err := strconv.Atoi(nodeIdStr); err == nil {
		return id
	}
	return 0
}

// GenerateRealityKeys generates a new X25519 key pair for Reality configuration
func GenerateRealityKeys(d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		// For now, return sample keys
		// TODO: Implement actual X25519 key generation
		return c.JSON(http.StatusOK, map[string]string{
			"private_key": "yBaw532IIUNuQWDTncozoBaLJmcd1JZzvsHUgVPxMk8",
			"public_key":  "7xhH4b_VkliBxGulljcyPOH-bYUA2dl-XAdZAsfhk04",
		})
	}
}
