package dims

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"nathejk.dk/pkg/memorystream"
	"nathejk.dk/pkg/memstat"
	"nathejk.dk/pkg/nats"
	"nathejk.dk/pkg/sqlpersister"
	"nathejk.dk/pkg/stream"
	"nathejk.dk/pkg/streaminterface"
	"nathejk.dk/table"
)

type dims struct {
	//stream streaminterface.Stream
	//db     *sql.DB
	swtch *stream.Switch
}

func New(natsstream *nats.NATSStream, db *sql.DB) *dims {
	//natsstream := nats.NewNATSStreamUnique(os.Getenv("STAN_DSN"), "tilmelding-api")
	//defer natsstream.Close()
	sqlw := sqlpersister.New(db)

	consumers := []streaminterface.Consumer{
		table.NewTeamStatus(sqlw),
		table.NewTeamMemberStatus(sqlw),
		table.NewPayment(sqlw),
	}

	memstream := memorystream.New()
	streammux := stream.NewStreamMux(memstream)
	streammux.Handles(natsstream, natsstream.Channels()...)
	swtch, err := stream.NewSwitch(streammux, consumers)
	if err != nil {
		log.Fatal(err)
	}

	d := dims{
		swtch: swtch,
		//stream: stream,
		//db:     db,
	}
	return &d
}

func (d *dims) RunWaitLive(ctx context.Context) error {
	live := make(chan struct{})
	go func() {
		err := d.swtch.Run(ctx, func() {
			memstat.PrintMemoryStats()
			fmt.Println(d.swtch.Stats().Format())
			live <- struct{}{}
		})
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Waiting for live
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-live:
	}
	return nil
}
