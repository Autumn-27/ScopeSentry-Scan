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

// NewMongoDbConnect 连接到MongoDB并返回一个MongoDBClient实例
func NewMongoDbConnect(Username string, Password string, IP string, Port string, Database string) (*Client, error) {
	encodedPassword := url.QueryEscape(Password)
	connectionURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/?maxPoolSize=50", Username, encodedPassword, IP, Port)
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
		return nil, err
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		fmt.Printf("mongodb ping error: %v", err)
		return nil, err
	}
	db := client.Database(Database)
	return &Client{client: client, database: db}, nil
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
	return collection.InsertMany(context.Background(), documents)
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
