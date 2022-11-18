package replay

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

type Replayer struct {
	node *api.FullNodeStruct
	rdb  *redis.Client

	maxHeight uint64

	*replayJob
}

var insJob *replayJob
var once sync.Once

func NewReplayer(node *api.FullNodeStruct, rdb *redis.Client, maxHeight uint64) *Replayer {
	r := &Replayer{
		node: node,
		rdb:  rdb,

		maxHeight: maxHeight,
	}

	once.Do(func() {
		insJob = &replayJob{}
	})

	r.replayJob = insJob

	return r
}

func (r *Replayer) ReplayChain(ctx context.Context) error {
	if r.jobIsRunning {
		str := fmt.Sprintf("The previous replay's job has begun at the time: %v, pls wait for it finishes or ctrl^c it.", r.startTime)
		log.Infof(str)
		return errors.New(str)
	} else {
		var err error

		{
			r.jobIsRunning = true
			r.startTime = time.Now()

			log.Infof("Replay runs at time: %v", r.startTime)
		}

		defer func() {
			r.jobIsRunning = false
			r.endTime = time.Now()

			log.Infof("Replay has finished the jobs: %v", r.endTime)
		}()

		tss, err := r.getReplayTipsets()
		if err != nil {
			return err
		}

		for _, busiTs := range tss {
			select {
			case <-ctx.Done():
				log.Errorf("Gap cancel by: %v", ctx.Err())
				return nil
			default:
			}

			t := *busiTs

			log.Infof("Replay tipset: %v", busiTs.Tipset)

			if err := r.insertMQ(ctx, t); err != nil {
				return err
			}
		}

		return nil
	}
}

func (r *Replayer) getReplayTipsets() ([]*busi.TipsetsState, error) {
	t := make([]*busi.TipsetsState, 0)

	if err := utils.EngineGroup[utils.DBExtract].Where("tipset <= ? and state != 1 and retry_times < 3", r.maxHeight).Find(&t); err != nil {
		log.Errorf("Replay, getReplayTipsets execute sql error: %v", err)
		return nil, err
	}

	return t, nil
}

func (r *Replayer) insertMQ(ctx context.Context, busiTs busi.TipsetsState) error {
	ts, err := r.node.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(busiTs.Tipset), types.EmptyTSK)
	if err != nil {
		log.Errorf("Replay, insertMQ ChainGetTipSetByHeight  err: %v", err)
		return err
	}

	busiTs.RetryTimes++

	if _, err := utils.EngineGroup[utils.DBExtract].ID(busiTs.Id).Update(&busiTs); err != nil {
		log.Errorf("Replay, insertMQ execute sql error: %v", err)
		return err
	}

	b, err := json.Marshal(&busi.Message{Version: int(busiTs.Version), Tipset: *ts})
	if err != nil {
		log.Errorf("Gap, T: %v marshal json error: %v", ts.Height(), err)
		return err
	}

	// log.Infof("Gap, push tipset: %v/version: %v to topic: %v", ts.Height(), version, topic.TopicName)
	err = r.rdb.LPush(busiTs.TopicName, b).Err()
	if err != nil {
		log.Errorf("Gap, push tipset: err: %v", ts.Height(), err)
		return err
	}

	return nil
}
