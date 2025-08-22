package v1

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/ebadidev/arch-manager/internal/coordinator"
	"github.com/ebadidev/arch-manager/internal/database"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/curve25519"
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
func NodeConfigUpdate(coordinator *coordinator.Coordinator, d *database.Database) echo.HandlerFunc {
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
				// Preserve existing basic node connection fields
				config.Id = node.Id
				config.Host = node.Host
				config.HttpToken = node.HttpToken
				config.HttpPort = node.HttpPort
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
				
				// Trigger configuration sync
				go coordinator.SyncConfigs()
				
				return c.JSON(http.StatusOK, config)
			}
		}
		
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Node not found",
		})
	}
}

// NodeConfigCreate creates a new node with full configuration
func NodeConfigCreate(coordinator *coordinator.Coordinator, d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var config database.Node
		if err := c.Bind(&config); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		
		d.Locker.Lock()
		defer d.Locker.Unlock()
		
		// Check node limit
		if len(d.Content.Nodes) > 5 {
			return c.JSON(http.StatusForbidden, map[string]string{
				"message": "Cannot add more nodes!",
			})
		}
		
		// Generate new node ID
		config.Id = d.GenerateNodeId()
		
		// Set default status
		config.PushStatus = database.NodeStatusProcessing
		config.PullStatus = database.NodeStatusProcessing
		
		// Add node to database
		d.Content.Nodes = append(d.Content.Nodes, &config)
		
		if err := d.Save(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to save configuration",
			})
		}
		
		// Trigger configuration sync
		go coordinator.SyncConfigs()
		
		return c.JSON(http.StatusCreated, config)
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
		// Generate a new private key
		privateKey := make([]byte, curve25519.ScalarSize)
		if _, err := rand.Read(privateKey); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate random private key: " + err.Error(),
			})
		}

		// Apply curve25519 clamping
		privateKey[0] &= 248
		privateKey[31] &= 127
		privateKey[31] |= 64

		// Generate public key from private key
		publicKey, err := curve25519.X25519(privateKey, curve25519.Basepoint)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate public key: " + err.Error(),
			})
		}

		// Generate random shortids (typically 1-5 shortids)
		shortIdCount := 3 // Generate 3 shortids by default
		shortIds := make([]string, shortIdCount)
		for i := 0; i < shortIdCount; i++ {
			shortIdBytes := make([]byte, 8)
			if _, err := rand.Read(shortIdBytes); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to generate shortid: " + err.Error(),
				})
			}
			shortIds[i] = base64.RawURLEncoding.EncodeToString(shortIdBytes)
		}

		// Encode keys to base64
		privateKeyB64 := base64.RawURLEncoding.EncodeToString(privateKey)
		publicKeyB64 := base64.RawURLEncoding.EncodeToString(publicKey)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"private_key": privateKeyB64,
			"public_key":  publicKeyB64,
			"short_ids":   shortIds,
		})
	}
}
