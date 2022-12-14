package watch

import (
	"context"
	"data-extraction-notify/internal/busi/core/cache"
	"data-extraction-notify/pkg/models/busi"
	"data-extraction-notify/pkg/utils"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/go-redis/redis"

	lotusapi "github.com/filecoin-project/lotus/api"
	log "github.com/sirupsen/logrus"
)

var tipsetsCache *cache.TipSetCache

type NotifyEvent struct {
	ctx  context.Context
	done func()

	cache        *cache.TipSetCache // caches tipsets for possible reversion
	tick         *time.Ticker
	jobIsRunning bool // if a job is running
}

func NewNotifyEvent(ctx context.Context, done func(), c *cache.TipSetCache) *NotifyEvent {

	w := &NotifyEvent{
		ctx:  ctx,
		done: done,

		cache:        cache.NewTipSetCache(),
		tick:         time.NewTicker(35 * time.Second),
		jobIsRunning: false,
	}

	return w
}

var notifyEvent *NotifyEvent

func PushTipsets(ctx context.Context, done func(), cli *redis.Client, changeSlice []*lotusapi.HeadChange) {
	if notifyEvent == nil {
		notifyEvent = NewNotifyEvent(ctx, done, tipsetsCache)
	}

	// Eliminate unnecessary "revert" data and
	// merge the same height tipset: apply0, apply1, applyN...
	for _, event := range changeSlice {
		if event.Type == "revert" {
			continue
		}

		log.Infof("recieve: %v", event.Val.Height())
		notifyEvent.cache.Add(event.Val)
	}

	if !notifyEvent.jobIsRunning {
		go notifyEvent.cacheConsumer(cli)
	}
}

func (n *NotifyEvent) cacheConsumer(cli *redis.Client) {
	f := func() error {
		buffer, err := n.cache.PopAll()
		if err != nil {
			return nil
		}

		// find relevant topics
		topics, err := topicsFind(context.Background())
		if err != nil || topics == nil {
			return nil
		}

		// Push tipset to multiple topics of mq.
		for _, tipset := range buffer {
			for _, topic := range topics {
				version, err := recordTipset(context.Background(), topic.Id, topic.TopicName, tipset)
				if err != nil {
					continue
				}

				b, err := json.Marshal(&busi.Message{Version: int(version), Tipset: *tipset})
				if err != nil {
					log.Errorf("T: %v marshal json error: %v", tipset.Height(), err)
					return err
				}

				log.Infof("push tipset: %v/version: %v to topic: %v", tipset.Height(), version, topic.TopicName)
				err = cli.LPush(topic.TopicName, b).Err()
				if err != nil {
					log.Errorf("push tipset: err: %v", tipset.Height(), err)
					return err
				}
			}
		}
		return nil
	}

	defer n.done()
	for {
		select {
		case <-n.ctx.Done():
			log.Errorf("ticktack, ctx done, receive signal: %s", n.ctx.Err().Error())
			return
		case <-n.tick.C:
			log.Infof("ticktack, cacheConsumer func running, push %v tipsets", n.cache.Len())
			f()
		}
	}
}

func topicsFind(ctx context.Context) ([]*busi.Topics, error) {
	t := make([]*busi.Topics, 0)
	if err := utils.EngineGroup[utils.DBExtract].Where("state = 0").Find(&t); err != nil {
		log.Errorf("topics find execute sql error: %v", err)
		return nil, err
	}

	if len(t) == 0 {
		log.Warning("No relevant data of topics.")
		return nil, nil
	}

	return t, nil
}

func recordTipset(ctx context.Context, topicId uint64, topicName string, tipset *types.TipSet) (uint32, error) {
	sql := fmt.Sprintf("select * from tipsets_state where topic_id = %v and tipset = %v and version = (select max(version) from tipsets_state where topic_id = %v and tipset = %v) limit 1",
		topicId, tipset.Height(), topicId, tipset.Height())

	result, err := utils.EngineGroup[utils.DBExtract].QueryString(sql)
	if err != nil {
		log.Errorf("record tipset execute sql error: %v", err)
		return 0, err
	}

	if len(result) > 0 {
		idTmp, _ := strconv.ParseUint(result[0]["id"], 10, 64)
		version, _ := strconv.ParseUint(result[0]["version"], 10, 64)

		t := busi.TipsetsState{
			Id:            idTmp,
			TopicId:       topicId,
			TopicName:     topicName,
			Tipset:        uint64(tipset.Height()),
			Version:       uint32(version) + 1,
			State:         0,
			NotFoundState: 0,
			RetryTimes:    0,
			Description:   "",
			LastUpdate:    time.Now(),
		}

		if _, err := utils.EngineGroup[utils.DBExtract].Where("id = ?", t.Id).Cols("version", "state", "not_found_state", "retry_times", "description").Update(&t); err != nil {
			log.Errorf("record tipset execute sql error: %v", err)
			return 0, err
		}

		return t.Version, nil
	} else {
		t := busi.TipsetsState{
			TopicId:   topicId,
			TopicName: topicName,
			Tipset:    uint64(tipset.Height()),

			Version:       0,
			State:         0,
			NotFoundState: 0,
			RetryTimes:    0,
			Description:   "",
			LastUpdate:    time.Now(),
		}

		if _, err := utils.EngineGroup[utils.DBExtract].Insert(&t); err != nil {
			log.Errorf("record tipset execute sql error: %v", err)
			return 0, err
		}

		return t.Version, nil
	}
}
