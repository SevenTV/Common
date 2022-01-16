package mongo

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
)

func collSync(ctx context.Context, inst Instance) error {
	for _, col := range collections {
		if err := inst.RawDatabase().RunCommand(ctx, bson.D{
			{Key: "collMod", Value: col.Name},
			{Key: "validator", Value: bson.M{"$jsonSchema": col.Validator}},
			{Key: "validationAction", Value: "error"},
			{Key: "validationLevel", Value: "strict"},
		}).Err(); err != nil {
			logrus.WithField("collection", col.Name).WithError(err).Error("mongo, failed to update collection validator")
		}
	}

	return nil
}

type collectionRef struct {
	Name      string
	Validator jsonSchema
	Indexes   []IndexModel
}

type jsonSchema struct {
	BSONType   BSONType               `json:"bsonType" bson:"bsonType"`
	Properties map[string]*jsonSchema `json:"properties,omitempty" bson:"properties,omitempty"`
	// A title for the validator
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	// A list of fields that are required to be present in the collection's documents
	Required []string `json:"required,omitempty" bson:"required,omitempty"`

	Pattern string   `json:"pattern,omitempty" bson:"pattern,omitempty"`
	Enum    []string `json:"enum,omitempty" bson:"enum,omitempty"`
	Maximum *int64   `json:"maximum,omitempty" bson:"maximum,omitempty"`
	Minimum *int64   `json:"minimum,omitempty" bson:"minimum,omitempty"`
	// Indicates the maximum length of the field
	MaxLength *int64 `json:"maxLength,omitempty" bson:"maxLength,omitempty"`
	// Indicates the minimum length of the field
	MinLength *int64 `json:"minLength,omitempty" bson:"minLength,omitempty"`
	// Indicates the maximum length of array
	MaxItems *int64 `json:"maxItems,omitempty" bson:"maxItems,omitempty"`
	// Indicates the minimum length of array
	MinItems *int64 `json:"minItems,omitempty" bson:"minItems,omitempty"`
	// Field must be a multiple of this value
	MultipleOf *int64 `json:"multipleOf,omitempty" bson:"multipleOf,omitempty"`

	// 	Field must match all specified schemas
	AllOf []*jsonSchema `json:"allOf,omitempty" bson:"allOf,omitempty"`
	// Field must match at least one of the specified schemas
	AnyOf []*jsonSchema `json:"anyOf,omitempty" bson:"anyOf,omitempty"`
	// Field must match exactly one of the specified schemas
	OneOf []*jsonSchema `json:"oneOf,omitempty" bson:"oneOf,omitempty"`
	// Field must not match the schema
	Not *jsonSchema `json:"not,omitempty" bson:"not,omitempty"`
	// Must be either a valid JSON Schema, or an array of valid JSON Schemas
	Items []*jsonSchema `json:"items,omitempty" bson:"items,omitempty"`
}

type BSONType string

const (
	BSONTypeDouble     BSONType = "double"
	BSONTypeString     BSONType = "string"
	BSONTypeObject     BSONType = "object"
	BSONTypeArray      BSONType = "array"
	BSONTypeBinary     BSONType = "binData"
	BSONTypeObjectId   BSONType = "objectId"
	BSONTypeBoolean    BSONType = "bool"
	BSONTypeDate       BSONType = "date"
	BSONTypeNull       BSONType = "null"
	BSONTypeRegular    BSONType = "regex"
	BSONTypeJavaScript BSONType = "javascript"
	BSONTypeInt32      BSONType = "int"
	BSONTypeTimestamp  BSONType = "timestamp"
	BSONTypeInt64      BSONType = "long"
	BSONTypeDecimal128 BSONType = "decimal"
	BSONTypeMinkey     BSONType = "minKey"
	BSONTypeMaxkey     BSONType = "maxKey"
)
