package repository

import (
	"context"
	"fmt"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// getNextSequence genera un ID secuencial usando un contador atómico en MongoDB
func getNextSequence(ctx context.Context, db *mongo.Database, collectionName string) (uint64, error) {
	// Usar la colección "counters" para mantener contadores por colección
	counterCol := db.Collection("counters")
	filter := bson.M{"_id": collectionName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	
	var result struct {
		ID  string `bson:"_id"`
		Seq uint64 `bson:"seq"`
	}
	
	// Intentar actualizar el contador (si existe) o crear uno nuevo
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	err := counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		// Si no existe, crear con valor inicial 0, luego incrementar a 1
		if err == mongo.ErrNoDocuments {
			_, insertErr := counterCol.InsertOne(ctx, bson.M{"_id": collectionName, "seq": 0})
			if insertErr != nil {
				// Puede que ya exista (race condition), intentar de nuevo
				err = counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
				if err != nil {
					return 0, fmt.Errorf("failed to get next sequence: %w", err)
				}
				return result.Seq, nil
			}
			// Ahora intentar de nuevo para obtener 1
			err = counterCol.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
			if err != nil {
				return 0, fmt.Errorf("failed to get next sequence after init: %w", err)
			}
		} else {
			return 0, fmt.Errorf("failed to get next sequence: %w", err)
		}
	}
	
	return result.Seq, nil
}

// parseUint64 convierte un string a uint64, útil para parsear IDs de URLs
func parseUint64(idStr string) (uint64, error) {
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id format: %w", err)
	}
	return id, nil
}

// uint64ToString convierte uint64 a string (útil para eventos y JSON)
func uint64ToString(id uint64) string {
	return strconv.FormatUint(id, 10)
}
