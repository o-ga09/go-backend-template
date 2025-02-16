package service

import "github.com/gin-gonic/gin"

type UserHandler struct{}

func (s *UserHandler) Find(c *gin.Context) {
	c.JSON(200, gin.H{
		"id":   1,
		"name": "test",
		"age":  20,
	})
}
