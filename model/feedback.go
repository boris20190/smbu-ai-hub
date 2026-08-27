package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
)

type Feedback struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id" gorm:"index"`
	Username    string `json:"username" gorm:"index;size:64"`
	Title       string `json:"title" gorm:"size:120"`
	Content     string `json:"content" gorm:"type:text"`
	Contact     string `json:"contact" gorm:"size:160"`
	CreatedTime int64  `json:"created_time" gorm:"bigint;index"`
}

func CreateFeedback(userId int, username string, title string, content string, contact string) (*Feedback, error) {
	feedback := &Feedback{
		UserId:      userId,
		Username:    strings.TrimSpace(username),
		Title:       strings.TrimSpace(title),
		Content:     strings.TrimSpace(content),
		Contact:     strings.TrimSpace(contact),
		CreatedTime: common.GetTimestamp(),
	}

	if err := DB.Create(feedback).Error; err != nil {
		return nil, err
	}

	return feedback, nil
}

func GetAllFeedbacks(startIdx int, num int) (feedbacks []*Feedback, total int64, err error) {
	query := DB.Model(&Feedback{})
	if err = query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&feedbacks).Error
	return feedbacks, total, err
}
