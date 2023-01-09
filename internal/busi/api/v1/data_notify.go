package v1

import (
	"data-extraction-notify/internal/busi/core"
	"data-extraction-notify/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Message queue topic sign in godoc
// @Description task group will sign in a mq topic use this API.
// @Tags DATA-EXTRACTION-API-Internal-V1-CallByTaskModel
// @Accept application/json,json
// @Produce application/json,json
// @Param Topic body core.Topic false "Topic"
// @Success 200 {object} nil
// @Failure 400 {object} utils.ResponseWithRequestId
// @Failure 500 {object} utils.ResponseWithRequestId
// @Router /api/v1/topic [post]
func TopicSignIn(c *gin.Context) {
	app := utils.Gin{C: c}

	var r core.Topic
	if err := c.ShouldBindJSON(&r); err != nil {
		app.HTTPResponse(http.StatusBadRequest, utils.NewResponse(utils.CodeBadRequest, err.Error(), nil))
		return
	}

	resp := core.TopicSignIn(c.Request.Context(), &r)
	if resp != nil {
		app.HTTPResponse(resp.HttpCode, resp.Response)
		return
	}

	app.HTTPResponseOK(nil)
}

// Message queue topic delete godoc
// @Description delete a topic.
// @Tags DATA-EXTRACTION-API-Internal-V1-CallByTaskModel
// @Accept application/json,json
// @Produce application/json,json
// @Param Topic body core.Topic false "Topic"
// @Success 200 {object} nil
// @Failure 400 {object} utils.ResponseWithRequestId
// @Failure 500 {object} utils.ResponseWithRequestId
// @Router /api/v1/topic [delete]
func TopicDelete(c *gin.Context) {
	app := utils.Gin{C: c}

	var r core.Topic
	if err := c.ShouldBindJSON(&r); err != nil {
		app.HTTPResponse(http.StatusBadRequest, utils.NewResponse(utils.CodeBadRequest, err.Error(), nil))
		return
	}

	resp := core.TopicSignOff(c.Request.Context(), &r)
	if resp != nil {
		app.HTTPResponse(resp.HttpCode, resp.Response)
		return
	}

	app.HTTPResponseOK(nil)
}

// Task state godoc
// @Description task will report tipset state with this API.
// @Tags DATA-EXTRACTION-API-Internal-V1-CallByTaskModel
// @Accept application/json,json
// @Produce application/json,json
// @Param TipsetState body core.TipsetState false "TipsetState"
// @Success 200 {object} nil
// @Failure 400 {object} utils.ResponseWithRequestId
// @Failure 500 {object} utils.ResponseWithRequestId
// @Router /api/v1/task_state [post]
func ReportTipsetState(c *gin.Context) {
	app := utils.Gin{C: c}

	var r core.TipsetState
	if err := c.ShouldBindJSON(&r); err != nil {
		app.HTTPResponse(http.StatusBadRequest, utils.NewResponse(utils.CodeBadRequest, err.Error(), nil))
		return
	}

	if err := r.Validate(); err != nil {
		app.HTTPResponse(http.StatusBadRequest, utils.NewResponse(utils.CodeBadRequest, err.Error(), nil))
		return
	}

	var f core.Force
	if err := c.ShouldBindQuery(&f); err != nil {
		app.HTTPResponse(http.StatusBadRequest, utils.NewResponse(utils.CodeBadRequest, err.Error(), nil))
		return
	}

	resp := core.ReportTipsetState(c.Request.Context(), &r, f.Force)
	if resp != nil {
		app.HTTPResponse(resp.HttpCode, resp.Response)
		return
	}

	app.HTTPResponseOK(nil)
}

// Walk tipsets godoc
// @Description walk the historical DAG's tipsets.
// @Tags DATA-EXTRACTION-API-Internal-V1-CallByManual
// @Accept application/json,json
// @Produce application/json,json
// @Param Walk query core.Walk false "Walk"
// @Success 200 {object} nil
// @Failure 400 {object} utils.ResponseWithRequestId
// @Failure 500 {object} utils.ResponseWithRequestId
// @Router /api/v1/walk [post]
func WalkTipsets(c *gin.Context) {
	app := utils.Gin{C: c}

	var r core.Walk
	if err := c.ShouldBindQuery(&r); err != nil {
		app.HTTPResponse(http.StatusBadRequest, utils.NewResponse(utils.CodeBadRequest, err.Error(), nil))
		return
	}

	if err := r.Validate(); err != nil {
		app.HTTPResponse(http.StatusBadRequest, utils.NewResponse(utils.CodeBadRequest, err.Error(), nil))
		return
	}

	lotus0, _ := c.Get(LOTUS0)
	r.Lotus0, _ = lotus0.(string)

	mq, _ := c.Get(MQ)
	r.Mq, _ = mq.(string)

	resp := core.WalkTipsetsRun(c.Request.Context(), &r)
	if resp != nil {
		app.HTTPResponse(resp.HttpCode, resp.Response)
		return
	}

	app.HTTPResponseOK(nil)
}

// AutoGapFill tipsets godoc
// @Description automatic fill the gap's tipsets.
// @Tags DATA-EXTRACTION-API-Internal-V1-CallByScheduler
// @Accept application/json,json
// @Produce application/json,json
// @Param Gap query core.Gap false "Gap"
// @Success 200 {object} nil
// @Failure 400 {object} utils.ResponseWithRequestId
// @Failure 500 {object} utils.ResponseWithRequestId
// @Router /api/v1/gapfill [post]
func GapFill(c *gin.Context) {
	app := utils.Gin{C: c}

	var r core.Gap

	lotus0, _ := c.Get(LOTUS0)
	r.Lotus0, _ = lotus0.(string)

	mq, _ := c.Get(MQ)
	r.Mq, _ = mq.(string)

	resp := core.GapFill(c.Request.Context(), &r)
	if resp != nil {
		app.HTTPResponse(resp.HttpCode, resp.Response)
		return
	}

	app.HTTPResponseOK(nil)
}

// TipsetRetry godoc
// @Description replay the failed tipsets.
// @Tags DATA-EXTRACTION-API-Internal-V1-CallByScheduler
// @Accept application/json,json
// @Produce application/json,json
// @Success 200 {object} nil
// @Failure 400 {object} utils.ResponseWithRequestId
// @Failure 500 {object} utils.ResponseWithRequestId
// @Router /api/v1/retry [post]
func ReplayTipsets(c *gin.Context) {
	app := utils.Gin{C: c}

	var r core.Retry

	lotus0, _ := c.Get(LOTUS0)
	r.Lotus0, _ = lotus0.(string)

	mq, _ := c.Get(MQ)
	r.Mq, _ = mq.(string)

	resp := core.ReplayTipsets(c.Request.Context(), &r)
	if resp != nil {
		app.HTTPResponse(resp.HttpCode, resp.Response)
		return
	}

	app.HTTPResponseOK(nil)
}
