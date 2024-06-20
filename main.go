package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ShipOwner struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Name          string `json:"name"`
	Address       string `json:"address"`
	PhoneNumber   string `json:"phone_number"`
	NumberOfShips int    `json:"number_of_ships"`
}

var db *gorm.DB
var err error

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/shipowners_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database: ", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&ShipOwner{}); err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	router.GET("/shipowners", GetShipOwners)
	router.POST("/shipowners", CreateShipOwner)
	router.PUT("/shipowners/:id", UpdateShipOwner)
	router.DELETE("/shipowners/:id", DeleteShipOwner)

	router.Run(":8080")
}

func GetShipOwners(c *gin.Context) {
	var shipowners []ShipOwner
	if err := db.Find(&shipowners).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipowners)
}

func CreateShipOwner(c *gin.Context) {
	var shipowner ShipOwner
	if err := c.ShouldBindJSON(&shipowner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Create(&shipowner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipowner)
}

func UpdateShipOwner(c *gin.Context) {
	id := c.Param("id")
	var shipowner ShipOwner
	if err := db.First(&shipowner, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ShipOwner not found"})
		return
	}
	if err := c.ShouldBindJSON(&shipowner); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := db.Save(&shipowner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shipowner)
}

func DeleteShipOwner(c *gin.Context) {
	id := c.Param("id")
	var shipowner ShipOwner
	if err := db.First(&shipowner, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ShipOwner not found"})
		return
	}
	if err := db.Delete(&shipowner).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ShipOwner deleted"})
}
