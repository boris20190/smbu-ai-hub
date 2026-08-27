package controller

import (
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

type createFeedbackRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Contact string `json:"contact"`
}

func CreateFeedback(c *gin.Context) {
	var req createFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	contact := strings.TrimSpace(req.Contact)

	if utf8.RuneCountInString(title) < 2 || utf8.RuneCountInString(title) > 80 {
		common.ApiErrorMsg(c, "反馈标题长度需为 2-80 个字符")
		return
	}
	if utf8.RuneCountInString(content) < 5 || utf8.RuneCountInString(content) > 2000 {
		common.ApiErrorMsg(c, "反馈内容长度需为 5-2000 个字符")
		return
	}
	if utf8.RuneCountInString(contact) > 120 {
		common.ApiErrorMsg(c, "联系方式不能超过 120 个字符")
		return
	}

	feedback, err := model.CreateFeedback(
		c.GetInt("id"),
		c.GetString("username"),
		title,
		content,
		contact,
	)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, feedback)
}

func GetAllFeedbacks(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	feedbacks, total, err := model.GetAllFeedbacks(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(feedbacks)
	common.ApiSuccess(c, pageInfo)
}
