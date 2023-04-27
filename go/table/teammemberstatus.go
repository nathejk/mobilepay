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

type TeamMemberStatus struct {
	MemberID types.MemberID
	TeamID   types.TeamID
	TeamType types.TeamType
	Status   string
}

type teammemberstatus struct {
	w tablerow.Consumer
}

func NewTeamMemberStatus(w tablerow.Consumer) *teammemberstatus {
	table := &teammemberstatus{w: w}
	if err := w.Consume(table.CreateTableSql()); err != nil {
		log.Fatalf("Error creating table %q", err)
	}
	return table
}

//go:embed teammemberstatus.sql
var teamMemberStatusSchema string

func (t *teammemberstatus) CreateTableSql() string {
	return teamMemberStatusSchema
}

func (t *teammemberstatus) Consumes() (subjs []streaminterface.Subject) {
	return []streaminterface.Subject{
		streaminterface.SubjectFromStr("nathejk"),
	}
}

func (t *teammemberstatus) HandleMessage(msg streaminterface.Message) {
	sql := "INSERT INTO teammemberstatus (memberId, teamId, teamType, status) VALUES (%q,%q,%q,%q) ON DUPLICATE KEY UPDATE teamId=VALUES(teamId), teamType=VALUES(teamType)"

	switch msg.Subject().Subject() {
	case "nathejk:senior.updated":
		var body messages.NathejkMemberUpdated
		if err := msg.Body(&body); err != nil {
			return
		}
		err := t.w.Consume(fmt.Sprintf(sql, body.MemberID, body.TeamID, types.TeamTypeKlan, types.MemberStatusRegistered))
		if err != nil {
			log.Fatalf("Error consuming sql %q", err)
		}

	case "nathejk:spejder.updated":
		var body messages.NathejkMemberUpdated
		if err := msg.Body(&body); err != nil {
			return
		}
		err := t.w.Consume(fmt.Sprintf(sql, body.MemberID, body.TeamID, types.TeamTypePatrulje, types.MemberStatusRegistered))
		if err != nil {
			log.Fatalf("Error consuming sql %q", err)
		}
		/*
			case "nathejk:senior.deleted":
				var body messages.NathejkMemberDeleted
				if err := msg.Body(&body); err != nil {
					return
				}
				err := c.w.Consume(fmt.Sprintf("DELETE FROM senior WHERE memberId=%q", body.MemberID))
				if err != nil {
					log.Fatalf("Error consuming sql %q", err)
				}*/
	}
}
