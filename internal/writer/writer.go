package writer

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ebadidev/arch-manager/internal/config"
	"github.com/ebadidev/arch-manager/internal/database"
	"github.com/ebadidev/arch-manager/internal/http/client"
	"github.com/ebadidev/arch-manager/internal/utils"
	"github.com/ebadidev/arch-node/pkg/xray"
)

type Writer struct {
	c        *config.Config
	hc       *client.Client
	database *database.Database
	xray     *xray.Xray
}

func (w *Writer) clients() []*xray.Client {
	var clients []*xray.Client
	for _, u := range w.database.Content.Users {
		if !u.Enabled {
			continue
		}
		clients = append(clients, &xray.Client{
			Email:    strconv.Itoa(u.Id),
			Password: u.ShadowsocksPassword,
			Method:   u.ShadowsocksMethod,
		})
	}
	return clients
}

// Protocol factory method - supports all protocols via arch-node package
func (w *Writer) makeProtocolInbound(node *database.Node, tag, password, network string, port int, clients []*xray.Client) (*xray.Inbound, error) {
	xc := xray.NewConfig(w.c.Xray.LogLevel)
	
	switch node.Protocol {
	case "shadowsocks":
		// For Shadowsocks, generate a proper password if not provided
		if password == "" {
			var err error
			password, err = utils.Key32()
			if err != nil {
				return nil, err
			}
		}
		return xc.MakeShadowsocksInbound(tag, password, node.Encryption, network, port, clients), nil
	case "vless":
		// For VLESS, generate a proper UUID
		uuid := utils.UUID()
		var security interface{}
		if node.Security == "reality" && node.SecuritySettings.Reality != nil {
			security = node.SecuritySettings.Reality
		} else if node.Security == "tls" && node.SecuritySettings.TLS != nil {
			security = node.SecuritySettings.TLS
		}
		return xc.MakeVlessInbound(tag, port, uuid, network, security), nil
	case "vmess":
		// For VMess, generate a proper UUID
		uuid := utils.UUID()
		return xc.MakeVmessInbound(tag, port, uuid, node.Encryption, network), nil
	case "trojan":
		var security interface{}
		if node.Security == "tls" && node.SecuritySettings.TLS != nil {
			security = node.SecuritySettings.TLS
		}
		// For Trojan, use password as-is
		return xc.MakeTrojanInbound(tag, port, password, network, security), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}
}

// Shadowsocks inbound factory (uses existing arch-node method)
func (w *Writer) makeShadowsocksInbound(tag, password, method, network string, port int, clients []*xray.Client) *xray.Inbound {
	xc := xray.NewConfig(w.c.Xray.LogLevel)
	return xc.MakeShadowsocksInbound(tag, password, method, network, port, clients)
}

// Protocol outbound factory - supports all protocols via arch-node package
func (w *Writer) makeProtocolOutbound(node *database.Node, tag, host, password, method string, port int) (*xray.Outbound, error) {
	xc := xray.NewConfig(w.c.Xray.LogLevel)
	
	switch node.Protocol {
	case "shadowsocks":
		return xc.MakeShadowsocksOutbound(tag, host, password, method, port), nil
	case "vless":
		// For VLESS: tag, address, port, uuid, network
		network := "tcp" // Default network for outbound
		if node.NetworkSettings.Transport != "" {
			network = node.NetworkSettings.Transport
		}
		return xc.MakeVlessOutbound(tag, host, port, password, network), nil
	case "vmess":
		// For VMess: tag, address, port, uuid, encryption, network
		network := "tcp" // Default network for outbound
		if node.NetworkSettings.Transport != "" {
			network = node.NetworkSettings.Transport
		}
		return xc.MakeVmessOutbound(tag, host, port, password, node.Encryption, network), nil
	case "trojan":
		// For Trojan: tag, address, port, password, network
		network := "tcp" // Default network for outbound
		if node.NetworkSettings.Transport != "" {
			network = node.NetworkSettings.Transport
		}
		return xc.MakeTrojanOutbound(tag, host, port, password, network), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}
}

