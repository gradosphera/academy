package upload

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestMaterialFilePath_String(t *testing.T) {
	miniAppID := uuid.New()
	productID := uuid.New()
	lessonID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		path MaterialFilePath
		want string
	}{
		{
			name: "Empty path",
			path: MaterialFilePath{},
			want: "",
		},
		{
			name: "Only MiniAppID",
			path: MaterialFilePath{MiniAppID: miniAppID},
			want: filepath.Join(miniAppMaterialDir, miniAppID.String()),
		},
		{
			name: "MiniAppID and ProductID",
			path: MaterialFilePath{MiniAppID: miniAppID, ProductID: productID},
			want: filepath.Join(
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
			),
		},
		{
			name: "MiniAppID, ProductID, and LessonID",
			path: MaterialFilePath{MiniAppID: miniAppID, ProductID: productID, LessonID: lessonID},
			want: filepath.Join(
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				lessonMaterialDir, lessonID.String(),
			),
		},
		{
			name: "All IDs",
			path: MaterialFilePath{MiniAppID: miniAppID, ProductID: productID, LessonID: lessonID, UserID: userID},
			want: filepath.Join(
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				lessonMaterialDir, lessonID.String(),
				userMaterialDir, userID.String(),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.path.String()

			if got != tt.want {
				t.Errorf("%q != %q", got, tt.want)
			}
		})
	}
}

func TestParseMaterialFilePath(t *testing.T) {
	fileID := uuid.New()
	miniAppID := uuid.New()
	productID := uuid.New()
	lessonID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name     string
		filePath string
		want     *MaterialFilePath
		isErr    bool
	}{
		{
			name:     "Empty path",
			filePath: fileID.String(),
			want:     &MaterialFilePath{},
		},
		{
			name: "Only MiniAppID",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				fileID.String(),
			}, "/"),
			want: &MaterialFilePath{MiniAppID: miniAppID},
		},
		{
			name: "MiniAppID and ProductID",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				fileID.String(),
			}, "/"),
			want: &MaterialFilePath{MiniAppID: miniAppID, ProductID: productID},
		},
		{
			name: "MiniAppID, ProductID, and LessonID",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				lessonMaterialDir, lessonID.String(),
				fileID.String(),
			}, "/"),
			want: &MaterialFilePath{MiniAppID: miniAppID, ProductID: productID, LessonID: lessonID},
		},
		{
			name: "All IDs",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				lessonMaterialDir, lessonID.String(),
				userMaterialDir, userID.String(),
				fileID.String(),
			}, "/"),
			want: &MaterialFilePath{
				MiniAppID: miniAppID,
				ProductID: productID,
				LessonID:  lessonID,
				UserID:    userID,
			},
		},
		{
			name:     "Invalid MiniAppID",
			filePath: "invalid-uuid",
			want:     nil,
			isErr:    true,
		},
		{
			name: "Invalid ProductID",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				"invalid-uuid", fileID.String(),
			}, "/"),
			want:  nil,
			isErr: true,
		},
		{
			name: "Invalid LessonID",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				"invalid-uuid", fileID.String(),
			}, "/"),
			want:  nil,
			isErr: true,
		},
		{
			name: "Invalid UserID",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				lessonMaterialDir, lessonID.String(),
				"invalid-uuid", fileID.String(),
			}, "/"),
			want:  nil,
			isErr: true,
		},
		{
			name:     "Too few components (should return empty)",
			filePath: "/",
			want:     nil,
			isErr:    true,
		},
		{
			name: "Too many components (should return empty)",
			filePath: strings.Join([]string{
				miniAppMaterialDir, miniAppID.String(),
				productMaterialDir, productID.String(),
				lessonMaterialDir, lessonID.String(),
				userMaterialDir, userID.String(),
				fileID.String(),
				uuid.New().String(),
			}, "/"),
			want:  nil,
			isErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMaterialFilePath(tt.filePath)
			if err != nil && !tt.isErr {
				t.Errorf("err: %v", err)
				return
			}

			if got.String() != tt.want.String() {
				t.Errorf("%q != %q", got, tt.want)
			}
			// t.Errorf("%q -> %q", got, tt.want)
		})
	}
}
