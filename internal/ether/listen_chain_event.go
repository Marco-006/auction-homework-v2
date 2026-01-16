package ether

import (
	"auction/internal/abi"
	"auction/internal/model"

	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

// 要监听合约的事件 topic hash 签名哈希
var (
	SigCreateAuction       = crypto.Keccak256Hash([]byte("AuctionCreated(address,uint256,uint256,uint256,uint256,address)"))
	SigcreateAuctionSimple = crypto.Keccak256Hash([]byte("AuctionCreated(uint256,address,address,uint256,uint256,uint256,uint256,address)"))
	SigBidUSD              = crypto.Keccak256Hash([]byte("BidPlaced(uint256,address,uint256,address)"))
	SigAuctionEnd          = crypto.Keccak256Hash([]byte("AuctionEnded(uint256,address,uint256)"))
)

type ListeningChainEvent struct {
	DB *gorm.DB
}

func (l *ListeningChainEvent) Listen() {

	AuctionAddress := "8888"
	log.Printf("SigCreateAuction: %s\n", SigCreateAuction.Hex())
	log.Printf("SigcreateAuctionSimple: %s\n", SigcreateAuctionSimple.Hex())
	log.Printf("SigBidUSD: %s\n", SigBidUSD.Hex())
	log.Printf("SigAuctionEnd: %s\n", SigAuctionEnd.Hex())

	db := l.DB
	// 连接  链
	client, err := ethclient.Dial("wss://sepolia.infura.io/ws/v3/712a252d3b78462587d679771e4e1dc3")
	if err != nil {
		log.Fatalln(err)
	}

	// 订阅日志
	query := ethereum.FilterQuery{
		//BlockHash: &blockHash,
		Addresses: []common.Address{common.HexToAddress(AuctionAddress)},
	}

	logsChan := make(chan types.Log)

	sub, err := client.SubscribeFilterLogs(context.Background(), query, logsChan)
	if err != nil {
		log.Fatalln("监听logs失败:", err)
	}

	filterer, err := abi.NewAbiFilterer(common.HexToAddress(AuctionAddress), client)
	if err != nil {
		log.Fatalln(err)
	}

	for {
		select {
		case err = <-sub.Err():
			log.Println("监听订阅事件err:", err)
		case vLog := <-logsChan:
			log.Printf("收到日志, topics数量: %d, 第一个topic: %s\n",
				len(vLog.Topics),
				vLog.Topics[0].Hex())

			switch vLog.Topics[0] {
			case SigCreateAuction:
				processEventCreateAuction(db, filterer, vLog)
			case SigcreateAuctionSimple:
				processEventCreateAuction(db, filterer, vLog)
			case SigBidUSD:
				processEventBidUSD(db, filterer, vLog)
			case SigAuctionEnd:
				processEventAuctionEnd(db, filterer, vLog)
			default:
				// 打印所有 topics 以便调试
				for i, topic := range vLog.Topics {
					log.Printf("Topic[%d]: %s\n", i, topic.Hex())
				}
				log.Println("接收到未定义事件:", vLog.Topics[0].Hex())
			}
		}
	}

}

// 处理创建合约事件
func processEventCreateAuction(db *gorm.DB, filterer *abi.AbiFilterer, vLog types.Log) {
	log.Println("监听到CreateAuction!")
	event, err := filterer.ParseAuctionCreated(vLog)
	if err != nil {
		log.Fatalln(err)
	}

	//保存入库
	auction := model.Auction{
		ContractAddress: vLog.Address.Hex(),
		AuctionID:       event.AuctionId.Uint64(),

		NFTContract: event.NftContract.Hex(),
		TokenId:     event.TokenId.String(),
		Seller:      event.Seller.Hex(),

		StartPrice: event.StartPrice.Uint64(),
		Status:     "active",
		StartTime:  time.Unix(event.StartTime.Int64(), 0),

		DurationMinutes: event.Duration.Uint64(),
		EndTime:         time.Unix(int64(vLog.BlockTimestamp), 0).Add(time.Duration(event.Duration.Uint64()) * time.Minute),
	}
	err = db.Create(&auction).Error
	if err != nil {
		log.Fatalln("CreateAuction处理失败:", err)
	}

	log.Println("CreateAuction处理完毕!")
}

// 处理出价合约事件
func processEventBidUSD(db *gorm.DB, filterer *abi.AbiFilterer, vLog types.Log) {
	log.Println("监听到BidUSD!")
	event, err := filterer.ParseBidPlaced(vLog)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("processEventBidUSD contract address..", vLog.Address.Hex())

	bid := model.Bid{
		AuctionID:   event.AuctionId.Uint64(),
		Bidder:      event.Bidder.Hex(),
		BidAmount:   event.Amount.String(),
		BidCurrency: event.Currency.String(),
	}
	err = db.Create(&bid).Error
	if err != nil {
		log.Fatalln("BidUSD处理失败:", err)
	}

	// 更新出价最高的人
	log.Println("BidUSD处理完毕!")
}

// 处理敲定合约事件
func processEventAuctionEnd(db *gorm.DB, filterer *abi.AbiFilterer, vLog types.Log) {
	log.Println("监听到AuctionEnd!")
	event, err := filterer.ParseAuctionEnded(vLog)
	if err != nil {
		log.Fatalln(err)
	}

	// 更新数据库 --update DB
	contractAddress := vLog.Address.Hex()
	auctionId := event.AuctionId
	winner := event.Winner
	// amount := event.Amount
	// currency := event.Currency

	db.Model(&model.Auction{}).Where("contract_address = ? and auction_id=?",
		contractAddress, auctionId).Updates(map[string]interface{}{
		"winner": winner.Hex(),
		"status": "end"})

	// repo.UpdateAuction()
	// repo.UpdateAuction(auctionId, winner, amount, currency)

	log.Println("AuctionEnd处理完毕!")
}
