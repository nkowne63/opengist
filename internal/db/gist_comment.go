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

func (c *GistComment) Update() error {
	c.UpdatedAt = time.Now().Unix()
	return db.Model(&GistComment{}).
		Where("id = ?", c.ID).
		Updates(map[string]interface{}{
			"content":    c.Content,
			"updated_at": c.UpdatedAt,
		}).Error
}

func (c *GistComment) Delete() error {
	return db.Delete(&GistComment{}, c.ID).Error
}

func GetGistCommentByID(id uint) (*GistComment, error) {
	var comment GistComment
	err := db.
		Preload("User").
		Preload("Gist").
		First(&comment, id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func GetGistComments(gistID uint) ([]*GistComment, error) {
	var comments []*GistComment
	err := db.
		Preload("User").
		Preload("Gist").
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
