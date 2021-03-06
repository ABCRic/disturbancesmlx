package dataobjects

import (
	"errors"
	"fmt"

	sq "github.com/gbl08ma/squirrel"
	"github.com/heetch/sqalx"
)

// Lobby is a station lobby
type Lobby struct {
	ID      string
	Name    string
	Station *Station
}

// GetLobbies returns a slice with all registered lobbies
func GetLobbies(node sqalx.Node) ([]*Lobby, error) {
	return getLobbiesWithSelect(node, sdb.Select())
}

// GetLobbiesForStation returns a slice with the lobbies for the specified station
func GetLobbiesForStation(node sqalx.Node, id string) ([]*Lobby, error) {
	return getLobbiesWithSelect(node, sdb.Select().Where(sq.Eq{"station_id": id}))
}

func getLobbiesWithSelect(node sqalx.Node, sbuilder sq.SelectBuilder) ([]*Lobby, error) {
	lobbies := []*Lobby{}

	tx, err := node.Beginx()
	if err != nil {
		return lobbies, err
	}
	defer tx.Commit() // read-only tx

	rows, err := sbuilder.Columns("station_lobby.id", "station_lobby.name", "station_lobby.station_id").
		From("station_lobby").
		RunWith(tx).Query()
	if err != nil {
		return lobbies, fmt.Errorf("getLobbiesWithSelect: %s", err)
	}
	defer rows.Close()

	var stationIDs []string
	for rows.Next() {
		var lobby Lobby
		var stationID string
		err := rows.Scan(
			&lobby.ID,
			&lobby.Name,
			&stationID)
		if err != nil {
			return lobbies, fmt.Errorf("getLobbiesWithSelect: %s", err)
		}
		if err != nil {
			return lobbies, fmt.Errorf("getLobbiesWithSelect: %s", err)
		}
		lobbies = append(lobbies, &lobby)
		stationIDs = append(stationIDs, stationID)
	}
	if err := rows.Err(); err != nil {
		return lobbies, fmt.Errorf("getLobbiesWithSelect: %s", err)
	}
	for i := range stationIDs {
		lobbies[i].Station, err = GetStation(tx, stationIDs[i])
		if err != nil {
			return lobbies, fmt.Errorf("getLobbiesWithSelect: %s", err)
		}
	}
	return lobbies, nil
}

// GetLobby returns the lobby with the given ID
func GetLobby(node sqalx.Node, id string) (*Lobby, error) {
	s := sdb.Select().
		Where(sq.Eq{"id": id})
	lobbies, err := getLobbiesWithSelect(node, s)
	if err != nil {
		return nil, err
	}
	if len(lobbies) == 0 {
		return nil, errors.New("Lobby not found")
	}
	return lobbies[0], nil
}

// Exits returns the exits of this lobby
func (lobby *Lobby) Exits(node sqalx.Node) ([]*Exit, error) {
	s := sdb.Select().
		Where(sq.Eq{"lobby_id": lobby.ID})
	return getExitsWithSelect(node, s)
}

// Schedules returns the schedules of this lobby
func (lobby *Lobby) Schedules(node sqalx.Node) ([]*Schedule, error) {
	s := sdb.Select().
		Where(sq.Eq{"lobby_id": lobby.ID})
	return getSchedulesWithSelect(node, s)
}

// Update adds or updates the Lobby
func (lobby *Lobby) Update(node sqalx.Node) error {
	tx, err := node.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = lobby.Station.Update(tx)
	if err != nil {
		return errors.New("AddLobby: " + err.Error())
	}

	_, err = sdb.Insert("lobby").
		Columns("id", "name", "station_id").
		Values(lobby.ID, lobby.Name, lobby.Station.ID).
		Suffix("ON CONFLICT (id) DO UPDATE SET name = ?, station_id = ?",
			lobby.Name, lobby.Station.ID).
		RunWith(tx).Exec()

	if err != nil {
		return errors.New("AddLobby: " + err.Error())
	}
	return tx.Commit()
}

// Delete deletes the Lobby
func (lobby *Lobby) Delete(node sqalx.Node) error {
	tx, err := node.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = sdb.Delete("lobby").
		Where(sq.Eq{"id": lobby.ID}).RunWith(tx).Exec()
	if err != nil {
		return fmt.Errorf("RemoveLobby: %s", err)
	}
	return tx.Commit()
}
