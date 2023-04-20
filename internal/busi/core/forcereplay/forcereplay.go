package forcereplay

import (
	"context"
	"data-extraction-notify/pkg/models/busi"
	"data-extraction-notify/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

type replayJob struct {
	jobIsRunning bool
	startTime    time.Time
	endTime      time.Time
}

type ForceReplayer struct {
	node *api.FullNodeStruct
	rdb  *redis.Client

	minHeight uint64
	maxHeight uint64

	*replayJob
}

var insJob *replayJob
var once sync.Once

func NewForceReplayer(node *api.FullNodeStruct, rdb *redis.Client, minHeight, maxHeight uint64) *ForceReplayer {
	r := &ForceReplayer{
		node: node,
		rdb:  rdb,

		minHeight: minHeight,
		maxHeight: maxHeight,
	}

	once.Do(func() {
		insJob = &replayJob{}
	})

	r.replayJob = insJob

	return r
}

func (r *ForceReplayer) ReplayChain(ctx context.Context) error {
	if r.jobIsRunning {
		str := fmt.Sprintf("The previous force replay's job has begun at the time: %v, pls wait for it finishes or ctrl^c it.", r.startTime)
		log.Infof(str)
		return errors.New(str)
	} else {
		var err error

		{
			r.jobIsRunning = true
			r.startTime = time.Now()

			log.Infof("Force replay runs at time: %v", r.startTime)
		}

		defer func() {
			r.jobIsRunning = false
			r.endTime = time.Now()

			log.Infof("Force replay has finished the jobs: %v", r.endTime)
		}()

		tss, err := r.getReplayTipsets()
		if err != nil {
			return err
		}

		for _, busiTs := range tss {
			select {
			case <-ctx.Done():
				log.Errorf("Force replay cancel by: %v", ctx.Err())
				return nil
			default:
			}

			t := *busiTs

			log.Infof("Force replay tipset: %v", busiTs.Tipset)

			if err := r.insertMQ(ctx, t); err != nil {
				return err
			}
		}

		return nil
	}
}

func (r *ForceReplayer) getReplayTipsets() ([]*busi.TipsetsState, error) {
	t := make([]*busi.TipsetsState, 0)

	if err := utils.EngineGroup[utils.DBExtract].Where("tipset between ? and ? and state != 1", r.minHeight, r.maxHeight).Find(&t); err != nil {
		log.Errorf("Force replay, getReplayTipsets execute sql error: %v", err)
		return nil, err
	}

	return t, nil
}

func (r *ForceReplayer) insertMQ(ctx context.Context, busiTs busi.TipsetsState) error {
	ts, err := r.node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(busiTs.Tipset), types.EmptyTSK)
	if err != nil {
		log.Errorf("Force replay, insertMQ ChainGetTipSetByHeight  err: %v", err)
		return err
	}

	busiTs.RetryTimes++

	if _, err := utils.EngineGroup[utils.DBExtract].ID(busiTs.Id).Update(&busiTs); err != nil {
		log.Errorf("Force replay, insertMQ execute sql error: %v", err)
		return err
	}

	b, err := json.Marshal(&busi.Message{Version: int(busiTs.Version), Tipset: *ts})
	if err != nil {
		log.Errorf("Force replay, T: %v marshal json error: %v", ts.Height(), err)
		return err
	}

	// log.Infof("Gap, push tipset: %v/version: %v to topic: %v", ts.Height(), version, topic.TopicName)
	err = r.rdb.LPush(busiTs.TopicName, b).Err()
	if err != nil {
		log.Errorf("Force replay, push tipset: err: %v", ts.Height(), err)
		return err
	}

	return nil
}
