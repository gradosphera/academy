package upload

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type MaterialFilePath struct {
	MiniAppID uuid.UUID
	ProductID uuid.UUID
	LessonID  uuid.UUID
	UserID    uuid.UUID

	ProductLevelID uuid.UUID
}

const (
	miniAppMaterialDir      = "ma"
	productMaterialDir      = "p"
	lessonMaterialDir       = "l"
	userMaterialDir         = "u"
	productLevelMaterialDir = "pl"
)

func (p *MaterialFilePath) String() string {
	if p == nil || p.MiniAppID == uuid.Nil {
		return ""
	}
	if p.LessonID != uuid.Nil && p.ProductLevelID != uuid.Nil {
		panic("unexpected material file path parameters")
	}

	if p.ProductID != uuid.Nil &&
		p.LessonID != uuid.Nil &&
		p.UserID != uuid.Nil {

		return filepath.Join(
			miniAppMaterialDir,
			p.MiniAppID.String(),
			productMaterialDir,
			p.ProductID.String(),
			lessonMaterialDir,
			p.LessonID.String(),
			userMaterialDir,
			p.UserID.String(),
		)
	}

	if p.ProductID != uuid.Nil && p.ProductLevelID != uuid.Nil {
		return filepath.Join(
			miniAppMaterialDir,
			p.MiniAppID.String(),
			productMaterialDir,
			p.ProductID.String(),
			productLevelMaterialDir,
			p.ProductLevelID.String(),
		)
	}

	if p.ProductID != uuid.Nil && p.LessonID != uuid.Nil {
		return filepath.Join(
			miniAppMaterialDir,
			p.MiniAppID.String(),
			productMaterialDir,
			p.ProductID.String(),
			lessonMaterialDir,
			p.LessonID.String(),
		)
	}

	if p.ProductID != uuid.Nil {
		return filepath.Join(
			miniAppMaterialDir,
			p.MiniAppID.String(),
			productMaterialDir,
			p.ProductID.String(),
		)
	}

	return filepath.Join(miniAppMaterialDir, p.MiniAppID.String())
}

func ParseMaterialFilePath(filePath string) (*MaterialFilePath, error) {
	ids := strings.Split(filePath, "/")

	// In some cases filename base can be non-uuid. Forr example generated excel analytics.
	//
	// filename := ids[len(ids)-1]
	// _, err := uuid.Parse(strings.TrimSuffix(filename, filepath.Ext(filename)))
	// if err != nil {
	// 	return nil, fmt.Errorf("filename is not uuid: %w", err)
	// }

	switch len(ids) {
	case 1:
		return &MaterialFilePath{}, nil
	case 3:
		if ids[0] != miniAppMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[0])
		}
		miniAppID, err := uuid.Parse(ids[1])
		if err != nil {
			return nil, fmt.Errorf("miniAppID is not uuid: %w", err)
		}
		return &MaterialFilePath{MiniAppID: miniAppID}, nil
	case 5:
		if ids[0] != miniAppMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[0])
		}
		if ids[2] != productMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[2])
		}
		miniAppID, err := uuid.Parse(ids[1])
		if err != nil {
			return nil, fmt.Errorf("miniAppID is not uuid: %w", err)
		}
		productID, err := uuid.Parse(ids[3])
		if err != nil {
			return nil, fmt.Errorf("productID is not uuid: %w", err)
		}
		return &MaterialFilePath{MiniAppID: miniAppID, ProductID: productID}, nil
	case 7:
		if ids[0] != miniAppMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[0])
		}
		if ids[2] != productMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[2])
		}
		miniAppID, err := uuid.Parse(ids[1])
		if err != nil {
			return nil, fmt.Errorf("miniAppID is not uuid: %w", err)
		}
		productID, err := uuid.Parse(ids[3])
		if err != nil {
			return nil, fmt.Errorf("productID is not uuid: %w", err)
		}

		switch ids[4] {
		case lessonMaterialDir:
			if ids[4] != lessonMaterialDir {
				return nil, fmt.Errorf("unexpected directory: %v", ids[2])
			}
			lessonID, err := uuid.Parse(ids[5])
			if err != nil {
				return nil, fmt.Errorf("lessonID is not uuid: %w", err)
			}
			return &MaterialFilePath{
				MiniAppID: miniAppID,
				ProductID: productID,
				LessonID:  lessonID,
			}, nil
		case productLevelMaterialDir:
			if ids[4] != productLevelMaterialDir {
				return nil, fmt.Errorf("unexpected directory: %v", ids[2])
			}
			productLevelID, err := uuid.Parse(ids[5])
			if err != nil {
				return nil, fmt.Errorf("productLevelID is not uuid: %w", err)
			}
			return &MaterialFilePath{
				MiniAppID:      miniAppID,
				ProductID:      productID,
				ProductLevelID: productLevelID,
			}, nil
		default:
			return nil, fmt.Errorf("unexpected directory: %v", ids[4])
		}
	case 9:
		if ids[0] != miniAppMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[0])
		}
		if ids[2] != productMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[2])
		}
		if ids[4] != lessonMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[2])
		}
		if ids[6] != userMaterialDir {
			return nil, fmt.Errorf("unexpected directory: %v", ids[2])
		}
		miniAppID, err := uuid.Parse(ids[1])
		if err != nil {
			return nil, fmt.Errorf("miniAppID is not uuid: %w", err)
		}
		productID, err := uuid.Parse(ids[3])
		if err != nil {
			return nil, fmt.Errorf("productID is not uuid: %w", err)
		}
		lessonID, err := uuid.Parse(ids[5])
		if err != nil {
			return nil, fmt.Errorf("lessonID is not uuid: %w", err)
		}
		userID, err := uuid.Parse(ids[7])
		if err != nil {
			return nil, fmt.Errorf("userID is not uuid: %w", err)
		}
		return &MaterialFilePath{
			MiniAppID: miniAppID,
			ProductID: productID,
			LessonID:  lessonID,
			UserID:    userID,
		}, nil
	default:
		return nil, fmt.Errorf("unexpected path")
	}
}
