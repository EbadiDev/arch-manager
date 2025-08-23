package database

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ebadidev/arch-manager/internal/config"
	"github.com/ebadidev/arch-manager/internal/utils"
	"github.com/ebadidev/arch-node/pkg/logger"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/random"
	"go.uber.org/zap"
)

type Content struct {
	Settings *Settings `json:"settings"`
	Stats    *Stats    `json:"stats"`
	Users    []*User   `json:"users"`
	Nodes    []*Node   `json:"nodes"`
}

type Database struct {
	Content *Content
	Locker  *sync.Mutex
	l       *logger.Logger
	c       *config.Config
}

func (d *Database) Init() error {
	d.Locker.Lock()
	defer d.Locker.Unlock()

	if utils.FileExist(d.c.Env.DatabasePath) {
		err := d.Load()
		return errors.WithStack(err)
	}

	err := d.Save()
	return errors.WithStack(err)
}

func (d *Database) Load() error {
	content, err := os.ReadFile(d.c.Env.DatabasePath)
	if err != nil {
		return errors.WithStack(err)
	}

	err = json.Unmarshal(content, d.Content)
	if err != nil {
		return errors.WithStack(err)
	}

	d.modify()

	err = validator.New().Struct(d)
	return errors.WithStack(err)
}

func (d *Database) modify() {
	for _, user := range d.Content.Users {
		if user.UsageResetAt == 0 {
			user.UsageResetAt = time.Now().UnixMilli()
		}
	}
}

func (d *Database) Save() error {
	content, err := json.Marshal(d.Content)
	if err != nil {
		return errors.WithStack(err)
	}

	err = os.WriteFile(d.c.Env.DatabasePath, content, 0755)
	return errors.WithStack(err)
}

func (d *Database) Close() {
	content, err := json.Marshal(d.Content)
	if err != nil {
		d.l.Error("database: close: cannot marshal data", zap.Error(errors.WithStack(err)))
	}

	if err = os.WriteFile(d.c.Env.DatabasePath, content, 0755); err != nil {
		d.l.Error("database: close: cannot save file", zap.Error(errors.WithStack(err)))
	}
}

func (d *Database) Backup() {
	d.Locker.Lock()
	defer d.Locker.Unlock()

	content, err := json.Marshal(d.Content)
	if err != nil {
		d.l.Error("database: cannot marshal data", zap.Error(errors.WithStack(err)))
	}

	path := strings.ToLower(fmt.Sprintf(d.c.Env.DatabaseBackupPath, time.Now().Format("Mon-15")))
	if err = os.WriteFile(path, content, 0755); err != nil {
		d.l.Fatal(
			"database: cannot save backup file", zap.String("file", path), zap.Error(errors.WithStack(err)),
		)
	}
}

func (d *Database) CountActiveUsers() int {
	activeUsersCount := len(d.Content.Users)
	for _, u := range d.Content.Users {
		if !u.Enabled {
			activeUsersCount--
		}
	}
	return activeUsersCount
}

func (d *Database) GenerateUserId() int {
	if len(d.Content.Users) > 0 {
		return d.Content.Users[len(d.Content.Users)-1].Id + 1
	} else {
		return 1
	}
}

func (d *Database) GenerateUserIdentity() string {
	return utils.UUID()
}

func (d *Database) GenerateUserPassword() string {
	for {
		// Generate a proper Shadowsocks 2022 compatible base64-encoded key
		key, err := utils.Key32()
		if err != nil {
			// Fallback to old method if key generation fails
			key = random.String(16)
		}
		
		isUnique := true
		for _, user := range d.Content.Users {
			if user.ShadowsocksPassword == key {
				isUnique = false
				break
			}
		}
		if isUnique {
			return key
		}
	}
}

func (d *Database) GenerateNodeId() int {
	if len(d.Content.Nodes) > 0 {
		return d.Content.Nodes[len(d.Content.Nodes)-1].Id + 1
	} else {
		return 1
	}
}

func New(l *logger.Logger, c *config.Config) *Database {
	return &Database{
		Locker: &sync.Mutex{},
		l:      l,
		c:      c,
		Content: &Content{
			Settings: &Settings{
				AdminPassword: "password",
				Host:          "127.0.0.1",
				TrafficRatio:  1,
				EncryptionOptions: EncryptionOptions{
					VMess:  []string{"auto", "none", "zero", "aes-128-gcm"},
					VLESS:  []string{"none"},
					Trojan: []string{"none"},
					SS: []string{
						"aes-128-gcm",
						"aes-256-gcm",
						"chacha20-poly1305",
						"xchacha20-poly1305",
						"chacha20-ietf-poly1305",
						"2022-blake3-aes-128-gcm",
						"2022-blake3-aes-256-gcm",
					},
				},
			},
			Stats: &Stats{
				TotalUsage:        0,
				TotalUsageBytes:   0,
				TotalUsageResetAt: time.Now().UnixMilli(),
			},
			Users: []*User{},
			Nodes: []*Node{},
		},
	}
}
