package structures

import (
	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicUser struct {
	ID          primitive.ObjectID     `json:"id"`
	UserType    UserType               `json:"type,omitempty"`
	Username    string                 `json:"username"`
	DisplayName string                 `json:"display_name"`
	RoleIDs     []primitive.ObjectID   `json:"roles"`
	Connections []PublicUserConnection `json:"connections"`
}

func (u *User) ToPublic() PublicUser {
	connections := make([]PublicUserConnection, len(u.Connections))
	for i, c := range u.Connections {
		connections[i] = c.ToPublic()
	}

	return PublicUser{
		ID:          u.ID,
		UserType:    u.UserType,
		Username:    u.Username,
		DisplayName: utils.Ternary(u.DisplayName != "", u.DisplayName, u.Username),
		RoleIDs:     u.RoleIDs,
		Connections: connections,
	}
}

type PublicUserConnection struct {
	ID         string                 `json:"id"`
	Platform   UserConnectionPlatform `json:"platform"`
	LinkedAt   int64                  `json:"linked_at"`
	EmoteSetID primitive.ObjectID     `json:"emote_set_id,omitempty"`
}

func (uc *UserConnection[D]) ToPublic() PublicUserConnection {
	return PublicUserConnection{
		ID:         uc.ID,
		Platform:   uc.Platform,
		LinkedAt:   uc.LinkedAt.UnixMilli(),
		EmoteSetID: uc.EmoteSetID,
	}
}
