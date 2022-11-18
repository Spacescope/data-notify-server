package gap

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

type gapJob struct {
	jobIsRunning bool
	startTime    time.Time
	endTime      time.Time
}

type Gapper struct {
	node *api.FullNodeStruct
	rdb  *redis.Client

	minHeight      uint64
	maxHeight      uint64
	indicatesTopic string

	topics []*busi.Topics

	*gapJob
}

var insJob *gapJob
var once sync.Once

func NewGapper(node *api.FullNodeStruct, rdb *redis.Client, minHeight, maxHeight uint64) *Gapper {
	g := &Gapper{
		node: node,
		rdb:  rdb,

		minHeight: minHeight,
		maxHeight: maxHeight,

		indicatesTopic: "all",
	}

	once.Do(func() {
		insJob = &gapJob{}
	})

	g.gapJob = insJob

	return g
}

func (g *Gapper) GapChain(ctx context.Context, ts *types.TipSet) error {
	if g.jobIsRunning {
		str := fmt.Sprintf("The previous gap's job has begun at the time: %v, pls wait for it finishes or ctrl^c it.", g.startTime)
		log.Infof(str)
		return errors.New(str)
	} else {
		var err error

		{
			g.jobIsRunning = true
			g.startTime = time.Now()
			g.topics = nil

			log.Infof("Gap runs at time: %v, works for topic: %v, from: %v - to: %v", g.startTime, g.indicatesTopic, g.minHeight, g.maxHeight)
		}

		defer func() {
			g.jobIsRunning = false
			g.endTime = time.Now()

			log.Infof("Gap has finished the jobs: %v", g.endTime)
		}()

		for ts.Height() >= abi.ChainEpoch(g.minHeight) && ts.Height() != 0 {
			select {
			case <-ctx.Done():
				log.Errorf("Gap cancel by: %v", ctx.Err())
				return nil
			default:
			}

			log.Infof("Gapper tipset: %v", ts.Height())
			// busi
			if err := g.insertMQ(ts); err != nil {
				return err
			}

			ts, err = g.node.ChainGetTipSet(ctx, ts.Parents())
			if err != nil {
				//log.Errorf("ChainGetTipSet, tipset: %v, err: %v", ts.Height(), err) // ts will be modified by ChainGetTipSet, it will possibly assign nil
				log.Errorf("ChainGetTipSet err: %v", err)
				return err
			}
		}
		return nil
	}
}

func (g *Gapper) topicsFind() ([]*busi.Topics, error) {
	t := make([]*busi.Topics, 0)

	if g.indicatesTopic == "all" {
		if err := utils.EngineGroup[utils.DBExtract].Where("state = 0").Find(&t); err != nil {
			log.Errorf("Gap, topics find execute sql error: %v", err)
			return nil, err
		}
	} else {
		if err := utils.EngineGroup[utils.DBExtract].Where("topic_name = ? and state = 0", g.indicatesTopic).Find(&t); err != nil {
			log.Errorf("Gap, topics find execute sql error: %v", err)
			return nil, err
		}
	}

	if len(t) == 0 {
		log.Warning("Gap, No relevant data of topics.")
		return nil, nil
	}

	return t, nil
}

func (g *Gapper) insertMQ(ts *types.TipSet) error {
	if g.topics == nil { // lazy
		var err error
		g.topics, err = g.topicsFind()
		if err != nil {
			return err
		}

		if len(g.topics) == 0 {
			s := fmt.Sprintf("Gap, No topic(s):%v found", g.indicatesTopic)
			log.Error(s)
			return errors.New(s)
		}
	}

	for _, topic := range g.topics {
		version, successful, err := g.recordTipset(topic.Id, topic.TopicName, ts)
		if err != nil || !successful {
			continue
		}

		b, err := json.Marshal(&busi.Message{Version: int(version), Tipset: *ts})
		if err != nil {
			log.Errorf("Gap, T: %v marshal json error: %v", ts.Height(), err)
			return err
		}

		// log.Infof("Gap, push tipset: %v/version: %v to topic: %v", ts.Height(), version, topic.TopicName)
		err = g.rdb.LPush(topic.TopicName, b).Err()
		if err != nil {
			log.Errorf("Gap, push tipset: err: %v", ts.Height(), err)
			return err
		}
	}

	return nil
}

func (g *Gapper) recordTipset(topicId uint64, topicName string, ts *types.TipSet) (uint32, bool, error) {
	tssState := make([]*busi.TipsetsState, 0)
	if err := utils.EngineGroup[utils.DBExtract].Where("topic_id = ? and tipset = ?", topicId, ts.Height()).Find(&tssState); err != nil {
		log.Errorf("Gapper, record tipset execute sql error: %v", err)
		return 0, false, err
	}

	if len(tssState) != 0 {
		log.Infof("Gapper, tipset: %v/topic: %v has been processed before", ts.Height(), topicName)
		return 0, false, nil
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

	if _, err := utils.EngineGroup[utils.DBExtract].Insert(&t); err != nil {
		log.Errorf("Gapper, record tipset execute sql error: %v", err)
		return 0, false, err
	}

	return t.Version, true, nil
}
