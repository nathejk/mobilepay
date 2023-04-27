package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nathejk/shared-go/messages"
	"github.com/nathejk/shared-go/types"

	"nathejk.dk/cmd/api/dims"
	"nathejk.dk/cmd/api/mobilepay"
	"nathejk.dk/pkg/nats"
	"nathejk.dk/pkg/streaminterface"
	"nathejk.dk/table"
)

type config struct {
	port int
	env  string

	db struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	stream struct {
		dsn string
	}
	mobilepay struct {
		reportApiKey string
	}
}

func main() {
	fmt.Println("Starting API service")

	var cfg config
	flag.IntVar(&cfg.port, "port", 80, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.mobilepay.reportApiKey, "report-api-key", os.Getenv("REPORT_API_KEY"), "Report API key created in Mobilepay Portal")
	flag.StringVar(&cfg.stream.dsn, "stan-dsn", os.Getenv("STAN_DSN"), "DSN for stream server")
	flag.Parse()

	natsstream := nats.NewNATSStreamUnique(cfg.stream.dsn, "mobilepay-cron")
	defer natsstream.Close()

	db, err := sql.Open("mysql", os.Getenv("DB_DSN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	d := dims.New(natsstream, db)
	d.RunWaitLive(ctx)

	query := table.NewQuery(db)
	api := mobilepay.NewApiClient("https://api.mobilepay.dk", cfg.mobilepay.reportApiKey)

	go func() {
		ProcessTransactions(api, query, natsstream)
		time.Sleep(10 * time.Minute)
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("REQUEST:\n%s", string(reqDump))
		w.Write([]byte(""))
	})
	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		transactions, _ := api.Transactions(1000, 1)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	})

	http.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		transfers, _ := api.Transfers(1000, 1)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transfers)
	})
	http.HandleFunc("/paymentpoints", func(w http.ResponseWriter, r *http.Request) {
		paymentpoints, _ := api.PaymentPoints(1000, 1)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(paymentpoints)
	})

	fmt.Println("Running webserver")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.port), nil))
}

func ProcessTransactions(api mobilepay.ApiClient, q table.Query, natsstream streaminterface.Stream) {
	response, _ := api.Transactions(1000, 1)
	for _, transaction := range response.Transactions {
		if transaction.Message == "" {
			continue
		}
		body := messages.NathejkPaymentReceived{
			Amount:    transaction.Amount,
			Currency:  transaction.Currency,
			Timestamp: transaction.Timestamp,
		}
		teamID := types.TeamID(transaction.Message)
		memberID := types.MemberID(transaction.Message)
		if q.GetTeamStatus(teamID) != types.SignupStatusNone {
			if q.TeamPaymentAmount(transaction.Timestamp, teamID) != nil {
				continue
			}
			body.TeamID = teamID
		} else if q.GetMemberStatus(memberID) != types.MemberStatusNone {
			if q.MemberPaymentAmount(transaction.Timestamp, memberID) != nil {
				continue
			}
			body.MemberID = memberID
		} else {
			continue
		}
		msg := natsstream.MessageFunc()(streaminterface.SubjectFromStr("nathejk:payment.received"))
		msg.SetBody(&body)
		msg.SetMeta(&messages.Metadata{Producer: "mobilepay-cron"})
		if err := natsstream.Publish(msg); err != nil {
			log.Print(err)
		}
	}
}
