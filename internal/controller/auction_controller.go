package controller

import (
	"auction/internal/dto"
	"auction/internal/model"
	"auction/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuctionController struct {
	Db gorm.DB
}

func NewAuctionController(db *gorm.DB) *AuctionController {
	return &AuctionController{Db: *db}
}

func (ctl *AuctionController) QueryAuction(c *gin.Context) {
	var req dto.QueryAuctionRequest

	err := c.ShouldBind(&req)
	if err != nil {
		log.Fatalln("参数绑定有误")
	}

	var auctions []model.Auction
	tx := ctl.Db.Model(&model.Auction{})

	if req.SellerAddress != "" {
		tx = tx.Where("seller = ?", req.SellerAddress)
	}

	if req.NftAddress != "" {
		tx = tx.Where("nft_contract = ?", req.NftAddress)
	}

	if req.TokenId != "" {
		tx = tx.Where("token_id = ?", req.TokenId)
	}

	// layout 必须是调用方已经定义好的时间格式常量，例如 "2006-01-02 15:04:05"
	layout := "2006-01-02 15:04:05"
	if req.BeginTime != "" && req.EndTime != "" {
		beginTime, err := time.ParseInLocation(layout, req.BeginTime, time.Local)
		if err != nil {
			log.Fatalln("开始时间格式有误")
		}
		endTime, err := time.ParseInLocation(layout, req.EndTime, time.Local)
		if err != nil {
			log.Fatalln("结束时间格式有误")
		}
		tx = tx.Where("end_time between ? and ?", beginTime, endTime)
	}

	// 排序：字段 + 方向
	orderStr := "created_at" // 默认字段
	// if req.SortField == "UpdatedAt" {
	// 	orderStr = "updated_at"
	// }
	// DESC 降序
	if req.SortOrder == "desc" || req.SortOrder == "DESC" {
		orderStr += " DESC"
	} else {
		orderStr += " ASC"
	}
	tx = tx.Order(orderStr)

	err = tx.Find(&auctions).Error
	if err != nil {
		log.Fatalln("query auctions fail")
	}
	utils.Success(c, auctions)
}

func (ctl *AuctionController) QueryAuctionBids(c *gin.Context) {
	auctionId := c.Param("auctionId")
	var bids []model.Bid
	if err := ctl.Db.Where("auction_id=?", auctionId).Find(&bids).Error; err != nil {
		c.JSON(404, gin.H{"error": "Post not found"})
		return
	}
	utils.Success(c, bids)
}

func (ctl *AuctionController) QueryNftsByWalletAdress(c *gin.Context) {
	auctionId := c.Param("id")
	var bids []model.Bid
	if err := ctl.Db.Where("auction_id=?", auctionId).Find(&bids).Error; err != nil {
		c.JSON(404, gin.H{"error": "Post not found"})
		return
	}
	utils.Success(c, bids)
}
