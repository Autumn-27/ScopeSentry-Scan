// mongodb-------------------------------------
// @file      : mongodb.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 21:53
// -------------------------------------------

package mongodb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
}

var MongodbClient *Client

//// NewMongoDbConnect 连接到MongoDB并返回一个MongoDBClient实例
//func NewMongoDbConnect(Username string, Password string, IP string, Port string, Database string) (*Client, error) {
//
//	encodedPassword := url.QueryEscape(Password)
//	connectionURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/?maxPoolSize=50", Username, encodedPassword, IP, Port)
//	clientOptions := options.Client().ApplyURI(connectionURI)
//	var MaxPoolSizevalue uint64 = 50
//	clientOptions.MaxPoolSize = &MaxPoolSizevalue
//	var MaxConnectingValue uint64 = 10
//	clientOptions.MaxConnecting = &MaxConnectingValue
//	var MinPoolSizeValue uint64 = 5
//	clientOptions.MinPoolSize = &MinPoolSizeValue
//	client, err := mongo.Connect(context.Background(), clientOptions)
//	if err != nil {
//		fmt.Printf("mongodb connect error: %v", err)
//		return nil, err
//	}
//
//	err = client.Ping(context.Background(), nil)
//	if err != nil {
//		fmt.Printf("mongodb ping error: %v", err)
//		return nil, err
//	}
//	db := client.Database(Database)
//	return &Client{client: client, database: db}, nil
//}

func Initialize() {
	encodedPassword := url.QueryEscape(global.AppConfig.MongoDB.Password)
	connectionURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/?maxPoolSize=50", global.AppConfig.MongoDB.User, encodedPassword, global.AppConfig.MongoDB.IP, global.AppConfig.MongoDB.Port)
	clientOptions := options.Client().ApplyURI(connectionURI)
	var MaxPoolSizevalue uint64 = 50
	clientOptions.MaxPoolSize = &MaxPoolSizevalue
	var MaxConnectingValue uint64 = 10
	clientOptions.MaxConnecting = &MaxConnectingValue
	var MinPoolSizeValue uint64 = 5
	clientOptions.MinPoolSize = &MinPoolSizeValue
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Printf("mongodb connect error: %v", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		fmt.Printf("mongodb ping error: %v", err)
	}
	db := client.Database(global.AppConfig.MongoDB.Database)

	MongodbClient = &Client{client: client, database: db}
}

// GetCollection 获取指定集合
func (c *Client) GetCollection(collectionName string) *mongo.Collection {
	return c.client.Database("ScopeSentry").Collection(collectionName)
}

// FindAll 查询多个文档
func (c *Client) FindAll(collectionName string, query, selector, result interface{}) error {
	collection := c.GetCollection(collectionName)
	cur, err := collection.Find(context.Background(), query, options.Find().SetProjection(selector))
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	return cur.All(context.Background(), result)
}

func (c *Client) Aggregate(collectionName string, pipeline, result interface{}) error {
	collection := c.GetCollection(collectionName)
	cur, err := collection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	return cur.All(context.Background(), result)
}

func (c *Client) FindOne(collectionName string, query, selector, result interface{}) error {
	collection := c.GetCollection(collectionName)
	return collection.FindOne(context.Background(), query, options.FindOne().SetProjection(selector)).Decode(result)
}

func (c *Client) FindFile(filename string) ([]byte, error) {
	// 使用 GridFS bucket
	bucket, err := gridfs.NewBucket(c.database)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	dStream, err := bucket.OpenDownloadStreamByName(filename)
	if err != nil {
		return nil, err
	}
	defer dStream.Close()

	if _, err := buf.ReadFrom(dStream); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Update 更新单个文档
func (c *Client) Update(collectionName string, selector, update interface{}) (*mongo.UpdateResult, error) {
	collection := c.GetCollection(collectionName)
	return collection.UpdateOne(context.Background(), selector, update)
}

// Upsert 更新或插入文档
func (c *Client) Upsert(collectionName string, selector, update interface{}) (*mongo.UpdateResult, error) {
	collection := c.GetCollection(collectionName)
	opts := options.Update().SetUpsert(true)
	return collection.UpdateOne(context.Background(), selector, update, opts)
}

// UpdateAll 更新多个文档
func (c *Client) UpdateAll(collectionName string, selector, update interface{}) (*mongo.UpdateResult, error) {
	collection := c.GetCollection(collectionName)
	return collection.UpdateMany(context.Background(), selector, update)
}

func (c *Client) InsertOne(collectionName string, document interface{}) (*mongo.InsertOneResult, error) {
	collection := c.GetCollection(collectionName)
	return collection.InsertOne(context.Background(), document)
}

func (c *Client) InsertMany(collectionName string, documents []interface{}) (*mongo.InsertManyResult, error) {
	collection := c.GetCollection(collectionName)
	// 设置 InsertMany 选项，允许非顺序插入
	opts := options.InsertMany().SetOrdered(false)
	return collection.InsertMany(context.Background(), documents, opts)
}

// Close 关闭与MongoDB的连接
func (c *Client) Close() {
	if c.client != nil {
		err := c.client.Disconnect(context.Background())
		if err != nil {
			fmt.Println("Error disconnecting from MongoDB:", err)
			return
		}
		fmt.Println("Disconnected from MongoDB.")
	}
}

func (c *Client) Ping() error {
	if c == nil {
		fmt.Println("MongoDBClient c is nil")
		return errors.New("mongodb client is not initialized")
	}
	err := c.client.Ping(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("MongoDB Ping失败: %v", err)
	}
	return nil
}

func (c *Client) FindFilesByPattern(pattern string) (map[string][]byte, error) {
	// 使用 GridFS bucket
	bucket, err := gridfs.NewBucket(c.database)
	if err != nil {
		return nil, err
	}

	// 创建正则表达式
	regex := primitive.Regex{Pattern: pattern, Options: "i"} // "i" 选项表示不区分大小写

	// 查询 GridFS 的 files 集合
	collection := c.database.Collection("fs.files")
	filter := bson.M{"filename": regex}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	results := make(map[string][]byte)
	for cursor.Next(context.TODO()) {
		var fileInfo bson.M
		if err := cursor.Decode(&fileInfo); err != nil {
			return nil, err
		}

		// 获取文件的 _id 字段
		id, ok := fileInfo["_id"].(primitive.ObjectID)
		if !ok {
			return nil, fmt.Errorf("文件信息中没有 _id 字段")
		}

		// 通过 _id 查找文件内容
		dStream, err := bucket.OpenDownloadStream(id)
		if err != nil {
			return nil, err
		}
		defer dStream.Close()

		var buf bytes.Buffer
		if _, err := buf.ReadFrom(dStream); err != nil {
			return nil, err
		}
		filename := fileInfo["filename"].(string)
		results[filename] = buf.Bytes()
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
