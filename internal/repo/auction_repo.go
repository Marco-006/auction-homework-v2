package repo

import (
	"auction/internal/model"

	"gorm.io/gorm"
)

type AuctionRepo struct {
	db gorm.DB
}

func (auctionRepo *AuctionRepo) UpdateAuction(auctionId uint64,
	winner string,
	amount string,
	currency string) {

	auctionRepo.db.Model(&model.Auction{}).Where("")

}

// event AuctionEnded(
// 	uint256 indexed auctionId,
// 	address indexed winner,
// 	uint256 amount,
// 	address currency);

// db.Model(&User{}).           // 告诉 GORM 操作哪张表
//    Where("id = ?", 123).     // 条件
//    Select("name", "age").    // 只改这两列
//    Updates(map[string]interface{}{
//        "name": "Tom",
//        "age":  18,
//    })

// auctionId := event.AuctionId
// winner := event.Winner
// amount := event.Amount
// currency := event.Currency
