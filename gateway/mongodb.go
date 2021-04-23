package gateway

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

type MongoClient struct {
	Client *mongo.Client
}

func New(url string) *MongoClient {
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return &MongoClient{Client: client}
}

func (mc *MongoClient) Close() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	mc.Client.Disconnect(ctx)
}

func (mc *MongoClient) ListDB() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	databases, err := mc.Client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)
}

func (mc *MongoClient) GetDocByID(db string, col string, result interface{}, field string, value interface{}) error {
	database := mc.Client.Database(db)
	collection := database.Collection(col)

	filter := bson.D{{Key: field, Value: value}}

	if field == "" {
		filter = bson.D{}
	}

	err := collection.FindOne(context.TODO(), filter).Decode(result)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (mc *MongoClient) DeleteDocByID(db string, col string, field string, value interface{}) (bool, error) {
	database := mc.Client.Database(db)
	collection := database.Collection(col)

	filter := bson.D{{Key: field, Value: value}}

	if field == "" {
		filter = bson.D{}
	}

	del, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return del.DeletedCount == 1, nil
}

func (mc *MongoClient) ListDocs(db string, col string, results interface{}, field string, value interface{}, limit int64, offset int64) error {
	database := mc.Client.Database(db)
	collection := database.Collection(col)

	filter := bson.D{{Key: field, Value: value}}

	if field == "" {
		filter = bson.D{}
	}

	op := options.Find()
	op.SetSkip(offset)
	op.SetLimit(limit)

	cursor, err := collection.Find(context.TODO(), filter, op)
	if err != nil {
		log.Println(err)
		return err
	}

	cursor.All(context.TODO(), results)

	return nil
}

func (mc *MongoClient) InsertDoc(db string, col string, doc interface{}) error {
	database := mc.Client.Database(db)
	collection := database.Collection(col)

	_, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (mc *MongoClient) UpdateDocByID(db string, col string, field string, value interface{}, updateKey string, updateValue interface{}) (bool, error) {
	database := mc.Client.Database(db)
	collection := database.Collection(col)

	filter := bson.M{field: value}
	updatedDoc := bson.D{{"$set", bson.D{{updateKey, updateValue}}}}

	if field == "" {
		filter = bson.M{}
	}

	res, err := collection.UpdateOne(context.TODO(), filter, updatedDoc)
	if err != nil {
		log.Println(err)
		return false, err
	}

	return res.ModifiedCount == 1, nil
}