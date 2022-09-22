package structures

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type UserBuilder struct {
	Update UpdateMap
	User   User

	initial User
	tainted bool
}

// NewUserBuilder: create a new user builder
func NewUserBuilder(user User) *UserBuilder {
	return &UserBuilder{
		Update:  UpdateMap{},
		User:    user,
		initial: user,
	}
}

// Initial returns a pointer to the value first passed to this Builder
func (ub *UserBuilder) Initial() *User {
	return &ub.initial
}

// IsTainted returns whether or not this Builder has been mutated before
func (ub *UserBuilder) IsTainted() bool {
	return ub.tainted
}

// MarkAsTainted taints the builder, preventing it from being mutated again
func (ub *UserBuilder) MarkAsTainted() {
	ub.tainted = true
}

// SetUsername: set the username for the user
func (ub *UserBuilder) SetUsername(username string) *UserBuilder {
	ub.User.Username = username
	ub.Update.Set("username", username)

	return ub
}

func (ub *UserBuilder) SetDisplayName(s string) *UserBuilder {
	ub.User.DisplayName = s
	ub.Update.Set("display_name", s)
	return ub
}

func (ub *UserBuilder) SetDiscriminator(discrim string) *UserBuilder {
	if discrim == "" {
		for i := 0; i < 4; i++ {
			discrim += strconv.Itoa(rand.Intn(9))
		}
	}

	ub.User.Discriminator = discrim
	ub.Update.Set("discriminator", discrim)
	return ub
}

// SetEmail: set the email for the user
func (ub *UserBuilder) SetEmail(email string) *UserBuilder {
	ub.User.Email = email
	ub.Update.Set("email", email)

	return ub
}

func (ub *UserBuilder) SetAvatarID(url string) *UserBuilder {
	ub.User.AvatarID = url
	ub.Update.Set("avatar_url", url)

	return ub
}

func (ub *UserBuilder) GetConnection(p UserConnectionPlatform, id ...string) (*UserConnectionBuilder[bson.Raw], int) {
	// Filter by ID?
	filterID := ""
	if len(id) > 0 {
		filterID = id[0]
	}

	// Find connection
	var conn UserConnection[bson.Raw]

	ind := -1
	for i, c := range ub.User.Connections {
		if p != "" && c.Platform != p {
			continue
		}
		if filterID != "" && c.ID != filterID {
			continue
		}
		conn = c
		ind = i
		break
	}

	return NewUserConnectionBuilder(conn), ind
}

func (ub *UserBuilder) AddConnection(conn UserConnection[bson.Raw]) *UserBuilder {
	for _, c := range ub.User.Connections {
		if c.ID == conn.ID {
			return ub // connection already exists.
		}
	}

	ub.User.Connections = append(ub.User.Connections, conn)
	ub.Update = ub.Update.AddToSet("connections", conn)

	return ub
}

func (ub *UserBuilder) RemoveConnection(id string) (*UserBuilder, int) {
	ind := -1
	for i := range ub.User.Connections {
		if ub.User.Connections[i].ID == "" {
			continue
		}
		if ub.User.Connections[i].ID != id {
			continue
		}
		ind = i
		break
	}
	if ind == -1 {
		return ub, ind // did not find index
	}

	copy(ub.User.Connections[ind:], ub.User.Connections[ind+1:])
	ub.User.Connections = ub.User.Connections[:len(ub.User.Connections)-1]
	ub.Update.Pull("connections", bson.M{"id": id})

	return ub, ind
}

func (ub *UserBuilder) AddEditor(id ObjectID, permissions UserEditorPermission, visible bool) (UserEditor, int, *UserBuilder) {
	i := 0
	ed := UserEditor{}

	for i, ed = range ub.User.Editors {
		if ed.ID == id {
			return ed, i, ub // editor already added.
		}
	}

	ed = UserEditor{
		ID:          id,
		Permissions: permissions,
		Visible:     visible,
		AddedAt:     time.Now(),
	}
	ub.User.Editors = append(ub.User.Editors, ed)
	ub.Update.AddToSet("editors", ed)
	return ed, i + 1, ub
}

func (ub *UserBuilder) UpdateEditor(id ObjectID, permissions UserEditorPermission, visible bool) (UserEditor, int, *UserBuilder) {
	i := 0
	ed := UserEditor{}

	for i, ed = range ub.User.Editors {
		if ed.ID == id {
			break
		}
	}
	if ed.ID.IsZero() {
		return ed, -1, ub
	}

	v := ub.User.Editors[i]
	v.Permissions = permissions
	v.Visible = visible
	ub.Update.Set(fmt.Sprintf("editors.%d", i), v)
	return ed, i, ub
}

func (ub *UserBuilder) RemoveEditor(id ObjectID) (UserEditor, int, *UserBuilder) {
	i := 0
	ed := UserEditor{}

	for i, ed = range ub.User.Editors {
		if ub.User.Editors[i].ID != id {
			continue
		}
		break
	}
	if ed.ID.IsZero() {
		return ed, -1, ub
	}

	copy(ub.User.Editors[i:], ub.User.Editors[i+1:])
	ub.User.Editors[len(ub.User.Editors)-1] = UserEditor{}
	ub.User.Editors = ub.User.Editors[:len(ub.User.Editors)-1]
	ub.Update.Pull("editors", bson.M{"id": id})
	return ed, i, ub
}