func (w *Writer) LocalConfig() (*xray.Config, error) {
	clients := w.clients()

	apiPort, err := utils.FreePort()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	xc := xray.NewConfig(w.c.Xray.LogLevel)
	xc.FindInbound("api").Port = apiPort

	// TODO: For now, maintain backward compatibility by creating Shadowsocks inbounds
	// In the future, this will be replaced with per-node protocol configuration
	var key string
	if len(clients) > 0 {
		// Create relay inbound (hardcoded port for now - will be per-node later)
		relayPort := 8443
		if key, err = utils.Key32(); err != nil {
			return nil, err
		}
		if utils.PortFree(relayPort) {
			xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
				"relay",
				key,
				config.ShadowsocksMethod,
				"tcp,udp",
				relayPort,
				clients,
			))
		}

		// Create reverse inbound (hardcoded port for now - will be per-node later)
		reversePort := 8444
		if key, err = utils.Key32(); err != nil {
			return nil, err
		}
		if utils.PortFree(reversePort) {
			xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
				"reverse",
				key,
				config.ShadowsocksMethod,
				"tcp,udp",
				reversePort,
				clients,
			))
		}

		// Create direct inbound (hardcoded port for now - will be per-node later)
		directPort := 8445
		if key, err = utils.Key32(); err != nil {
			return nil, err
		}
		if utils.PortFree(directPort) {
			xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
				"direct",
				key,
				config.ShadowsocksMethod,
				"tcp,udp",
				directPort,
				clients,
			))
		}
	}

	// Add routing rules
	if len(clients) > 0 {
		xc.Routing.Rules = append(xc.Routing.Rules, &xray.Rule{
			InboundTag:  []string{"direct"},
			OutboundTag: "out",
		})

		if len(w.database.Content.Nodes) > 0 {
			xc.Routing.Rules = append(xc.Routing.Rules, &xray.Rule{
				InboundTag:  []string{"relay"},
				BalancerTag: "relay",
			})
			xc.Routing.Rules = append(xc.Routing.Rules, &xray.Rule{
				InboundTag:  []string{"reverse"},
				BalancerTag: "portal",
			})
		}
	}

	// Add balancers
	if len(w.database.Content.Nodes) > 0 {
		xc.Routing.Balancers = append(xc.Routing.Balancers, &xray.Balancer{Tag: "relay", Selector: []string{}})
		xc.Routing.Balancers = append(xc.Routing.Balancers, &xray.Balancer{Tag: "portal", Selector: []string{}})
	}

	// Configure nodes
	for _, s := range w.database.Content.Nodes {
		inboundPort, err := utils.FreePort()
		if err != nil {
			return nil, errors.WithStack(err)
		}

		// Create reverse connection setup - ALWAYS use Shadowsocks for internal manager-to-node communication
		// The selected protocol (VMess, VLESS, etc.) is only used for client-facing inbounds
		if key, err = utils.Key32(); err != nil {
			return nil, err
		}
		xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
			fmt.Sprintf("internal-%d", s.Id),
			key,
			config.Shadowsocks2022Method,
			"tcp",
			inboundPort,
			nil,
		))

		// Create client-facing inbound using the node's configured protocol
		// Use the node's configured listening port
		clientPort := s.ListeningPort
		if clientPort == 0 {
			// Fallback to random port if not configured
			clientPort, err = utils.FreePort()
			if err != nil {
				return nil, errors.WithStack(err)
			}
		}
		
		clientInbound, err := w.makeProtocolInbound(
			s,
			fmt.Sprintf("client-%d", s.Id),
			"", // password will be generated inside makeProtocolInbound
			"tcp",
			clientPort,
			nil,
		)
		if err != nil {
			// Fallback to Shadowsocks if protocol inbound creation fails
			if key, err = utils.Key32(); err != nil {
				return nil, err
			}
			clientInbound = xc.MakeShadowsocksInbound(
				fmt.Sprintf("client-%d", s.Id),
				key,
				config.Shadowsocks2022Method, // Use 2022 method for consistency
				"tcp",
				clientPort,
				nil,
			)
		}
		xc.Inbounds = append(xc.Inbounds, clientInbound)

		xc.Reverse.Portals = append(xc.Reverse.Portals, &xray.ReverseItem{
			Tag:    fmt.Sprintf("portal-%d", s.Id),
			Domain: fmt.Sprintf("s%d.reverse.proxy", s.Id),
		})

		xc.Routing.Rules = append(xc.Routing.Rules, &xray.Rule{
			InboundTag:  []string{fmt.Sprintf("internal-%d", s.Id)},
			OutboundTag: fmt.Sprintf("portal-%d", s.Id),
		})

		xc.FindBalancer("portal").Selector = append(
			xc.FindBalancer("portal").Selector,
			fmt.Sprintf("portal-%d", s.Id),
		)

		// Create relay outbound
		outboundRelayPort, err := utils.FreePort()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if key, err = utils.Key32(); err != nil {
			return nil, err
		}
		xc.Outbounds = append(xc.Outbounds, xc.MakeShadowsocksOutbound(
			fmt.Sprintf("relay-%d", s.Id),
			s.Host,
			key,
			config.Shadowsocks2022Method,
			outboundRelayPort,
		))

		xc.FindBalancer("relay").Selector = append(
			xc.FindBalancer("relay").Selector,
			fmt.Sprintf("relay-%d", s.Id),
		)
	}

	return xc, nil
}

