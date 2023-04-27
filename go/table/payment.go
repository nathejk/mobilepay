package table

import (
	"fmt"
	"log"
	"time"

	"github.com/nathejk/shared-go/messages"
	"github.com/nathejk/shared-go/types"
	"github.com/seaqrs/tablerow"
	"nathejk.dk/pkg/streaminterface"

	_ "embed"
)

type Payment struct {
	Timestamp time.Time `sql:"ts"`
	Amount    float32   `sql:"amount"`
	Currency  string
	TeamID    types.TeamID
	MemberID  types.MemberID
}

type payment struct {
	w tablerow.Consumer
}

func NewPayment(w tablerow.Consumer) *payment {
	table := &payment{w: w}
	if err := w.Consume(table.CreateTableSql()); err != nil {
		log.Fatalf("Error creating table %q", err)
	}
	return table
}

//go:embed payment.sql
var paymentSchema string

func (t *payment) CreateTableSql() string {
	return paymentSchema
}

func (t *payment) Consumes() (subjs []streaminterface.Subject) {
	return []streaminterface.Subject{
		streaminterface.SubjectFromStr("nathejk"),
	}
}

func (t *payment) HandleMessage(msg streaminterface.Message) {
	switch msg.Subject().Subject() {
	case "nathejk:payment.received":
		var body messages.NathejkPaymentReceived
		if err := msg.Body(&body); err != nil {
			return
		}
		args := []any{body.Timestamp, body.Amount, body.Currency, body.TeamID, body.MemberID}
		err := t.w.Consume(fmt.Sprintf("INSERT INTO payment SET ts=%q, amount=%d, currency=%q, teamId=%q, memberId=%q ON DUPLICATE KEY UPDATE ts=VALUES(ts)", args...))
		if err != nil {
			log.Fatalf("Error consuming sql %q", err)
		}
	}
}
