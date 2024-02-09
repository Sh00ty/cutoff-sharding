package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Sh00ty/cutoff-sharding/internal/app"
	cutoffs "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs"
	in_mem_repo "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs/repo/in-mem"
	"github.com/Sh00ty/cutoff-sharding/pkg/id/snowflake"
	"github.com/go-chi/chi/v5"
)

type timer struct {
	createdAt time.Time
}

func (t timer) GetTime() uint64 {
	// держим 100k rps на одном поде
	return uint64(time.Since(t.createdAt).Milliseconds())
}

func main() {
	keyGen := snowflake.NewGenerator(0x3FF, timer{createdAt: time.Now()})

	manager := cutoffs.NewCutOffManager(
		in_mem_repo.NewInMemRepo[snowflake.ID, int64](),
		uint64(10),
		keyGen,
	)

	app := app.NewApp(manager, keyGen)
	mux := chi.NewMux()
	mux.Post("/describe/", app.HandleDescribeID)
	mux.Post("/set/", app.HandleSet)
	mux.Post("/get/", app.HandleGet)
	mux.Post("/add-node/", app.HandleAddNode)

	log.Println("start listening on 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Printf("\nerror in lister and serv: %v", err)
	}
	log.Println("closed")
}
