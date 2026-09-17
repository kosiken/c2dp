package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// C2PAManifest holds the raw JSON manifest report c2patool produced for a
// post's signed image, stored as-is in a jsonb column.
type C2PAManifest json.RawMessage

func (m C2PAManifest) MarshalJSON() ([]byte, error) {
	if len(m) == 0 {
		return []byte("null"), nil
	}
	return m, nil
}

func (m *C2PAManifest) UnmarshalJSON(data []byte) error {
	*m = append((*m)[0:0], data...)
	return nil
}

func (m C2PAManifest) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	return []byte(m), nil
}

func (m *C2PAManifest) Scan(value any) error {
	if value == nil {
		*m = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*m = append((*m)[0:0], v...)
		return nil
	case string:
		*m = C2PAManifest(v)
		return nil
	default:
		return fmt.Errorf("unsupported type for C2PAManifest: %T", value)
	}
}

type Post struct {
	ID           uint         `json:"id" gorm:"primaryKey"`
	UserID       uint         `json:"user_id" gorm:"not null;index"`
	User         User         `json:"user" gorm:"constraint:OnDelete:CASCADE"`
	Caption      string       `json:"caption" gorm:"size:2200"`
	ImagePath    string       `json:"-" gorm:"not null"`
	ImageURL     string       `json:"image_url" gorm:"-"`
	ManifestData C2PAManifest `json:"manifest_data,omitempty" gorm:"type:jsonb"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}
