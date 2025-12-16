package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	// start a goroutine that will close this auction when the auction interval elapses
	go ar.startAuctionCloser(auctionEntity.Id, auctionEntity.Timestamp)

	return nil
}

// startAuctionCloser schedules a goroutine that will update auction status to Completed
// after the configured auction interval. It is safe to run multiple goroutines.
func (ar *AuctionRepository) startAuctionCloser(id string, start time.Time) {
	interval := getAuctionInterval()
	endTime := start.Add(interval)

	go func() {
		// if endTime is in the past, close immediately
		wait := time.Until(endTime)
		if wait > 0 {
			timer := time.NewTimer(wait)
			<-timer.C
			timer.Stop()
		}

		// attempt to update auction status to Completed only if it's still Active
		filter := bson.M{"_id": id, "status": auction_entity.Active}
		update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

		if _, err := ar.Collection.UpdateOne(context.Background(), filter, update); err != nil {
			logger.Error("Error trying to close expired auction", err)
			return
		}
	}()
}

// getAuctionInterval reads AUCTION_INTERVAL env var and returns a duration.
// On parse error or empty value, defaults to 5 minutes to match bid behavior.
func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5
	}

	return duration
}
