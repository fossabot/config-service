package dbhandler

import (
	"encoding/json"
	"fmt"
	"kubescape-config-service/types"

	"go.mongodb.org/mongo-driver/bson"
)

func GetUpdateFieldValuesCommand(m map[string]interface{}, fieldName string) bson.D {
	return bson.D{bson.E{Key: "$set", Value: Map2BsonD(m, fieldName)}}
}

func GetUpdateFieldValueCommand(i interface{}, fieldName string) bson.D {
	return bson.D{bson.E{Key: "$set", Value: bson.D{bson.E{Key: fieldName, Value: i}}}}
}

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
	return GetUpdateFieldValuesCommand(m, ""), nil
}

func Map2BsonD(m map[string]interface{}, fieldName string) bson.D {
	var result bson.D
	if fieldName != "" {
		fieldName += "."
	}
	for k, v := range m {
		result = append(result, bson.E{Key: fmt.Sprintf("%s%s", fieldName, k), Value: v})
	}
	return result
}
