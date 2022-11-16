package utils

import (
	"context"

	jsonrpc "github.com/filecoin-project/go-jsonrpc"
	lotusapi "github.com/filecoin-project/lotus/api"
	log "github.com/sirupsen/logrus"
)

// Exponential backoff
func LotusHandshake(ctx context.Context, lotus0 string) (*lotusapi.FullNodeStruct, jsonrpc.ClientCloser, error) {
	log.Infof("connect to lotus0: %v", lotus0)

	var (
		err    error
		closer jsonrpc.ClientCloser
	)

	// authToken := "<value found in ~/.lotus/token>"
	// headers := http.Header{"Authorization": []string{"Bearer " + authToken}}
	// addr := "127.0.0.1:1234"

	var api lotusapi.FullNodeStruct
	// closer, err := jsonrpc.NewMergeClient(context.Background(), "ws://"+addr+"/rpc/v0", "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, headers)
	closer, err = jsonrpc.NewMergeClient(context.Background(), lotus0, "Filecoin", []interface{}{&api.Internal, &api.CommonStruct.Internal}, nil)
	if err != nil {
		log.Errorf("connecting to lotus failed: %s", err)
		return nil, nil, err
	}
	return &api, closer, nil
}
