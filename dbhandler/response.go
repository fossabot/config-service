package dbhandler

import (
	"config-service/utils/log"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	//error messages
	MissingName      = "name is required"
	MissingGUID      = "document guid is required"
	DocumentNotFound = "document not found"
)

func ResponseInternalServerError(c *gin.Context, msg string, err error) {
	log.LogNTraceError(msg, err, c)
	errText := msg
	if err != nil {
		errText = fmt.Sprintf("%s error: %s", msg, err.Error())
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": errText})
}

func ResponseDocumentNotFound(c *gin.Context) {
	log.LogNTrace(DocumentNotFound, c)
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": DocumentNotFound})
}

func ResponseMissingGUID(c *gin.Context) {
	ResponseBadRequest(c, MissingGUID)
}

func ResponseDuplicateNames(c *gin.Context, name ...string) {
	var msg string
	if len(name) == 0 {
		msg = "name already exists"
	} else if len(name) == 1 {
		msg = fmt.Sprintf("name %s already exists", name[0])
	} else {
		msg = fmt.Sprintf("names %s already exist", strings.Join(name, ","))
	}
	ResponseBadRequest(c, msg)
}

func ResponseMissingName(c *gin.Context) {
	ResponseBadRequest(c, MissingName)
}

func ResponseBadRequest(c *gin.Context, msg string) {
	log.LogNTrace(msg, c)
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
}

func ResponseFailedToBindJson(c *gin.Context, err error) {
	log.LogNTraceError("failed to bind json", err, c)
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
