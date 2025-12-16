package auction

import (
    "context"
    "fullcycle-auction_go/configuration/database/mongodb"
    "fullcycle-auction_go/internal/entity/auction_entity"
    "os"
    "testing"
    "time"
)

// This is an integration test that requires a running MongoDB instance.
// Set MONGODB_URL and MONGODB_DB environment variables before running.
func TestAuctionAutoClose(t *testing.T) {
    // require mongodb
    if os.Getenv("MONGODB_URL") == "" || os.Getenv("MONGODB_DB") == "" {
        t.Skip("Skipping integration test; MONGODB_URL or MONGODB_DB is not set")
    }

    // set short auction interval for test
    os.Setenv("AUCTION_INTERVAL", "1s")

    ctx := context.Background()
    db, err := mongodb.NewMongoDBConnection(ctx)
    if err != nil {
        t.Fatalf("could not connect to mongodb: %v", err)
    }

    auctionRepo := NewAuctionRepository(db)

    auctionEntity, err := auction_entity.CreateAuction("product-test", "category", "a longer description for testing", auction_entity.New)
    if err != nil {
        t.Fatalf("error creating auction entity: %v", err)
    }

    if err := auctionRepo.CreateAuction(ctx, auctionEntity); err != nil {
        t.Fatalf("error inserting auction: %v", err)
    }

    // wait for the auction to be closed (interval is 1s)
    time.Sleep(2 * time.Second)

    auctionClosed, err := auctionRepo.FindAuctionById(ctx, auctionEntity.Id)
    if err != nil {
        t.Fatalf("error finding auction by id: %v", err)
    }

    if auctionClosed.Status != auction_entity.Completed {
        t.Fatalf("expected auction status Completed, got %v", auctionClosed.Status)
    }

    // cleanup
    _, _ = auctionRepo.Collection.DeleteOne(ctx, map[string]interface{}{"_id": auctionEntity.Id})
}
