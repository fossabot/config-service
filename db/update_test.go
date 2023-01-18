package db

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMap2BsonD(t *testing.T) {
	testMap := map[string]interface{}{
		"state": map[string]interface{}{
			"onboarding": map[string]interface{}{
				"completed":   true,
				"companySize": "1-10",
				"interests":   []string{"a", "b", "c"},
			},
		},
		"updatedTime": "2020-01-01T00:00:00Z",
	}
	fieldName := ""
	got := map2BsonD(testMap, fieldName)
	expected := bson.D{
		bson.E{Key: "state.onboarding.completed", Value: true},
		bson.E{Key: "state.onboarding.companySize", Value: "1-10"},
		bson.E{Key: "state.onboarding.interests", Value: []string{"a", "b", "c"}},
		bson.E{Key: "updatedTime", Value: "2020-01-01T00:00:00Z"},
	}

	if !reflect.DeepEqual(expected, got) {
		t.Errorf("map2BsonD() = %v, expected %v", got, expected)
	}
}
