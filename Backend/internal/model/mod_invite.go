package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PermissionName string

const (
	PermissionStudentManagement      PermissionName = "Student Management"
	PermissionProductsControl        PermissionName = "Products Control"
	PermissionSubscriptionManagement PermissionName = "Subscription Management"
	PermissionAnalytics              PermissionName = "Analytics"
	PermissionBranding               PermissionName = "Branding"
	PermissionAccountSettings        PermissionName = "Account Settings"
	PermissionStudentInteraction     PermissionName = "Student Interaction"
)

type ModInvite struct {
	bun.BaseModel `bun:"table:mod_invites"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	MiniAppID uuid.UUID `bun:"mini_app_id,type:uuid,notnull" json:"mini_app_id"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,nullzero" json:"user_id"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	MiniApp     *User                  `bun:"rel:belongs-to,join:mini_app_id=id" json:"mini_app,omitempty"`
	User        *User                  `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
	Permissions []*ModInvitePermission `bun:"rel:has-many,join:id=invite_id" json:"permissions,omitempty"`
}

type ModInvitePermission struct {
	bun.BaseModel `bun:"table:mod_invite_permissions"`

	InviteID       uuid.UUID      `bun:"invite_id,pk,type:uuid" json:"invite_id"`
	PermissionName PermissionName `bun:"permission_name,pk,type:varchar(60)" json:"permission_name"`

	Permission *Permission `bun:"rel:belongs-to,join:permission_name=name" json:"permission"`
}

type Permission struct {
	bun.BaseModel `bun:"table:permissions"`

	Name        PermissionName `bun:"name,pk,type:varchar(60)" json:"name"`
	Description string         `bun:"description,type:text" json:"description"`
}

func NewModInvite(miniAppID uuid.UUID) *ModInvite {
	now := time.Now()
	return &ModInvite{
		ID:        uuid.New(),
		MiniAppID: miniAppID,
		UpdatedAt: now,
		CreatedAt: now,
	}
}

func NewModInvitePermission(inviteID uuid.UUID, permissionName PermissionName) *ModInvitePermission {
	return &ModInvitePermission{
		InviteID:       inviteID,
		PermissionName: permissionName,
	}
}

type CreateModInviteRequest struct {
	Permissions []PermissionName `json:"permissions"`
}

func (r *CreateModInviteRequest) Validate() error {
	if len(r.Permissions) == 0 {
		return fmt.Errorf("no permissions provided for mod invite")
	}

	for _, p := range r.Permissions {
		switch p {
		case PermissionStudentManagement:
		case PermissionProductsControl:
		case PermissionSubscriptionManagement:
		case PermissionAnalytics:
		case PermissionBranding:
		case PermissionAccountSettings:
		case PermissionStudentInteraction:
		default:
			return fmt.Errorf("unsupported permission: %q", p)
		}
	}

	return nil
}

type EditModInviteRequest struct {
	Permissions []PermissionName `json:"permissions"`
}

func (r *EditModInviteRequest) Validate() error {
	if len(r.Permissions) == 0 {
		return fmt.Errorf("no permissions provided for mod invite")
	}

	for _, p := range r.Permissions {
		switch p {
		case PermissionStudentManagement:
		case PermissionProductsControl:
		case PermissionSubscriptionManagement:
		case PermissionAnalytics:
		case PermissionBranding:
		case PermissionAccountSettings:
		case PermissionStudentInteraction:
		default:
			return fmt.Errorf("unsupported permission: %q", p)
		}
	}

	return nil
}

type FilterModInvitesRequest struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}
