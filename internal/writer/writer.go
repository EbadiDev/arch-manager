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

// Protocol factory method - currently only supports Shadowsocks, extensible for future protocols
func (w *Writer) makeProtocolInbound(node *database.Node, tag, password, network string, port int, clients []*xray.Client) (*xray.Inbound, error) {
	switch node.Protocol {
	case "shadowsocks":
		return w.makeShadowsocksInbound(tag, password, node.Encryption, network, port, clients), nil
	case "vless":
		// TODO: Implement when arch-node package supports MakeVlessInbound
		return nil, fmt.Errorf("VLESS protocol not yet supported - waiting for arch-node package update")
	case "vmess":
		// TODO: Implement when arch-node package supports MakeVmessInbound  
		return nil, fmt.Errorf("VMess protocol not yet supported - waiting for arch-node package update")
	case "trojan":
		// TODO: Implement when arch-node package supports MakeTrojanInbound
		return nil, fmt.Errorf("Trojan protocol not yet supported - waiting for arch-node package update")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}
}

// Shadowsocks inbound factory (uses existing arch-node method)
func (w *Writer) makeShadowsocksInbound(tag, password, method, network string, port int, clients []*xray.Client) *xray.Inbound {
	xc := xray.NewConfig(w.c.Xray.LogLevel)
	return xc.MakeShadowsocksInbound(tag, password, method, network, port, clients)
}

// Protocol outbound factory - currently only supports Shadowsocks
func (w *Writer) makeProtocolOutbound(node *database.Node, tag, host, password, method string, port int) (*xray.Outbound, error) {
	switch node.Protocol {
	case "shadowsocks":
		return w.makeShadowsocksOutbound(tag, host, password, method, port), nil
	case "vless":
		// TODO: Implement when arch-node package supports MakeVlessOutbound
		return nil, fmt.Errorf("VLESS outbound not yet supported - waiting for arch-node package update")
	case "vmess":
		// TODO: Implement when arch-node package supports MakeVmessOutbound
		return nil, fmt.Errorf("VMess outbound not yet supported - waiting for arch-node package update")
	case "trojan":
		// TODO: Implement when arch-node package supports MakeTrojanOutbound
		return nil, fmt.Errorf("Trojan outbound not yet supported - waiting for arch-node package update")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", node.Protocol)
	}
}

// Shadowsocks outbound factory (uses existing arch-node method)
func (w *Writer) makeShadowsocksOutbound(tag, host, password, method string, port int) *xray.Outbound {
	xc := xray.NewConfig(w.c.Xray.LogLevel)
	return xc.MakeShadowsocksOutbound(tag, host, password, method, port)
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

		// Create reverse connection setup
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

	// Create reverse outbound connection
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

	// Create remote inbound (hardcoded port for now - will be per-node later)
	remotePort := 8446
	if utils.PortFree(remotePort) {
		xc.Inbounds = append(xc.Inbounds, xc.MakeShadowsocksInbound(
			"remote",
			password,
			config.ShadowsocksMethod,
			"tcp",
			remotePort,
			w.clients(),
		))
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
