package model

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const defaultMiniAppLanguage = "en"

type PaymentService string

const (
	PaymentServiceTON       PaymentService = "ton"
	PaymentServiceWayForPay PaymentService = "wayforpay"
)

type MiniApp struct {
	bun.BaseModel `bun:"table:mini_apps"`

	ID                    uuid.UUID        `bun:"id,pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	PlanID                PlanID           `bun:"plan_id,type:varchar(100),notnull" json:"-"`
	BotID                 int64            `bun:"bot_id,type:bigint,nullzero" json:"bot_id"`
	BotToken              string           `bun:"bot_token,type:varchar(100),notnull" json:"-"`
	OwnerTelegramID       int64            `bun:"owner_telegram_id,type:bigint,notnull" json:"-"`
	Name                  string           `bun:"name,type:varchar(100),notnull" json:"name"`
	Logo                  string           `bun:"logo,type:varchar(255),notnull" json:"logo"`
	LogoSize              int64            `bun:"logo_size,type:bigint,notnull" json:"logo_size"`
	TeacherBio            string           `bun:"teacher_bio,type:text,notnull" json:"teacher_bio"`
	TeacherLinks          []string         `bun:"teacher_links,type:varchar(255)[],notnull" json:"teacher_links"`
	TeacherAvatar         string           `bun:"teacher_avatar,type:varchar(255),notnull" json:"teacher_avatar"`
	TeacherAvatarSize     int64            `bun:"teacher_avatar_size,type:bigint,notnull" json:"teacher_avatar_size"`
	ColorTheme            json.RawMessage  `bun:"color_theme,type:jsonb,notnull,default:'{}'" json:"color_theme"`
	Language              string           `bun:"language,type:varchar(100),notnull" json:"language"`
	PaymentMetadata       json.RawMessage  `bun:"payment_metadata,type:jsonb,nullzero" json:"-"`
	ActivePaymentServices []PaymentService `bun:"active_payment_services,type:payment_service[],notnull,default:'{}'" json:"active_payment_services"`
	URL                   string           `bun:"url,type:varchar(255),notnull" json:"url"`
	Support               string           `bun:"support,type:varchar(255),notnull" json:"support"`
	Analytics             json.RawMessage  `bun:"analytics,type:jsonb,notnull" json:"analytics"`
	IsActive              bool             `bun:"is_active,type:boolean,notnull" json:"is_active"`

	StorageSize      int64  `bun:"storage_size,type:bigint,notnull" json:"-"`
	TotalProducts    int64  `bun:"total_products,type:bigint,notnull" json:"-"`
	TotalStudents    int64  `bun:"total_students,type:bigint,notnull" json:"-"`
	TotalEvents      int64  `bun:"total_events,type:bigint,notnull" json:"-"`
	MaxStorageSize   *int64 `bun:"max_storage_size,type:bigint,nullzero" json:"-"`
	MaxTotalProducts *int64 `bun:"max_total_products,type:bigint,nullzero" json:"-"`
	MaxTotalStudents *int64 `bun:"max_total_students,type:bigint,nullzero" json:"-"`
	MaxTotalEvents   *int64 `bun:"max_total_events,type:bigint,nullzero" json:"-"`

	DeletedAt *time.Time `bun:"deleted_at,type:timestamptz,nullzero" json:"deleted_at,omitempty"`
	UpdatedAt time.Time  `bun:"updated_at,type:timestamptz,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time  `bun:"created_at,type:timestamptz,notnull,default:current_timestamp" json:"created_at"`

	Owner         *User       `bun:"rel:has-one,join:id=mini_app_id,join:owner_telegram_id=telegram_id" json:"owner,omitempty"`
	Slides        []*Material `bun:"rel:has-many,join:id=mini_app_id" json:"slides,omitempty"`
	TOS           []*Material `bun:"rel:has-many,join:id=mini_app_id" json:"tos,omitempty"`
	PrivacyPolicy []*Material `bun:"rel:has-many,join:id=mini_app_id" json:"privacy_policy,omitempty"`

	Products []*Product `bun:"rel:has-many,join:id=mini_app_id" json:"products,omitempty"`
}

func NewMiniApp() *MiniApp {
	now := time.Now().UTC()
	return &MiniApp{
		ID:        uuid.New(),
		UpdatedAt: now,
		CreatedAt: now,
	}
}

