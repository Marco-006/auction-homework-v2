package model

import "time"

type Bid struct {
	ID uint64 `gorm:"primaryKey;autoIncrement"`
	// ContractAddress string `gorm:"size:42;not null"`
	AuctionID uint64 `gorm:"not null;index:idx_auction_id"`

	// 每次出价支持不同币种；但是拍卖的记录里面，仅支持美金
	Bidder      string `gorm:"size:42;not null;index:idx_bidder"`
	BidAmount   string `gorm:"size:100;not null"`
	BidCurrency string `gorm:"size:100"`
	BidUSD      string `gorm:"size:100"`

	// TxHash      string    `gorm:"size:66;not null"`
	BlockNumber uint64 `gorm:"not null"` // 区块号
	// BidTime     time.Time `gorm:"not null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}
