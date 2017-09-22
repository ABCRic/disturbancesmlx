package resource

import (
	"github.com/gbl08ma/disturbancesmlx/dataobjects"
	"github.com/heetch/sqalx"
	"github.com/yarf-framework/yarf"
)

// Connection composites resource
type Connection struct {
	resource
}

type apiConnection struct {
	From                  *dataobjects.Station `msgpack:"-" json:"-"`
	To                    *dataobjects.Station `msgpack:"-" json:"-"`
	TypicalWaitingSeconds int                  `msgpack:"typWaitS" json:"typWaitS"`
	TypicalStopSeconds    int                  `msgpack:"typStopS" json:"typStopS"`
	TypicalSeconds        int                  `msgpack:"typS" json:"typS"`
}

type apiConnectionWrapper struct {
	apiConnection `msgpack:",inline"`
	FromID        string `msgpack:"from" json:"from"`
	ToID          string `msgpack:"to" json:"to"`
}

func (r *Connection) WithNode(node sqalx.Node) *Connection {
	r.node = node
	return r
}

func (n *Connection) Get(c *yarf.Context) error {
	tx, err := n.Beginx()
	if err != nil {
		return err
	}
	defer tx.Commit() // read-only tx

	if c.Param("from") != "" && c.Param("to") != "" {
		connection, err := dataobjects.GetConnection(tx, c.Param("from"), c.Param("to"))
		if err != nil {
			return err
		}
		data := apiConnectionWrapper{
			apiConnection: apiConnection(*connection),
			FromID:        connection.From.ID,
			ToID:          connection.To.ID,
		}

		RenderData(c, data)
	} else {
		connections, err := dataobjects.GetConnections(tx)
		if err != nil {
			return err
		}
		apiconnections := make([]apiConnectionWrapper, len(connections))
		for i := range connections {
			apiconnections[i] = apiConnectionWrapper{
				apiConnection: apiConnection(*connections[i]),
				FromID:        connections[i].From.ID,
				ToID:          connections[i].To.ID,
			}
		}
		RenderData(c, apiconnections)
	}
	return nil
}
