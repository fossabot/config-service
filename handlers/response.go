package handlers

import (
	"config-service/utils/log"
	"fmt"
	"net/http"
	"sort"
	"strings"

	plural "github.com/gertd/go-pluralize"
	"github.com/gin-gonic/gin"
)

const (
	//error messages
	MissingKey       = "%s is required"
	DocumentNotFound = "document not found"
)

var pluralize = plural.NewClient()

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

func ResponseDuplicateNames(c *gin.Context, names ...string) {
	ResponseDuplicateKey(c, "name", names...)
}

func ResponseDuplicateKey(c *gin.Context, key string, values ...string) {
	dupNames := map[string][]string{key: values}
	ResponseDuplicateKeysNValues(c, dupNames)
}

func ResponseDuplicateKeysNValues(c *gin.Context, key2Values map[string][]string) {
	var msg string
	keys := make([]string, 0, len(key2Values))
	for k := range key2Values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for key, values := range key2Values {
		sort.Strings(values)
		if len(msg) > 0 {
			msg += ", "
		}
		if len(values) == 0 {
			msg = key + " already exists"
		} else if len(values) == 1 {
			msg = fmt.Sprintf("%s %s already exists", key, values[0])
		} else {
			msg = fmt.Sprintf("%s %s already exist", pluralize.Plural(key), strings.Join(values, ","))
		}
	}
	ResponseBadRequest(c, msg)
}

func ResponseMissingGUID(c *gin.Context) {
	ResponseMissingKey(c, "guid")
}

func ResponseMissingName(c *gin.Context) {
	ResponseMissingKey(c, "name")
}

func ResponseMissingKey(c *gin.Context, key string) {
	ResponseBadRequest(c, fmt.Sprintf(MissingKey, key))
}

func ResponseBulkNotSupported(c *gin.Context) {
	ResponseBadRequest(c, "bulk operations are not supported")
}

func ResponseBadRequest(c *gin.Context, msg string) {
	log.LogNTrace(msg, c)
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg})
}

func ResponseFailedToBindJson(c *gin.Context, err error) {
	log.LogNTraceError("failed to bind json", err, c)
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}
