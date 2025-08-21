package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/ebadidev/arch-manager/internal/coordinator"
	"github.com/ebadidev/arch-manager/internal/database"
	"github.com/ebadidev/arch-manager/internal/writer"
	"github.com/labstack/echo/v4"
)

func NodesConfigsShow(cdr *coordinator.Coordinator, writer *writer.Writer, d *database.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		d.Locker.Lock()
		defer d.Locker.Unlock()

		nodeId := c.Param("id")
		var node *database.Node
		for _, n := range d.Content.Nodes {
			if strconv.Itoa(n.Id) == nodeId {
				node = n
				node.PulledAt = time.Now().UnixMilli()
				node.PullStatus = database.NodeStatusAvailable

				if err := d.Save(); err != nil {
					return errors.WithStack(err)
				}
			}
		}
		if node == nil {
			return c.NoContent(http.StatusNotFound)
		}

		configs := writer.RemoteConfig(node, cdr.State().XrayUpdatedAt(), cdr.State().XraySharedPassword())

		return c.JSON(http.StatusOK, configs)
	}
}
