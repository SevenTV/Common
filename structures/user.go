package structures

import (
	"context"

	"github.com/SevenTV/Common/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserBuilder struct {
	User *User
}

// FetchByID: Get an emote by its ID
func (b UserBuilder) FetchByID(ctx context.Context, id primitive.ObjectID) (*UserBuilder, error) {
	doc := mongo.Collection(mongo.CollectionNameUsers).FindOne(ctx, bson.M{
		"_id": id,
	})
	if err := doc.Err(); err != nil {
		return nil, err
	}

	var user *User
	if err := doc.Decode(&user); err != nil {
		return nil, err
	}

	b.User = user
	return &b, nil
}

type User struct {
	ID     primitive.ObjectID `json:"id" bson:"_id"`
	Email  string             `json:"email" bson:"email"`
	Emotes []UserEmote        `json:"emotes" bson:"emotes"`
}

type UserEmote struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	Alias     string             `json:"alias,omitempty" bson:"alias,omitempty"`
	ZeroWidth bool               `json:"zero_width,omitempty" bson:"zero_width,omitempty"`
}
