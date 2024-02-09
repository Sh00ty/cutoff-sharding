package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Sh00ty/cutoff-sharding/internal/repos/redis"
	"github.com/Sh00ty/cutoff-sharding/pkg/configuration/vnodes"
	cutoffs "github.com/Sh00ty/cutoff-sharding/pkg/cut-offs"
	"github.com/Sh00ty/cutoff-sharding/pkg/id/snowflake"
)

type App struct {
	manager *cutoffs.CutOffManager[snowflake.ID, int64]
	keyGen  *snowflake.Generator
}

func NewApp(manager *cutoffs.CutOffManager[snowflake.ID, int64], keyGen *snowflake.Generator) *App {
	return &App{manager: manager, keyGen: keyGen}
}

func writeErr(w http.ResponseWriter, format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Println(str)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("{\"error\": \"%s\"}", str)))
}

type SetIn struct {
	Val int64 `json:"val"`
}

type SetOut struct {
	ID uint64 `json:"id"`
}

func (a *App) HandleSet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}
	in := SetIn{}
	err = json.Unmarshal(body, &in)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}

	id := a.keyGen.GenerateID()
	err = a.manager.Set(ctx, id.(snowflake.ID), in.Val)
	if err != nil {
		writeErr(w, "failed to set key: %v", err)
		return
	}
	res, err := json.Marshal(SetOut{ID: uint64(id.(snowflake.ID))})
	if err != nil {
		writeErr(w, "failed to marshal set out %v", err)
		return
	}
	w.Write(res)
}

type GetIn struct {
	ID uint64 `json:"id"`
}

type GetOut struct {
	ID  uint64 `json:"id"`
	Val int64  `json:"val"`
}

func (a *App) HandleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}
	in := GetIn{}
	err = json.Unmarshal(body, &in)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}

	val, err := a.manager.Get(ctx, snowflake.ID(in.ID))
	if err != nil {
		writeErr(w, "failed to get key: %v", err)
		return
	}
	res, err := json.Marshal(GetOut{ID: uint64(in.ID), Val: *val})
	if err != nil {
		writeErr(w, "failed to marshal get out %v", err)
		return
	}
	w.Write(res)
}

type AddNodeIn struct {
	Places []uint64 `json:"places"`
	Addr   string   `json:"addr"`
}

type AddNodeOut struct {
	Ok bool `json:"ok"`
}

func (a *App) HandleAddNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}
	in := AddNodeIn{}
	err = json.Unmarshal(body, &in)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}
	var conf *vnodes.Configuration[snowflake.ID, int64]
	cutOff, err := a.manager.GetLast(ctx)
	if err != nil {
		conf = &vnodes.Configuration[snowflake.ID, int64]{}
		conf.Nodes = make([]cutoffs.Node[snowflake.ID, int64], len(in.Places))
	} else {
		if cutOff.Conf.Type() != cutoffs.Vnodes {
			writeErr(w, "unexpected configuration %v", cutOff.Conf.Type())
			return
		}
		c, ok := cutOff.Conf.Clone().(*vnodes.Configuration[snowflake.ID, int64])
		if !ok {
			panic("cant cant to vnodes conf")
		}
		conf = c
	}

	clnt, err := redis.New(ctx, in.Addr)
	if err != nil {
		writeErr(w, "failed to connect to redis: %v", err)
		return
	}

	for _, place := range in.Places {
		if place >= uint64(len(conf.Nodes)) {
			writeErr(w, "can't set new node in vnodes conf of len %d on %d", len(conf.Nodes), place)
			return
		}
		conf.Nodes[place] = cutoffs.Node[snowflake.ID, int64]{
			ID:   cutoffs.NodeID(in.Addr),
			Repo: clnt,
		}
	}

	err = a.manager.CreateCutOff(ctx, conf)
	if err != nil {
		writeErr(w, "failed to create cutoff: %v", err)
		return
	}

	res, err := json.Marshal(AddNodeOut{Ok: true})
	if err != nil {
		writeErr(w, "failed to marshal get out %v", err)
		return
	}
	w.Write(res)
}

type DescribeIDIn struct {
	ID uint64 `json:"id"`
}

type DescribeIDOut struct {
	ID      uint64 `json:"id"`
	Time    uint64 `json:"time"`
	Pod     uint16 `json:"pod"`
	Counter uint32 `json:"counter"`
	Hash    uint64 `json:"hash"`
}

func (a *App) HandleDescribeID(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}
	in := DescribeIDIn{}
	err = json.Unmarshal(body, &in)
	if err != nil {
		writeErr(w, "can't read body: %v", err)
		return
	}
	id := snowflake.ID(in.ID)
	res, err := json.Marshal(DescribeIDOut{
		ID:      uint64(in.ID),
		Time:    id.GetTime(),
		Pod:     id.GetPod(),
		Counter: uint32(id.GetCount()),
		Hash:    id.Hash(),
	})
	if err != nil {
		writeErr(w, "failed to marshal describe out %v", err)
		return
	}
	w.Write(res)
}
