package busi

import (
	"context"
	"data-extraction-notify/internal/busi/core/watch"
	"time"

	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type NotifyServer struct {
	Lotus0 string
	Mq     string
	rdb    *redis.Client
}

func NewNotifyServer(lotus0, mq string) *NotifyServer {
	rdb := redis.NewClient(&redis.Options{
		Addr:     mq,
		Password: "",
		DB:       0,
	})

	log.Infof("connect to mq: %v\n", mq)
	if err := rdb.Ping().Err(); err != nil {
		panic(err)
	}

	return &NotifyServer{lotus0, mq, rdb}
}

func NotifyServerStart(ctx context.Context, done func(), lotus0, mq string) {
	defer done()
	s := NewNotifyServer(lotus0, mq)

	for {
		cancelSignal, _ := s.Watcher(ctx, done)
		if cancelSignal { // cancel due to signal
			return
		}
	}
}

func (s *NotifyServer) Watcher(ctx context.Context, done func()) (bool, error) { //(exit due to signal, return due to network problem/lotus problem)
	api, closer, err := s.lotusHandshake(ctx)
	if err != nil {
		log.Fatalf("calling chain head: %s", err)
	}
	defer closer()

	notifyChannel, err := api.ChainNotify(ctx)
	if err != nil {
		log.Fatalf("calling ChainNotify error: %s", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Errorf("watcher, ctx done, receive signal: %s", ctx.Err().Error())
			return true, nil
		// case <-time.After(time.Second * 30): // heartbeat
		// 	if _, err := api.ID(ctx); err != nil {
		// 		log.Errorf("keepalive failed, err: %s\n", err)
		// 		return false, err
		// 	}
		case headerSlice, ok := <-notifyChannel:
			if !ok {
				log.Errorf("calling ChainNotify channel err: %s", err)
				return false, err
			}
			log.Info("Get the notify channel event.")
			watch.PushTipsets(ctx, done, s.rdb, headerSlice)
		}
	}
}

// Exponential backoff
func (s *NotifyServer) lotusHandshake(ctx context.Context) (*lotusapi.FullNodeStruct, jsonrpc.ClientCloser, error) {
	log.Infof("connect to lotus0: %v", s.Lotus0)

	const MAXSLEEP int = 512
	var (
		err    error
		closer jsonrpc.ClientCloser
	)

	// authToken := "<value found in ~/.lotus/token>"
	// headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	// addr := "127.0.0.1:1234"

	var api lotusapi.FullNodeStruct
	for numsec := 1; numsec < MAXSLEEP; numsec <<= 1 {
		// closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
		closer, err = jsonrpc.NewMergeClient(context.Background(), s.Lotus0, "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, nil)
		if err == nil {
			return &api, closer, nil
		}
		log.Errorf("connecting to lotus failed: %s", err)
		if numsec <= MAXSLEEP/2 {
			time.Sleep(time.Duration(numsec) * time.Second)
		}
	}
	return nil, nil, err
}
