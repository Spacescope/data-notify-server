package walk

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

type walkJob struct {
	jobIsRunning bool
	startTime    time.Time
	endTime      time.Time
}

type Walker struct {
	node *api.FullNodeStruct
	rdb  *redis.Client

	minHeight      uint64
	maxHeight      uint64
	indicatesTopic string

	topics []*busi.Topics

	*walkJob
}

var insJob *walkJob
var once sync.Once

func NewWalker(node *api.FullNodeStruct, rdb *redis.Client, minHeight, maxHeight uint64, indicatesTopic string) *Walker {

	w := &Walker{
		node: node,
		rdb:  rdb,

		minHeight: minHeight,
		maxHeight: maxHeight,

		indicatesTopic: indicatesTopic,
	}

	once.Do(func() {
		insJob = &walkJob{}
	})

	w.walkJob = insJob

	return w
}

func (w *Walker) WalkChain(ctx context.Context, ts *types.TipSet, force bool) error {
	if w.jobIsRunning {
		str := fmt.Sprintf("The previous walk's job has begun at the time: %v, pls wait for it finishes or ctrl^c it.", w.startTime)
		log.Infof(str)
		return errors.New(str)
	} else {
		var err error

		{
			w.jobIsRunning = true
			w.startTime = time.Now()
			w.topics = nil

			log.Infof("Walk runs at time: %v, works for topic: %v, from: %v - to: %v", w.startTime, w.indicatesTopic, w.minHeight, w.maxHeight)
		}

		defer func() {
			w.jobIsRunning = false
			w.endTime = time.Now()

			log.Infof("Walk has finished the jobs: %v", w.endTime)
		}()

		for ts.Height() >= abi.ChainEpoch(w.minHeight) && ts.Height() != 0 {
			select {
			case <-ctx.Done():
				log.Errorf("Walk cancel by: %v", ctx.Err())
				return nil
			default:
			}

			log.Infof("Walk tipset: %v", ts.Height())
			// busi
			if err := w.insertMQ(ts, force); err != nil {
				return err
			}

			ts, err = w.node.ChainGetTipSet(ctx, ts.Parents())
			if err != nil {
				//log.Errorf("ChainGetTipSet, tipset: %v, err: %v", ts.Height(), err) // ts will be modified by ChainGetTipSet, it will possibly assign nil
				log.Errorf("ChainGetTipSet err: %v", err)
				return err
			}
		}
		return nil
	}
}

func (w *Walker) topicsFind() ([]*busi.Topics, error) {
	t := make([]*busi.Topics, 0)

	if w.indicatesTopic == "all" {
		if err := utils.EngineGroup[utils.DBExtract].Where("state = 0").Find(&t); err != nil {
			log.Errorf("Walk, topics find execute sql error: %v", err)
			return nil, err
		}
	} else {
		if err := utils.EngineGroup[utils.DBExtract].Where("topic_name = ? and state = 0", w.indicatesTopic).Find(&t); err != nil {
			log.Errorf("Walk, topics find execute sql error: %v", err)
			return nil, err
		}
	}

	if len(t) == 0 {
		log.Warning("Walk, No relevant data of topics.")
		return nil, nil
	}

	return t, nil
}

func (w *Walker) insertMQ(ts *types.TipSet, force bool) error {
	if w.topics == nil { // lazy
		var err error
		w.topics, err = w.topicsFind()
		if err != nil {
			return err
		}

		if len(w.topics) == 0 {
			s := fmt.Sprintf("Walk, No topic(s):%v found", w.indicatesTopic)
			log.Error(s)
			return errors.New(s)
		}
	}

	for _, topic := range w.topics {
		version, successful, err := w.recordTipset(topic.Id, topic.TopicName, ts, force)
		if err != nil || !successful {
			continue
		}

		b, err := json.Marshal(&busi.Message{Version: int(version), Tipset: *ts, Force: force})
		if err != nil {
			log.Errorf("Walk, T: %v marshal json error: %v", ts.Height(), err)
			return err
		}

		// log.Infof("Walk, push tipset: %v/version: %v to topic: %v", ts.Height(), version, topic.TopicName)
		err = w.rdb.LPush(topic.TopicName, b).Err()
		if err != nil {
			log.Errorf("Walk, push tipset: err: %v", ts.Height(), err)
			return err
		}
	}

	return nil
}

func (w *Walker) recordTipset(topicId uint64, topicName string, ts *types.TipSet, force bool) (uint32, bool, error) {
	var (
		tsState busi.TipsetsState
	)
	b, err := utils.EngineGroup[utils.DBExtract].Where("topic_id = ? and tipset = ?", topicId, ts.Height()).Get(&tsState)
	if err != nil {
		log.Errorf("Walk, record tipset execute sql error: %v", err)
		return 0, false, err
	}

	t := busi.TipsetsState{
		TopicId:   topicId,
		TopicName: topicName,
		Tipset:    uint64(ts.Height()),

		Version:       0,
		State:         0,
		NotFoundState: 0,
		RetryTimes:    0,
		Description:   "",
		LastUpdate:    time.Now(),
	}

	if !force {
		if b {
			log.Infof("Walk, tipset: %v/topic: %v has been processed before", ts.Height(), topicName)
			return 0, false, nil
		}

		if _, err := utils.EngineGroup[utils.DBExtract].Insert(&t); err != nil {
			log.Errorf("Walk, record tipset execute sql error: %v", err)
			return 0, false, err
		}

		return t.Version, true, nil
	} else {
		log.Infof("Forced walk, tipset: %v/topic: %v", ts.Height(), topicName)

		if b {
			tsState.State = 0
			tsState.RetryTimes++

			if _, err := utils.EngineGroup[utils.DBExtract].Where("id = ?", tsState.Id).Cols("state").Update(&tsState); err != nil {
				log.Errorf("Forced Walk, record tipset execute sql error: %v", err)
				return 0, false, err
			}
		} else {
			if _, err := utils.EngineGroup[utils.DBExtract].Insert(&t); err != nil {
				log.Errorf("Forced Walk, record tipset execute sql error: %v", err)
				return 0, false, err
			}
		}

		return t.Version, true, nil
	}
}
