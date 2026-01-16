package ether

import (
	"auction/internal/abi"
	"auction/internal/model"

	"context"
	"log"
	"math/big"
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

// ListeningChainEvent 监听区块链事件并处理重组。
type ListeningChainEvent struct {
	DB *gorm.DB
}

func (l *ListeningChainEvent) Listen() {
	AuctionAddress := "0x2E0D1b09Ae4f5e2dd36b914D56Ac7d04CE73d2B7"
	log.Printf("[INFO] SigCreateAuction: %s\n", SigCreateAuction.Hex())
	log.Printf("[INFO] SigcreateAuctionSimple: %s\n", SigcreateAuctionSimple.Hex())
	log.Printf("[INFO] SigBidUSD: %s\n", SigBidUSD.Hex())
	log.Printf("[INFO] SigAuctionEnd: %s\n", SigAuctionEnd.Hex())

	db := l.DB
	// 连接链
	client, err := ethclient.Dial("wss://sepolia.infura.io/ws/v3/712a252d3b78462587d679771e4e1dc3")
	if err != nil {
		log.Fatalf("[ERROR] 连接以太坊客户端失败: %v\n", err)
	}

	// 获取最后处理的区块号
	lastProcessedBlock := getLastProcessedBlock(db)
	log.Printf("[INFO] 从区块 %d 开始监听\n", lastProcessedBlock)

	// 将 lastProcessedBlock 转换为 *big.Int
	fromBlock := big.NewInt(0).SetUint64(lastProcessedBlock)

	// 订阅日志
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		Addresses: []common.Address{common.HexToAddress(AuctionAddress)},
	}

	logsChan := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logsChan)
	if err != nil {
		log.Fatalf("[ERROR] 监听日志失败: %v\n", err)
	}

	filterer, err := abi.NewAbiFilterer(common.HexToAddress(AuctionAddress), client)
	if err != nil {
		log.Fatalf("[ERROR] 创建过滤器失败: %v\n", err)
	}

	var latestBlockHash common.Hash

	// 订阅新区块头部以检测区块链重组
	headerChan := make(chan *types.Header)
	sub, err = client.SubscribeNewHead(context.Background(), headerChan)
	if err != nil {
		log.Fatalf("[ERROR] 订阅新区块头部失败: %v\n", err)
	}

	for {
		select {
		case err = <-sub.Err():
			log.Printf("[ERROR] 监听订阅事件错误: %v\n", err)
		case vLog := <-logsChan:
			log.Printf("[INFO] 收到日志, topics数量: %d, 第一个topic: %s\n",
				len(vLog.Topics),
				vLog.Topics[0].Hex())

			// 处理事件
			processLogEvent(db, filterer, vLog)

			// 更新最后处理的区块号
			updateLastProcessedBlock(db, vLog.BlockNumber)
		case header := <-headerChan:
			log.Printf("[INFO] 收到新区块头部: %s\n", header.Hash().Hex())

			// 检测区块链重组
			if latestBlockHash != (common.Hash{}) && latestBlockHash != header.ParentHash {
				log.Printf("[WARN] 检测到区块链重组: 最新区块哈希 %s, 父区块哈希 %s\n", latestBlockHash.Hex(), header.ParentHash.Hex())
				rollbackEvents(l.DB, header.Number.Uint64())
			}
			latestBlockHash = header.Hash()
		}
	}
}

// rollbackEvents 在区块链重组的情况下处理事件回滚。
func rollbackEvents(db *gorm.DB, rollbackToBlock uint64) {
	log.Printf("[INFO] 开始回滚到区块: %d\n", rollbackToBlock)

	// 删除或标记从回滚区块开始的日志表事件
	if err := db.Where("block_number >= ?", rollbackToBlock).Delete(&model.ProcessedLog{}).Error; err != nil {
		log.Printf("[ERROR] 回滚日志表事件失败: %v\n", err)
		return
	}

	// 删除或标记从回滚区块开始的业务表事件
	if err := db.Where("block_number >= ?", rollbackToBlock).Delete(&model.Auction{}).Error; err != nil {
		log.Printf("[ERROR] 回滚拍卖表事件失败: %v\n", err)
		return
	}

	if err := db.Where("block_number >= ?", rollbackToBlock).Delete(&model.Bid{}).Error; err != nil {
		log.Printf("[ERROR] 回滚出价表事件失败: %v\n", err)
		return
	}

	log.Printf("[INFO] 回滚完成，已删除区块号 >= %d 的日志和业务表事件\n", rollbackToBlock)
}

func processLogEvent(db *gorm.DB, filterer *abi.AbiFilterer, vLog types.Log) {
	switch vLog.Topics[0] {
	case SigCreateAuction:
		log.Printf("[INFO] 处理事件: CreateAuction\n")
		processEventCreateAuction(db, filterer, vLog)
	case SigcreateAuctionSimple:
		log.Printf("[INFO] 处理事件: CreateAuctionSimple\n")
		processEventCreateAuction(db, filterer, vLog)
	case SigBidUSD:
		log.Printf("[INFO] 处理事件: BidUSD\n")
		processEventBidUSD(db, filterer, vLog)
	case SigAuctionEnd:
		log.Printf("[INFO] 处理事件: AuctionEnd\n")
		processEventAuctionEnd(db, filterer, vLog)
	default:
		// 打印所有 topics 以便调试
		for i, topic := range vLog.Topics {
			log.Printf("[DEBUG] Topic[%d]: %s\n", i, topic.Hex())
		}
		log.Printf("[WARN] 接收到未定义事件: %s\n", vLog.Topics[0].Hex())
	}
}

func getLastProcessedBlock(db *gorm.DB) uint64 {
	var lastBlock model.ProcessedLog
	// db.Last 很牛逼的用法
	if err := db.Last(&lastBlock).Error; err != nil {
		log.Printf("[WARN] 无法获取最后处理的区块号，默认从区块 0 开始: %v\n", err)
		return 0
	}
	log.Printf("[INFO] 最后处理的区块号: %d\n", lastBlock.BlockNumber)
	return lastBlock.BlockNumber
}

func updateLastProcessedBlock(db *gorm.DB, blockNumber uint64) {
	log.Printf("[INFO] 更新最后处理的区块号: %d\n", blockNumber)
	if err := db.Create(&model.ProcessedLog{BlockNumber: blockNumber}).Error; err != nil {
		log.Printf("[ERROR] 更新最后处理的区块号失败: %v\n", err)
	}
}

// 处理创建合约事件
func processEventCreateAuction(db *gorm.DB, filterer *abi.AbiFilterer, vLog types.Log) {
	log.Println("监听到CreateAuction!")
	event, err := filterer.ParseAuctionCreated(vLog)
	if err != nil {
		log.Fatalln(err)
	}

	// 保存入库
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

	// 更新数据库
	contractAddress := vLog.Address.Hex()
	auctionId := event.AuctionId
	winner := event.Winner

	db.Model(&model.Auction{}).Where("contract_address = ? and auction_id=?",
		contractAddress, auctionId).Updates(map[string]interface{}{
		"winner": winner.Hex(),
		"status": "end"})

	log.Println("AuctionEnd处理完毕!")
}
