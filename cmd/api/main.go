package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konflic/ip_checking_bot/helpers"
	_ "github.com/lib/pq"
)

func main() {
	router := gin.Default()
	router.GET("/get_users", get_users)
	router.GET("/get_user/:username", get_user_by_id)
	router.DELETE("/remove/:username/:request")
	router.Run("localhost:8080")
}

func get_users(c *gin.Context) {
	usernames := helpers.GetDistinctUsernames(helpers.InitDb())
	c.IndentedJSON(http.StatusOK, usernames)
}

func get_user_by_id(c *gin.Context) {
	username := c.Param("username")
	requests := helpers.GetAllUserRequests(username, helpers.InitDb())
	c.IndentedJSON(http.StatusOK, requests)
}

func delete_user_request(c *gin.Context) {
	username := c.Param("username")
	request := c.Param("request")
	helpers.DeleteUserIPRequest(username, request, helpers.InitDb())
	c.IndentedJSON(http.StatusOK, "Removed")
}
