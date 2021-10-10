package structures

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserBuilder struct {
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