type CreateMiniAppRequest struct {
	BotToken string `json:"bot_token"`
	InitData string `json:"init_data"`

	Name string `json:"name"`
}

func (r *CreateMiniAppRequest) Validate() error {
	if r.InitData == "" {
		return fmt.Errorf("empty init_data")
	}
	if r.Name == "" {
		r.Name = "temp_" + rand.Text()
	}

	return nil
}

func (r *CreateMiniAppRequest) ToMiniApp(botID, ownerID int64) *MiniApp {
	now := time.Now().UTC()
	return &MiniApp{
		ID:              uuid.New(),
		PlanID:          DefaultPlanID,
		BotID:           botID,
		BotToken:        r.BotToken,
		OwnerTelegramID: ownerID,
		Name:            r.Name,
		TeacherLinks:    []string{},
		Language:        defaultMiniAppLanguage,
		IsActive:        true,

		UpdatedAt: now,
		CreatedAt: now,
	}
}

type EditMiniAppAccountRequest struct {
	BotToken        string `json:"bot_token"`
	MiniAppName     string `json:"mini_app_name"`
	MiniAppURL      string `json:"mini_app_url"`
	MiniAppIsActive bool   `json:"mini_app_is_active"`

	ActivePaymentServices    []PaymentService          `json:"active_payment_services"`
	PaymentMetadataTON       *PaymentMetadataTON       `json:"payment_metadata_ton"`
	PaymentMetadataWayForPay *PaymentMetadataWayForPay `json:"payment_metadata_wayforpay"`

	TeacherFirstName    string          `json:"teacher_first_name"`
	TeacherLastName     string          `json:"teacher_last_name"`
	TeacherLanguage     string          `json:"teacher_language"`
	TeacherBio          string          `json:"teacher_bio"`
	TeacherLinks        []string        `json:"teacher_links"`
	TeacherColorTheme   json.RawMessage `json:"teacher_color_theme"`
	TeacherDeleteAvatar bool            `json:"teacher_delete_avatar"`
}

type PaymentMetadata struct {
	PaymentMetadataTON
	PaymentMetadataWayForPay
}
type PaymentMetadataWayForPay struct {
	WayForPayLogin      string `json:"wayforpay_login,omitempty"`
	WayForPaySecretKey  string `json:"wayforpay_secret_key,omitempty"`
	WayForPayDomainName string `json:"wayforpay_domain_name,omitempty"`
}
type PaymentMetadataTON struct {
	TONAddress string `json:"ton_address,omitempty"`
}

func (r *EditMiniAppAccountRequest) UpdateMiniApp(miniApp *MiniApp, botID int64) (bool, error) {
	isChanged := false

	if botID != 0 && botID != miniApp.BotID {
		miniApp.BotID = botID
		isChanged = true
	}
	if r.MiniAppName != miniApp.Name {
		miniApp.Name = r.MiniAppName
		isChanged = true
	}

	if r.TeacherBio != miniApp.TeacherBio {
		miniApp.TeacherBio = r.TeacherBio
		isChanged = true
	}
	if slices.Compare(r.TeacherLinks, miniApp.TeacherLinks) != 0 {
		miniApp.TeacherLinks = r.TeacherLinks
		isChanged = true
	}

	for _, s := range r.ActivePaymentServices {
		switch s {
		case PaymentServiceTON:
		case PaymentServiceWayForPay:
		default:
			return false, fmt.Errorf("invalid payment service name: %q", s)
		}
	}
	if slices.Compare(r.ActivePaymentServices, miniApp.ActivePaymentServices) != 0 {
		miniApp.ActivePaymentServices = r.ActivePaymentServices
		isChanged = true
	}

	if r.MiniAppURL != miniApp.URL {
		miniApp.URL = r.MiniAppURL
		isChanged = true
	}

	if r.MiniAppIsActive != miniApp.IsActive {
		miniApp.IsActive = r.MiniAppIsActive
		isChanged = true
	}

	miniApp.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}

