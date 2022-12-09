package legs

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
)

func CommitPayloadToMetaCache(ctx context.Context, providerID string, collectionName string, data ipld.Node, client *mongo.Client) error {
	_, span := otel.Tracer(TracerName).Start(ctx, "CommitPayloadToMetaCache(Mongo)")
	defer span.End()
	dataBuffer := bytes.NewBuffer(nil)
	err := dagjson.Encode(data, dataBuffer)
	if err != nil {
		return err
	}
	dataJson := map[string]interface{}{}
	err = json.Unmarshal(dataBuffer.Bytes(), &dataJson)
	if err != nil {
		return err
	}

	locationCollection := client.Database(providerID).Collection(collectionName)
	result, err := locationCollection.InsertOne(context.TODO(), dataJson)
	if err != nil {
		return err
	}
	logger.Debugf("insert a doc into mongo, ID: %s", result.InsertedID)

	return nil
}
