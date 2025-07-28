package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	initdata "github.com/telegram-mini-apps/init-data-golang"
	"github.com/uptrace/bun"
)

type UserRole string

const (
	UserRoleOwner     = "owner"
	UserRoleModerator = "moderator"
	UserRoleStudent   = "student"
)

type User struct {
	bun.BaseModel `bun:"table:users"`

	ID               uuid.UUID       `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	MiniAppID        uuid.UUID       `bun:"mini_app_id,type:uuid" json:"mini_app_id"`
	Role             UserRole        `bun:"role,type:user_role,notnull" json:"-"`
	TelegramID       int64           `bun:"telegram_id,type:bigint,notnull" json:"telegram_id"`
	TelegramUsername string          `bun:"telegram_username,type:varchar(32),notnull" json:"telegram_username"`
	FirstName        string          `bun:"first_name,type:varchar(100),notnull" json:"first_name"`
	LastName         string          `bun:"last_name,type:varchar(100),notnull" json:"last_name"`
	Avatar           string          `bun:"avatar,type:varchar(255),notnull" json:"avatar"`
	AvatarSize       int64           `bun:"avatar_size,type:bigint,notnull" json:"avatar_size"`
	Language         string          `bun:"language,type:varchar(100),notnull" json:"language"`
	ColorTheme       json.RawMessage `bun:"color_theme,type:jsonb,notnull" json:"color_theme"`
	IsActive         bool            `bun:"is_active,type:boolean,notnull" json:"is_active"`

	UpdatedAt time.Time `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	ProductAccess []*ProductAccess `bun:"rel:has-many,join:id=user_id" json:"product_access,omitempty"`
}

func NewUser() *User {
	now := time.Now().UTC()
	return &User{
		ID:        uuid.New(),
		UpdatedAt: now,
		CreatedAt: now,
	}
}

func defaultTelegramUsername(telegramID int64) string {
	return "user_" + strconv.Itoa(int(telegramID))
}

func (u *User) UpdateOnSignIn(newUser *User) (isChanged bool) {
	// Role could be changed on sign-in using invites.
	if u.Role != newUser.Role {
		u.Role = newUser.Role
		isChanged = true
	}

	// Do not include other TG data so that can be changed only manually after creation.

	if u.TelegramID != newUser.TelegramID {
		u.TelegramID = newUser.TelegramID
		isChanged = true
	}

	if newUser.TelegramUsername == "" {
		newUser.TelegramUsername = defaultTelegramUsername(u.TelegramID)
	}

	if u.TelegramUsername != newUser.TelegramUsername {
		u.TelegramUsername = newUser.TelegramUsername
		isChanged = true
	}

	return isChanged
}

type SignInJWTResp struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type SignInAdminRequest struct {
	InitData string `json:"init_data"`
}

type SignInModRequest struct {
	MiniAppID uuid.UUID `json:"mini_app_id"`
	InviteID  uuid.UUID `json:"invite_id"`
	InitData  string    `json:"init_data"`
}

func (r *SignInModRequest) Validate() error {
	if r.InitData == "" {
		return fmt.Errorf("invalid init_data")
	}
	if r.InviteID == uuid.Nil && r.MiniAppID == uuid.Nil {
		return fmt.Errorf("invite_id nor mini_app_id was provideds")
	}
	if r.InviteID != uuid.Nil && r.MiniAppID != uuid.Nil {
		return fmt.Errorf("both invite_id and mini_app_id was provideds")
	}

	return nil
}

type SignInRequest struct {
	MiniAppName string `json:"mini_app_name"`
	InitData    string `json:"init_data"`
}

func (r *SignInRequest) Validate() error {
	if r.MiniAppName == "" {
		return fmt.Errorf("invalid mini-app name")
	}
	if r.InitData == "" {
		return fmt.Errorf("invalid init_data")
	}

	return nil
}

type SignInWithInviteRequest struct {
	MiniAppName string    `json:"mini_app_name"`
	InitData    string    `json:"init_data"`
	InviteID    uuid.UUID `json:"invite_id"`
}

func (r *SignInWithInviteRequest) Validate() error {
	if r.MiniAppName == "" {
		return fmt.Errorf("invalid mini-app name")
	}
	if r.InitData == "" {
		return fmt.Errorf("invalid init_data")
	}
	if r.InviteID == uuid.Nil {
		return fmt.Errorf("invalid invite_id")
	}

	return nil
}

func NewSignInWithTelegramMiniApp(
	initData *initdata.InitData,
	miniAppID uuid.UUID,
	role UserRole,
) *User {

	u := NewUser()

	u.MiniAppID = miniAppID
	u.TelegramID = initData.User.ID
	u.TelegramUsername = initData.User.Username
	u.FirstName = initData.User.FirstName
	u.LastName = initData.User.LastName
	u.IsActive = true
	u.Role = role

	if u.TelegramUsername == "" {
		u.TelegramUsername = defaultTelegramUsername(u.TelegramID)
	}

	return u
}

type EditUserRequest struct {
	FirstName    string          `json:"first_name"`
	LastName     string          `json:"last_name"`
	Language     string          `json:"language"`
	ColorTheme   json.RawMessage `json:"color_theme"`
	DeleteAvatar bool            `json:"delete_avatar"`
}

func (r *EditUserRequest) UpdateUser(u *User) (bool, error) {
	isChanged := false

	if r.FirstName != u.FirstName {
		u.FirstName = r.FirstName
		isChanged = true
	}
	if r.LastName != u.LastName {
		u.LastName = r.LastName
		isChanged = true
	}
	if r.Language != u.Language {
		u.Language = r.Language
		isChanged = true
	}
	if string(r.ColorTheme) != string(u.ColorTheme) {
		u.ColorTheme = r.ColorTheme
		isChanged = true
	}

	u.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}

type ListBannedUserRequest struct {
	Limit  uint `json:"limit"`
	Offset uint `json:"offset"`
}
