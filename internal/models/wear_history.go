package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WearRecord logs an outfit the user actually wore (via "Wear this today").
// ComboKey is the deduplication key of the outfit's top+bottom pair — the
// recommendation engine grants a wear-history bonus when a candidate
// combination shares this key.
type WearRecord struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID   primitive.ObjectID `bson:"userId" json:"userId"`
	ComboKey string             `bson:"comboKey" json:"comboKey"`
	ItemIDs  []string           `bson:"itemIds" json:"itemIds"`
	WornAt   time.Time          `bson:"wornAt" json:"wornAt"`
}
