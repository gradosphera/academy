package v1

import "strings"

// Limit for requests with pagination.
const (
	defaultLimit = 100
	maxLimit     = 1000
)

// Mini-App limits.
const (
	miniAppLogoSizeLimit = 5_000_000

	slidesNumberLimit = 5
	slidesSizeLimit   = 5_000_000

	slideTitleLimit       = 55
	slideDescriptionLimit = 75

	userTelegramUsernameLimit = 32
	userNameLimit             = 30
	userBioLimit              = 350
	userLinksNumberLimit      = 8
	privacyPolicySizeLimit    = 3_000_000
	tosSizeLimit              = 3_000_000
)

// User limits.
const (
	ownerAvatarSizeLimit = 4_000_000
	avatarSizeLimit      = 3_000_000
)

// Product limits.
const (
	productTitleLimit       = 55
	productContentTypeLimit = 20
	productDescriptionLimit = 400
	productCoverSizeLimit   = 5_000_000

	productLevelNameLimit = 45
)

// Lesson limits.
const (
	moduleNameLimit = 30

	lessonTitleLimit     = 35
	lessonCoverSizeLimit = 5_000_000

	videoLessonDescriptionLimit = 450
	videoLessonSizeLimit        = 4_001_000_000

	audioLessonDescriptionLimit = 100
	audioLessonSizeLimit        = 501_000_000

	// eventLessonDescriptionLimit = 450 // Same as videoLessonDescriptionLimit.
	// eventLessonSizeLimit        = 4_001_000_000 // Same as videoLessonSizeLimit.

	materialSizeLimit  = 51_000_000
	materialLinkLimit  = 5 // Checked on postgres using check_materials_limit function.
	materialTitleLimit = 35

	bonusMaterialSizeLimit  = 51_000_000
	bonusMaterialLinkLimit  = 5 // Checked on postgres using check_materials_limit function.
	bonusMaterialTitleLimit = 35
)

var allowedVideoExt map[string]struct{} = map[string]struct{}{
	".mp4": {},
}
var allowedAudioExt map[string]struct{} = map[string]struct{}{
	".mp3": {},
}
var allowedMaterialExt map[string]struct{} = map[string]struct{}{
	".zip":  {},
	".txt":  {},
	".pdf":  {},
	".doc":  {},
	".docx": {},
	".epub": {},
	".ppt":  {},
	".pptx": {},
	".pps":  {},
	".xls":  {},
	".xlsx": {},
}

func isPictureAllowed(contentType, fileExt string) bool {
	switch strings.ToLower(fileExt) {
	case ".png", ".jpeg", ".jpg":
	default:
		return false
	}

	switch contentType {
	case "", "image/png", "image/jpeg":
	default:
		return false
	}

	return true
}

// Limits on user uploaded homework materials.
const (
	submitLessonFileSizeLimit = 51_000_000
	submitLessonFilesLimit    = 5
	submitLessonLinksLimit    = 5
)

const homeworkQuestionLimit = 300
