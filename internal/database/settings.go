package database

type Settings struct {
	AdminPassword     string            `json:"admin_password" validate:"required,min=8,max=32"`
	Host              string            `json:"host" validate:"required,max=128"`
	TrafficRatio      float64           `json:"traffic_ratio" validate:"min=1,max=1024"`
	SingetServer      string            `json:"singet_server" validate:"omitempty,url"`
	ResetPolicy       string            `json:"reset_policy" validate:"omitempty,oneof=monthly"`
	EncryptionOptions EncryptionOptions `json:"encryption_options"`
}

type EncryptionOptions struct {
	VMess  []string `json:"vmess"`
	VLESS  []string `json:"vless"`
	Trojan []string `json:"trojan"`
	SS     []string `json:"ss"`
}
