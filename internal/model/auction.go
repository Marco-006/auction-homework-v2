package model

import "time"

// 首字母大写
type Auction struct {
	// 拍卖合约本身信息
	ID uint64 `gorm:"primaryKey;autoIncrement"`
	// ChainID         uint64 `gorm:"not null;index:uniq_chain_contract_auction,unique,priority:1"`
	ContractAddress string `gorm:"size:42;not null;index:uniq_contract_auction,unique,priority:1"`
	AuctionID       uint64 `gorm:"not null;index:uniq_contract_auction,unique,priority:2"`

	// NFT
	NFTContract string `gorm:"size:42;not null;index:idx_nft,priority:1"`
	TokenId     string `gorm:"size:80;not null;index:idx_nft,priority:2"`
	Seller      string `gorm:"size:42;not null;index:idx_seller"`

	StartPrice       uint64 `gorm:"not null;default:0"`
	HighestBidder    string `gorm:"size:42"`
	HighestBidAmount string `gorm:"size:100;default:0"`
	BidCount         uint64 `gorm:"not null;default:0"`
	Winner           string `gorm:"size:42;"`

	// CreatedTxHash string `gorm:"size:66;not null"`
	// EndedTxHash   string `gorm:"size:66;not null"`
	// CreatedBlock  uint64 `gorm:"not null"`
	// UpdatedBlock  uint64 `gorm:"not null"`

	// active cancle end
	Status string `gorm:"size:20;not null;index:idx_status"`

	// 时间
	StartTime       time.Time `gorm:"not null"`
	DurationMinutes uint64    `json:"durationMinutes" gorm:"not null"`
	EndTime         time.Time `gorm:"not null"`

	CreatedAt		time.Time  `gorm:"autoCreateTime"`
	UpdateAt		time.Time  `gorm:"autoUpdateTime"`
}
