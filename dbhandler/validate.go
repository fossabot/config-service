package dbhandler

import (
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
)

type Validator[T types.DocContent] func(c *gin.Context, docs []T) (verifiedDocs []T, valid bool)

func ValidateGUIDExistence[T types.DocContent](c *gin.Context, docs []T) ([]T, bool) {
	guid := c.Param(consts.GUIDField)
	if guid != "" && len(docs) > 1 {
		ResponseBadRequest(c, "GUID in path is not allowed in bulk request")
	}
	for i := range docs {
		if ; guid != "" {
			docs[i].SetGUID(guid)
		}
		if docs[i].GetGUID() == "" {
			ResponseMissingGUID(c)
			return nil, false
		}
	}
	return docs, true
}

type UniqueKeyValueInfo[T types.DocContent] func() (key string, mandatory bool, valueGetter func(T) string)

func ValidateUniqueValues[T types.DocContent](uniqueKeyValues ...UniqueKeyValueInfo[T]) func(c *gin.Context, docs []T) ([]T, bool) {
	return func(c *gin.Context, docs []T) ([]T, bool) {
		filter := NewFilterBuilder()
		projection := NewProjectionBuilder()
		keys2Values := map[string][]string{}
		for _, uniqueKeyValue := range uniqueKeyValues {
			key, mandatory, valueGetter := uniqueKeyValue()
			values := []string{}
			for _, doc := range docs {
				value := valueGetter(doc)
				if mandatory && value == "" {
					ResponseMissingKey(c, key) //TODO: change to more generic error
					return nil, false
				}
				if slices.Contains(values, value) {
					ResponseDuplicateKey(c, key, value) //TODO: change to more generic error
					return nil, false
				}
				values = append(values, value)
			}
			if len(filter.Get()) > 0 {
				filter.WarpOr()
			}
			if len(docs) > 1 {
				filter.WithIn(key, values)
			} else {
				filter.WithValue(key, values[0])
			}
			projection.Include(key)
			keys2Values[key] = values
		}

		if existingDocs, err := FindForCustomer[T](c, filter, projection.Get()); err != nil {
			ResponseInternalServerError(c, "failed to read documents", err)
			return nil, false
		} else if len(existingDocs) > 0 {
			key2ExistingValues := map[string][]string{}
			for _, uniqueKeyValue := range uniqueKeyValues {
				key, _, valueGetter := uniqueKeyValue()
				values := keys2Values[key]
				for _, doc := range existingDocs {
					value := valueGetter(doc)
					if slices.Contains(values, value) && !slices.Contains(key2ExistingValues[key], value) {
						key2ExistingValues[key] = append(key2ExistingValues[key], value)
					}
				}
			}
			ResponseDuplicateKeysNValues(c, key2ExistingValues)
			return nil, false
		}
		return docs, true
	}
}

func NameKeyGetter[T types.DocContent]() (key string, mandatory bool, valueGetter func(T) string) {
	return "name", true, func(doc T) string { return doc.GetName() }
}
