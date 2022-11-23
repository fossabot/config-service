package dbhandler

import (
	"encoding/json"
	"fmt"
	"kubescape-config-service/types"

	"go.mongodb.org/mongo-driver/bson"
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
func GetUpdateDocCommand[T types.DocContent](i T, excludeFields ...string) (bson.D, error) {
	var m map[string]interface{}
	if data, err := json.Marshal(i); err != nil {
		return nil, err
	} else if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	for _, f := range excludeFields {
		delete(m, f)
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
