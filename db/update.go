package db

import (
	"config-service/types"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"k8s.io/utils/strings/slices"
)

// helpers to build db update commands

// GetUpdateFieldValuesCommand creates update command for multiple values for a field
// note - if the field is an array, the values will be added to the array
func GetUpdateFieldValuesCommand(m map[string]interface{}, fieldName string) bson.D {
	return bson.D{bson.E{Key: "$set", Value: map2BsonD(m, fieldName)}}
}

func GetUpdateDocValuesCommand(m map[string]interface{}) bson.D {
	return bson.D{bson.E{Key: "$set", Value: map2BsonD(m, "")}}
}

// GetUpdateFieldValueCommand creates update command for a single value for a field
func GetUpdateFieldValueCommand(i interface{}, fieldName string) bson.D {
	return bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: fieldName, Value: i}}}}
}

// GetUpdateFieldValueCommand creates update command for a DocContent removing excluded fields
// if includeFields is not empty, only the fields in the list will be included
func GetUpdateDocCommand[T types.DocContent](i T, includeFields []string, excludeFields ...string) (bson.D, error) {
	var m map[string]interface{}
	if data, err := json.Marshal(i); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	for _, f := range excludeFields {
		delete(m, f)
	}
	if len(includeFields) > 0 {
		for k := range m {
			if !slices.Contains(includeFields, k) {
				delete(m, k)
			}
		}
	}
	if len(m) == 0 {
		return nil, NoFieldsToUpdateError{}
	}
	return GetUpdateDocValuesCommand(m), nil
}

// helper to build bson.D from map
func map2BsonD(m map[string]interface{}, fieldName string) bson.D {
	var result bson.D
	if fieldName != "" {
		fieldName += "."
	}
	for k, v := range m {
		result = append(result, bson.E{Key: fmt.Sprintf("%s%s", fieldName, k), Value: v})
	}
	return result
}
