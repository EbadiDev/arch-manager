package pages

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/ebadidev/arch-manager/internal/config"
	"github.com/ebadidev/arch-manager/internal/database"
	"github.com/labstack/echo/v4"
)

func Profile(config *config.Config, d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Response().Header().Set("Pragma", "no-cache")
		c.Response().Header().Set("Expires", "0")

		userId := c.QueryParams().Get("u")
		for _, u := range d.Content.Users {
			if u.Identity == userId {
				content, err := os.ReadFile(filepath.Join(config.Env.AppDirectory, "web/profile.html"))
				if err != nil {
					return err
				}
				return c.HTML(http.StatusOK, string(content))
			}
		}

		content, err := os.ReadFile(filepath.Join(config.Env.AppDirectory, "web/profile-404.html"))
		if err != nil {
			return err
		}
		return c.HTML(http.StatusOK, string(content))
	}
}
