package v1

import (
	"data-extraction-notify/internal/busi/core"
	"data-extraction-notify/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Message queue topic sign in godoc
// @Description task group will sign in a mq topic use this API.
// @Tags DATA-EXTRACTION-API-Internal-V1
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
// @Tags DATA-EXTRACTION-API-Internal-V1
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

// task state godoc
// @Description task will report tipset state with this API.
// @Tags DATA-EXTRACTION-API-Internal-V1
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

	resp := core.ReportTipsetState(c.Request.Context(), &r)
	if resp != nil {
		app.HTTPResponse(resp.HttpCode, resp.Response)
		return
	}

	app.HTTPResponseOK(nil)
}
