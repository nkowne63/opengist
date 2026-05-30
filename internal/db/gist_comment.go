package db

import (
	"time"
)

type GistComment struct {
	ID uint `gorm:"primaryKey"`

	GistID uint  `gorm:"index;not null"`
	Gist   *Gist `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	UserID uint  `gorm:"index;not null"`
	User   *User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Content   string `gorm:"type:text;not null"`
	CreatedAt int64
	UpdatedAt int64
}

func (c *GistComment) Create() error {
	now := time.Now().Unix()
	if c.CreatedAt == 0 {
		c.CreatedAt = now
	}
	if c.UpdatedAt == 0 {
		c.UpdatedAt = c.CreatedAt
	}
	return db.Create(c).Error
}

func GetGistComments(gistID uint) ([]*GistComment, error) {
	var comments []*GistComment
	err := db.
		Preload("User").
		Where("gist_id = ?", gistID).
		Order("created_at asc, id asc").
		Find(&comments).Error
	return comments, err
}

func CountGistComments(gistID uint) (int64, error) {
	var count int64
	err := db.
		Model(&GistComment{}).
		Where("gist_id = ?", gistID).
		Count(&count).Error
	return count, err
}
