package database

// NodeStatus represents the status of a server (node).
type NodeStatus string

const (
	NodeStatusProcessing  NodeStatus = ""
	NodeStatusAvailable              = "available"
	NodeStatusDirty                  = "dirty"
	NodeStatusUnavailable            = "unavailable"
)

// Node represents a server (node) in the system.
type Node struct {
	// Basic Info
	Id         int        `json:"id"`
	Host       string     `json:"host" validate:"required,max=128"`
	HttpToken  string     `json:"http_token" validate:"required"`
	HttpPort   int        `json:"http_port" validate:"required,min=1,max=65536"`
	Usage      float64    `json:"usage" validate:"min=0"`
	UsageBytes int64      `json:"usage_bytes" validate:"min=0"`
	PushStatus NodeStatus `json:"push_status"`
	PullStatus NodeStatus `json:"pull_status"`
	PushedAt   int64      `json:"pushed_at"`
	PulledAt   int64      `json:"pulled_at"`

	// Core Configuration
	CoreType string `json:"core_type" validate:"required,oneof=xray"`

	// Protocol Configuration
	Protocol     string `json:"protocol" validate:"required,oneof=shadowsocks vmess vless trojan"`
	ServerName   string `json:"server_name" validate:"required,max=128"`
	ServerAddr   string `json:"server_address" validate:"required,max=128"`
	ServerIP     string `json:"server_ip" validate:"required,ip"`
	ServerPort   string `json:"server_port" validate:"required"`
	Encryption   string `json:"encryption" validate:"required"`

	// Network Configuration
	ListeningIP   string `json:"listening_ip" validate:"required,ip"`
	ListeningPort int    `json:"listening_port" validate:"required,min=1,max=65536"`
	SendThrough   string `json:"send_through" validate:"omitempty,ip"`

	// Advanced Settings
	DNSSettings     DNSConfig     `json:"dns_settings"`
	RoutingSettings RoutingConfig `json:"routing_settings"`
	NetworkSettings NetworkConfig `json:"network_settings"`

	// Security Configuration
	Security         string         `json:"security" validate:"required,oneof=tls reality none"`
	SecuritySettings SecurityConfig `json:"security_settings"`

	// Certificate & Fragment
	CertMode      string `json:"cert_mode" validate:"required,oneof=http file dns none"`
	Fragment      bool   `json:"fragment"`
	FragmentValue string `json:"fragment_value" validate:"omitempty"`
}

type DNSConfig struct {
	Servers []string          `json:"servers"`
	Hosts   map[string]string `json:"hosts"`
	Tag     string            `json:"tag"`
}

type RoutingConfig struct {
	Rules []RoutingRule `json:"rules"`
	Tag   string        `json:"tag"`
}

type RoutingRule struct {
	Type        string   `json:"type"`
	Domain      []string `json:"domain,omitempty"`
	IP          []string `json:"ip,omitempty"`
	Port        string   `json:"port,omitempty"`
	OutboundTag string   `json:"outbound_tag"`
}

type NetworkConfig struct {
	Transport           string                 `json:"transport" validate:"required,oneof=tcp http ws grpc kcp httpupgrade xhttp"`
	AcceptProxyProtocol bool                   `json:"accept_proxy_protocol"`
	Settings            map[string]interface{} `json:"settings"`
}

type SecurityConfig struct {
	TLS     *TLSConfig     `json:"tls,omitempty"`
	Reality *RealityConfig `json:"reality,omitempty"`
}

type TLSConfig struct {
	ServerName         string   `json:"server_name"`
	RejectUnknownSni   bool     `json:"reject_unknown_sni"`
	AllowInsecure      bool     `json:"allow_insecure"`
	Fingerprint        string   `json:"fingerprint"`
	SNI                string   `json:"sni"`
	CurvePreferences   string   `json:"curve_preferences"`
	ALPN               []string `json:"alpn"`
	ServerNameToVerify string   `json:"server_name_to_verify"`
}

type RealityConfig struct {
	Show         bool     `json:"show"`
	Dest         string   `json:"dest"`
	PrivateKey   string   `json:"private_key"`
	MinClientVer string   `json:"min_client_ver"`
	MaxClientVer string   `json:"max_client_ver"`
	MaxTimeDiff  int      `json:"max_time_diff"`
	ProxyProtocol int     `json:"proxy_protocol"`
	ShortIDs     []string `json:"short_ids"`
	ServerNames  []string `json:"server_names"`
	Fingerprint  string   `json:"fingerprint"`
	SpiderX      string   `json:"spider_x"`
	PublicKey    string   `json:"public_key"`
}
