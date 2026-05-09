package controllers

import (
	"net/http"

	"backend/db"
	"backend/models"

	"github.com/gin-gonic/gin"
)

type ConfigInput struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

func GetConfig(c *gin.Context) {
	var configs []models.AppConfig
	db.DB.Find(&configs)
	
	configMap := make(map[string]string)
	for _, conf := range configs {
		configMap[conf.Key] = conf.Value
	}

	c.JSON(http.StatusOK, configMap)
}

func UpdateConfig(c *gin.Context) {
	var input ConfigInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var config models.AppConfig
	// Create or update
	if err := db.DB.Where("key = ?", input.Key).First(&config).Error; err != nil {
		config = models.AppConfig{Key: input.Key, Value: input.Value}
		db.DB.Create(&config)
	} else {
		config.Value = input.Value
		db.DB.Save(&config)
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
