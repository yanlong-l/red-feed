package repository

import (
	"context"
	"red-feed/internal/domain"
)

type HistoryRecordRepository interface {
	AddRecord(ctx context.Context, r domain.HistoryRecord) error
}
