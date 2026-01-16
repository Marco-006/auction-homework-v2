package main

import (
	"auction/internal/controller"
	"auction/internal/database"
	"auction/internal/ether"
	"auction/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	db := database.InitDB()
	listener := &ether.ListeningChainEvent{DB: db}
	go listener.Listen()

	// 初始化controller
	auctionController := controller.NewAuctionController(db)

	// 设置路由
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			utils.Success(c, "api is running")
		})
		v1.GET("/auctions", auctionController.QueryAuction)
		v1.GET("/:auctionId/bids", auctionController.QueryAuctionBids)
	}

	//启动服务
	err := router.Run(":8080")
	if err != nil {
		log.Fatalln("服务启动失败:", err)
	}
}
