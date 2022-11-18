package core

import (
	"context"
	"data-extraction-notify/pkg/models/busi"
	"data-extraction-notify/pkg/utils"
	"net/http"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"

	"data-extraction-notify/internal/busi/core/gap"
	"data-extraction-notify/internal/busi/core/replay"
	"data-extraction-notify/internal/busi/core/walk"
)

func TopicSignIn(ctx context.Context, r *Topic) *utils.BuErrorResponse {
	var (
		has bool
		err error
		t   busi.Topics
	)

	if has, err = utils.EngineGroup[utils.DBExtract].Where("topic_name = ?", r.Topic).Get(&t); err != nil {
		log.Errorf("topic sign in execute sql error: %v", err)
		return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
	}

	if !has {
		t.TopicName = r.Topic
		if _, err = utils.EngineGroup[utils.DBExtract].Insert(&t); err != nil {
			log.Errorf("topic sign in execute sql error: %v", err)
			return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
		}
	}

	return nil
}

func TopicSignOff(ctx context.Context, r *Topic) *utils.BuErrorResponse {
	var t busi.Topics

	if _, err := utils.EngineGroup[utils.DBExtract].Where("topic_name = ?", r.Topic).Delete(&t); err != nil {
		log.Errorf("topic sign off execute sql error: %v", err)
		return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
	}

	return nil
}

func ReportTipsetState(ctx context.Context, r *TipsetState) *utils.BuErrorResponse {
	var (
		hasTopic  bool
		hasTipset bool

		topic  busi.Topics
		tipset busi.TipsetsState

		err error
	)

	// find topic
	if hasTopic, err = utils.EngineGroup[utils.DBExtract].Where("topic_name = ?", r.Topic).Get(&topic); err != nil {
		log.Errorf("report tipset state execute sql error: %v", err)
		return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
	}
	if !hasTopic {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: utils.ErrDataExtractNotifyTopicNotFoundErr}
	}

	// find current version tipset of this topic
	if hasTipset, err = utils.EngineGroup[utils.DBExtract].Where("topic_id = ? and tipset = ? and version = ?", topic.Id, r.Tipset, r.Version).Get(&tipset); err != nil {
		log.Errorf("report tipset state execute sql error: %v", err)
		return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
	}
	if !hasTipset {
		log.Errorf("couldn't find version: %v of tipset: %v, maybe it was updated.", r.Version, r.Tipset)
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: utils.ErrDataExtractNotifyTipsetNotFoundErr}
	}

	if tipset.State == 1 {
		log.Warnf("The previous tipset's task has been successfully executed: %+v, request params are: %+v", tipset, r)
		return nil
	} else {
		tipset.State = r.State
		tipset.Description = r.Description
		if r.State == 2 && r.NotFoundState == 1 {
			tipset.NotFoundState = r.NotFoundState
		}
		if _, err = utils.EngineGroup[utils.DBExtract].Where("id = ?", tipset.Id).Update(&tipset); err != nil {
			log.Errorf("report tipset state execute sql error: %v", err)
			return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
		}
	}

	return nil
}

// https://github.com/filecoin-project/lily/blob/master/chain/walk/walker.go#L43
func WalkTipsetsRun(ctx context.Context, r *Walk) *utils.BuErrorResponse {
	lotusAPI, closer, err := utils.LotusHandshake(ctx, r.Lotus0)
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
	}
	defer closer()

	rdb := redis.NewClient(&redis.Options{
		Addr:     r.Mq,
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	log.Infof("connect to mq: %v", r.Mq)
	if err := rdb.Ping().Err(); err != nil {
		log.Error(err)
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	//----------

	head, err := lotusAPI.ChainHead(ctx)
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}

	}
	if head.Height() < abi.ChainEpoch(r.MinHeight) {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: utils.ErrDataExtractNotifyHeightErr}
	}

	start := head
	// Start at maxHeight+1 so that the tipset at maxHeight becomes the parent for any tasks that need to make a diff between two tipsets.
	// A walk where min==max must still process two tipsets to be sure of extracting data.
	if head.Height() > abi.ChainEpoch(r.MaxHeight+1) {
		start, err = lotusAPI.ChainGetTipSetByHeight(ctx, abi.ChainEpoch(r.MaxHeight), head.Key())
		if err != nil {
			return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
		}
	}

	if err := walk.NewWalker(lotusAPI, rdb, r.MinHeight, r.MaxHeight, r.Topic).WalkChain(ctx, start); err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	return nil
}

// There are two main points in time when gaps happened
// a. network fluctuations
// b. system(Lotus0, notify server, task model, mq) upgrade
// we should schedule this api per 1 minute, check the height from "current head-10-60height" to "current head-10height"(30mins enough to upgrade the system)
func GapFill(ctx context.Context, r *Gap) *utils.BuErrorResponse {
	lotusAPI, closer, err := utils.LotusHandshake(ctx, r.Lotus0)
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
	}
	defer closer()

	rdb := redis.NewClient(&redis.Options{
		Addr:     r.Mq,
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	log.Infof("connect to mq: %v", r.Mq)
	if err := rdb.Ping().Err(); err != nil {
		log.Error(err)
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	//----------
	head, err := lotusAPI.ChainHead(ctx)
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	maxHeight := head.Height() - 10
	minHeight := head.Height() - 70

	if maxHeight <= 0 || minHeight <= 0 {
		return nil
	}

	start, err := lotusAPI.ChainGetTipSetByHeight(ctx, maxHeight, head.Key())
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	if err := gap.NewGapper(lotusAPI, rdb, uint64(minHeight), uint64(start.Height())).GapChain(ctx, start); err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	return nil
}

func ReplayTipsets(ctx context.Context, r *Retry) *utils.BuErrorResponse {
	lotusAPI, closer, err := utils.LotusHandshake(ctx, r.Lotus0)
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusInternalServerError, Response: utils.ErrInternalServer}
	}
	defer closer()

	rdb := redis.NewClient(&redis.Options{
		Addr:     r.Mq,
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	log.Infof("connect to mq: %v", r.Mq)
	if err := rdb.Ping().Err(); err != nil {
		log.Error(err)
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	//----------
	head, err := lotusAPI.ChainHead(ctx)
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	maxHeight := head.Height() - 20

	if maxHeight <= 0 {
		return nil
	}

	start, err := lotusAPI.ChainGetTipSetByHeight(ctx, maxHeight, head.Key())
	if err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	if err := replay.NewReplayer(lotusAPI, rdb, uint64(start.Height())).ReplayChain(ctx); err != nil {
		return &utils.BuErrorResponse{HttpCode: http.StatusOK, Response: &utils.Response{Code: utils.CodeInternalServer, Message: err.Error()}}
	}

	return nil
}
