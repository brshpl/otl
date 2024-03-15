package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/brshpl/otl/internal/usecase"
	"github.com/brshpl/otl/pkg/logger"
)

type oneTimeLinkRoutes struct {
	t usecase.OneTimeLink
	l logger.Interface
}

func newOneTimeLinkRoutes(handler *gin.RouterGroup, t usecase.OneTimeLink, l logger.Interface) *oneTimeLinkRoutes {
	r := &oneTimeLinkRoutes{t, l}

	h := handler.Group("/oneTimeLink")
	{
		h.POST("/create", r.create)
		h.POST("/get", r.getWithJSON)
	}

	return r
}

type createRequest struct {
	Data string `json:"data" binding:"required"`
}

type createResponse struct {
	Link string `json:"link"`
}

func (r *oneTimeLinkRoutes) create(c *gin.Context) {
	var request createRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error("http - v1 - create: %s", err)
		errorResponse(c, http.StatusBadRequest, "invalid request body")

		return
	}

	link, err := r.t.Create(c.Request.Context(), request.Data)
	if err != nil {
		r.l.Error("http - v1 - create: %s", err)
		errorResponse(c, http.StatusInternalServerError, "database problems")

		return
	}

	c.JSON(http.StatusOK, createResponse{link})
}

type getRequest struct {
	Link string `json:"link" binding:"required"`
}

type getResponse struct {
	Data string `json:"data"`
}

func (r *oneTimeLinkRoutes) getWithJSON(c *gin.Context) {
	var request getRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		r.l.Error("http - v1 - get: %s", err)
		errorResponse(c, http.StatusBadRequest, "invalid request body")

		return
	}

	r.get(c, request)
}

func (r *oneTimeLinkRoutes) getWithParam(c *gin.Context) {
	link := c.Param("link")

	r.get(c, getRequest{link})
}

func (r *oneTimeLinkRoutes) get(c *gin.Context, request getRequest) {
	data, err := r.t.Get(c.Request.Context(), request.Link)
	if errors.Is(err, usecase.ErrLinkExpired) {
		errorResponse(c, http.StatusGone, "link expired")

		return
	} else if errors.Is(err, usecase.ErrInvalidLink) {
		errorResponse(c, http.StatusBadRequest, "invalid link")

		return
	} else if err != nil {
		r.l.Error("http - v1 - get: %s", err)
		errorResponse(c, http.StatusInternalServerError, "oneTimeLink service problems")

		return
	}

	c.JSON(http.StatusOK, getResponse{data})
}