func (r *EditMiniAppAccountRequest) UpdateTeacher(u *User) (bool, error) {
	isChanged := false

	if r.TeacherFirstName != u.FirstName {
		u.FirstName = r.TeacherFirstName
		isChanged = true
	}
	if r.TeacherLastName != u.LastName {
		u.LastName = r.TeacherLastName
		isChanged = true
	}
	if r.TeacherLanguage != u.Language {
		u.Language = r.TeacherLanguage
		isChanged = true
	}
	if string(r.TeacherColorTheme) != string(u.ColorTheme) {
		u.ColorTheme = r.TeacherColorTheme
		isChanged = true
	}

	u.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}

type EditMiniAppBrandingRequest struct {
	ColorTheme *MiniAppColorTheme `json:"color_theme"`
	Language   string             `json:"language"`
	Support    string             `json:"support"`
	DeleteLogo bool               `json:"delete_logo"`
}

type MiniAppColorTheme struct {
	AccentColor string `json:"accent_color"`
}

func (r *EditMiniAppBrandingRequest) UpdateMiniApp(miniApp *MiniApp) (bool, error) {
	isChanged := false

	if r.ColorTheme != nil {
		colorTheme, err := json.Marshal(r.ColorTheme)
		if err != nil {
			return false, fmt.Errorf("json.Marshal(r.ColorTheme): %w", err)
		}
		if string(colorTheme) != string(miniApp.ColorTheme) {
			miniApp.ColorTheme = colorTheme
			isChanged = true
		}
	}

	if r.Language != miniApp.Language {
		miniApp.Language = r.Language
		isChanged = true
	}

	if r.Support != miniApp.Support {
		miniApp.Support = r.Support
		isChanged = true
	}

	miniApp.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}

type EditMiniAppAnalyticsRequest struct {
	FacebookPixel   string `json:"facebook_pixel"`
	GTag            string `json:"g_tag"`
	GoogleAnalytics string `json:"google_analytics"`
}

func (r *EditMiniAppAnalyticsRequest) UpdateMiniApp(miniApp *MiniApp) (bool, error) {
	isChanged := false

	analytics, err := json.Marshal(r)
	if err != nil {
		return false, fmt.Errorf("json.Marshal: %w", err)
	}
	if string(analytics) != string(miniApp.Analytics) {
		miniApp.Analytics = analytics
		isChanged = true
	}

	miniApp.UpdatedAt = time.Now().UTC()

	return isChanged, nil
}

type EditSlidesRequest struct {
	SlidesOrder []uuid.UUID `json:"slides_order"`

	Index       int64  `json:"new_slide_index"`
	Title       string `json:"new_slide_title"`
	Description string `json:"new_slide_description"`
	URL         string `json:"new_slide_url"`
}

func (r *EditSlidesRequest) Validate() error {
	slides := make(map[uuid.UUID]struct{}, len(r.SlidesOrder))

	for _, s := range r.SlidesOrder {
		if s == uuid.Nil {
			return fmt.Errorf("invalid slide id")
		}

		if _, ok := slides[s]; ok {
			return fmt.Errorf("duplicated slide id")
		}

		slides[s] = struct{}{}
	}

	return nil
}

func (r *EditSlidesRequest) ToMaterial(
	miniAppID uuid.UUID,
	filename string,
	size int64,
) *Material {

	p := NewMaterial()

	p.MiniAppID = miniAppID
	p.Index = r.Index
	p.Category = MaterialCategorySlides
	p.ContentType = MaterialTypePicture
	p.Title = r.Title
	p.Description = r.Description
	p.URL = r.URL

	p.Filename = filename
	p.Size = size

	return p
}

type MiniAppInfo struct {
	StorageSize      int64  `bun:"storage_size" json:"storage_size"`
	TotalProducts    int64  `bun:"total_products" json:"total_products"`
	TotalStudents    int64  `bun:"total_students" json:"total_students"`
	TotalEvents      int64  `bun:"total_events" json:"total_events"`
	MaxStorageSize   *int64 `bun:"max_storage_size" json:"max_storage_size"`
	MaxTotalProducts *int64 `bun:"max_total_products" json:"max_total_products"`
	MaxTotalStudents *int64 `bun:"max_total_students" json:"max_total_students"`
	MaxTotalEvents   *int64 `bun:"max_total_events" json:"max_total_events"`
}

type ListMiniAppsRequest struct {
	InitData string `json:"init_data"`
}
