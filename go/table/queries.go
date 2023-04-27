package table

import (
	"database/sql"
	"log"
	"time"

	"github.com/nathejk/shared-go/types"
)

type Query interface {
	GetTeamStatus(teamID types.TeamID) (status types.SignupStatus)
	GetMemberStatus(memberID types.MemberID) (status types.MemberStatus)
	TeamPaymentAmount(timestamp time.Time, teamID types.TeamID) (amount *float32)
	MemberPaymentAmount(timestamp time.Time, memberID types.MemberID) (amount *float32)
}

type query struct {
	stmts map[string]*sql.Stmt
}

func NewQuery(db *sql.DB) *query {
	prepare := func(query string) *sql.Stmt {
		stmt, err := db.Prepare(query)
		if err != nil {
			log.Fatal(err)
		}
		return stmt
	}
	q := &query{stmts: map[string]*sql.Stmt{}}
	q.stmts["teamstatus"] = prepare("SELECT status FROM teamstatus WHERE teamId = ?")
	q.stmts["memberstatus"] = prepare("SELECT status FROM teammemberstatus WHERE memberId = ?")
	q.stmts["teampaymentexists"] = prepare("SELECT amount FROM payment WHERE ts = ? AND teamId = ?")
	q.stmts["memberpaymentexists"] = prepare("SELECT amount FROM payment WHERE ts = ? AND memberId = ?")

	return q
}

func (q *query) GetTeamStatus(teamID types.TeamID) (status types.SignupStatus) {
	if nil != q.stmts["teamstatus"].QueryRow(teamID).Scan(&status) {
		return types.SignupStatus("")
	}
	return
}

func (q *query) GetMemberStatus(memberID types.MemberID) (status types.MemberStatus) {
	if nil != q.stmts["memberstatus"].QueryRow(memberID).Scan(&status) {
		return types.MemberStatus("")
	}
	return
}

func (q *query) TeamPaymentAmount(timestamp time.Time, teamID types.TeamID) (amount *float32) {
	if nil != q.stmts["teampaymentexists"].QueryRow(timestamp, teamID).Scan(&amount) {
		return nil
	}
	return
}

func (q *query) MemberPaymentAmount(timestamp time.Time, memberID types.MemberID) (amount *float32) {
	if nil != q.stmts["memberpaymentexists"].QueryRow(timestamp, memberID).Scan(&amount) {
		return nil
	}
	return
}
