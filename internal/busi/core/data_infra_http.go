package core

import (
	"context"
	"data-extraction-notify/pkg/models/busi"
	"data-extraction-notify/pkg/utils"
	"net/http"

	log "github.com/sirupsen/logrus"
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

	// find tipset of this topic
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
