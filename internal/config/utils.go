// config-------------------------------------
// @file      : utils.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/4/8 20:58
// -------------------------------------------

package config

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDictId(tp string, name string) string {
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := mongodb.MongodbClient.FindOne("dictionary", bson.M{"name": name, "category": tp}, bson.M{"_id": 1}, &result)
	if err != nil {
		return ""
	}
	return result.ID.Hex()
}
