package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/service"

	"github.com/gin-gonic/gin"
)

func GetCCSwitchMirrorManifest(c *gin.Context) {
	common.ApiSuccess(c, service.GetCCSwitchMirrorManifestView())
}

func CreateCCSwitchMirrorDownloadTicket(c *gin.Context) {
	ticket, err := service.CreateCCSwitchMirrorDownloadTicket(c.Param("platform"), c.Param("asset_id"), c.GetInt("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	common.ApiSuccess(c, ticket)
}

func DownloadCCSwitchMirrorAsset(c *gin.Context) {
	download, err := service.ResolveCCSwitchMirrorDownload(c.Param("platform"), c.Param("asset_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.Header("Cache-Control", "private, max-age=3600")
	c.FileAttachment(download.Path, download.FileName)
}

func DownloadCCSwitchMirrorAssetWithTicket(c *gin.Context) {
	if err := service.ValidateCCSwitchMirrorDownloadTicket(
		c.Param("platform"),
		c.Param("asset_id"),
		c.Query("user_id"),
		c.Query("expires_at"),
		c.Query("signature"),
	); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	DownloadCCSwitchMirrorAsset(c)
}
