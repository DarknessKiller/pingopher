package dto_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DarknessKiller/pingopher/internal/dto"
	"github.com/DarknessKiller/pingopher/internal/model"
)

func TestToHistoriesUsesSharedPingTime(t *testing.T) {
	pingTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	histories := []*model.History{
		{DNS: model.DNS{Name: "a"}, PingDateTime: sql.NullTime{Time: pingTime, Valid: true}},
		{DNS: model.DNS{Name: "b"}, PingDateTime: sql.NullTime{Time: pingTime, Valid: true}},
	}
	histories[0].CreatedAt = pingTime.Add(time.Millisecond)
	histories[1].CreatedAt = pingTime.Add(2 * time.Millisecond)

	results := dto.ToHistories(histories).Results
	if !results[0].Timestamp.Equal(pingTime) {
		t.Fatalf("timestamp = %v, want shared ping time %v", results[0].Timestamp, pingTime)
	}
	if !results[0].Timestamp.Equal(results[1].Timestamp) {
		t.Fatalf("resolver timestamps differ: %v vs %v", results[0].Timestamp, results[1].Timestamp)
	}
}

func TestToHistoriesFallsBackToCreatedAt(t *testing.T) {
	createdAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	histories := []*model.History{{DNS: model.DNS{Name: "a"}}}
	histories[0].CreatedAt = createdAt

	results := dto.ToHistories(histories).Results
	if !results[0].Timestamp.Equal(createdAt) {
		t.Fatalf("timestamp = %v, want CreatedAt fallback %v", results[0].Timestamp, createdAt)
	}
}