func (w *Writer) RemoteConfig(node *database.Node, lastUpdate time.Time, password string) *xray.Config {
	xc := xray.NewConfig(w.c.Xray.LogLevel)

	xc.Metadata = &xray.Metadata{
		UpdatedAt: lastUpdate.Format(time.RFC3339),
		UpdatedBy: w.database.Content.Settings.Host,
	}

	// Create relay inbound on node
	relayOutbound := w.xray.Config().FindOutbound(fmt.Sprintf("relay-%d", node.Id))
	if relayOutbound != nil {
		xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
			"direct",
			relayOutbound.Settings.Servers[0].Password,
			relayOutbound.Settings.Servers[0].Method,
			"tcp",
			relayOutbound.Settings.Servers[0].Port,
			nil,
		))
		xc.Routing.Rules = append(
			xc.Routing.Rules,
			&xray.Rule{
				InboundTag:  []string{"direct"},
				OutboundTag: "out",
			},
		)
	}

	// Create reverse outbound connection - Always use Shadowsocks for internal communication
	internalOutbound := w.xray.Config().FindInbound(fmt.Sprintf("internal-%d", node.Id))
	if internalOutbound != nil {
		xc.Outbounds = append(xc.Outbounds, xc.MakeShadowsocksOutbound(
			"internal",
			w.database.Content.Settings.Host,
			internalOutbound.Settings.Password,
			internalOutbound.Settings.Method,
			internalOutbound.Port,
		))
		xc.Reverse.Bridges = append(xc.Reverse.Bridges, &xray.ReverseItem{
			Tag:    "bridge",
			Domain: fmt.Sprintf("s%d.reverse.proxy", node.Id),
		})
		xc.Routing.Rules = append(
			xc.Routing.Rules,
			&xray.Rule{
				InboundTag:  []string{"bridge"},
				Domain:      []string{fmt.Sprintf("full:s%d.reverse.proxy", node.Id)},
				OutboundTag: "internal",
			},
			&xray.Rule{
				InboundTag:  []string{"bridge"},
				OutboundTag: "out",
			},
		)
	}

	// Create client-facing inbound using node's configured protocol
	clientInbound, err := w.makeProtocolInbound(node, "remote", password, "tcp", node.ListeningPort, w.clients())
	if err == nil && clientInbound != nil {
		xc.Inbounds = append(xc.Inbounds, clientInbound)
		xc.Routing.Rules = append(
			xc.Routing.Rules,
			&xray.Rule{
				InboundTag:  []string{"remote"},
				OutboundTag: "out",
			},
		)
	}

	return xc
}

func New(config *config.Config, database *database.Database, xray *xray.Xray) *Writer {
	return &Writer{c: config, database: database, xray: xray}
}
