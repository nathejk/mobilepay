package table

import (
	"fmt"
	"log"

	"github.com/nathejk/shared-go/messages"
	"github.com/nathejk/shared-go/types"
	"github.com/seaqrs/tablerow"
	"nathejk.dk/pkg/streaminterface"

	_ "embed"
)

type Status struct {
	TeamID types.TeamID       `sql:"teamId"`
	Status types.SignupStatus `sql:"status"`
}

type status struct {
	w tablerow.Consumer
}

func NewTeamStatus(w tablerow.Consumer) *status {
	table := &status{w: w}
	if err := w.Consume(table.CreateTableSql()); err != nil {
		log.Fatalf("Error creating table %q", err)
	}
	return table
}

//go:embed teamstatus.sql
var teamStatusSchema string

func (t *status) CreateTableSql() string {
	return teamStatusSchema
}

func (t *status) Consumes() (subjs []streaminterface.Subject) {
	return []streaminterface.Subject{
		streaminterface.SubjectFromStr("nathejk"),
	}
}

func (t *status) HandleMessage(msg streaminterface.Message) {
	switch msg.Subject().Subject() {
	case "nathejk:patrulje.status.changed", "nathejk:klan.status.changed":
		var body messages.NathejkTeamStatusChanged
		if err := msg.Body(&body); err != nil {
			return
		}
		err := t.w.Consume(fmt.Sprintf("INSERT INTO teamstatus SET teamId=%q, status=%q ON DUPLICATE KEY UPDATE status=VALUES(status)", body.TeamID, body.Status))
		if err != nil {
			log.Fatalf("Error consuming sql %q", err)
		}
	}
}
