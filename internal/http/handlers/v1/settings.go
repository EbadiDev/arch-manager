package v1

import (
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/ebadidev/arch-manager/internal/coordinator"
	"github.com/ebadidev/arch-manager/internal/database"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

func SettingsShow(d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, d.Content.Settings)
	}
}

func SettingsUpdate(coordinator *coordinator.Coordinator, d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var r database.Settings
		if err := c.Bind(&r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Cannot parse the request body.",
			})
		}
		if err := validator.New().Struct(r); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": fmt.Sprintf("Validation error: %v", err.Error()),
			})
		}

		d.Locker.Lock()
		defer d.Locker.Unlock()

		d.Content.Settings = &r

		if err := d.Save(); err != nil {
			return errors.WithStack(err)
		}

		go coordinator.SyncConfigs()

		return c.JSON(http.StatusOK, r)
	}
}

func SettingsXrayRestart(coordinator *coordinator.Coordinator) echo.HandlerFunc {
	return func(c echo.Context) error {
		go coordinator.SyncConfigs()
		return c.NoContent(http.StatusNoContent)
	}
}
