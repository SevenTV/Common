package structures

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Name     string             `json:"name" bson:"name"`
	Position int32              `json:"position" bson:"position"`
	Color    int32              `json:"color" bson:"color"`
	Allowed  int64              `json:"allowed" bson:"allowed"`
	Denied   int64              `json:"denied" bson:"denied"`
	Default  bool               `json:"default,omitempty" bson:"default"`
}

const (
	RolePermissionEmoteCreate          int64 = 1 << iota // 1 - Allows creating emotes
	RolePermissionEmoteEditOwned                         // 2 - Allows editing own emotes
	RolePermissionEmoteEditAll                           // 4 - (Elevated) Allows editing all emotes
	RolePermissionCreateReports                          // 8 - Allows creating reports
	RolePermissionManageReports                          // 16 - (Elevated) Allows managing reports
	RolePermissionBanUsers                               // 32 - (Elevated) Allows banning other users
	RolePermissionAdministrator                          // 64 - (Dangerous, Elevated) GRANTS ALL PERMISSIONS
	RolePermissionManageRoles                            // 128 - (Elevated) Allows managing roles
	RolePermissionManageUsers                            // 256 - (Elevated) Allows managing users
	RolePermissionManageEditors                          // 512 - Allows adding and removing editors from own channel
	RolePermissionEditEmoteGlobalState                   // 1024 - (Elevated) Allows editing the global state of an emote
	RolePermissionEditApplicationMeta                    // 2048 - (Elevated) Allows editing global app metadata, such as the active featured broadcast
	RolePermissionManageEntitlements                     // 4096 - (Elevated) Allows granting and revoking entitlements to and from users
	RolePermissionUseZeroWidthEmote                      // 8192 - Allows zero-width emotes to be enabled
	RolePermissionUseCustomAvatars                       // 16384 - Allows setting a custom avatar

	RolePermissionAll int64 = (1 << iota) - 1
)
