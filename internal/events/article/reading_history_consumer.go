package article

import (
	"context"
	"github.com/IBM/sarama"
	"red-feed/internal/domain"
	"red-feed/internal/repository"
	"red-feed/pkg/logger"
	"red-feed/pkg/saramax"
	"time"
)

type HistoryConsumer struct {
	client sarama.Client
	repo   repository.HistoryRecordRepository
	l      logger.Logger
}

func (hc *HistoryConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("history",
		hc.client)
	if err != nil {
		return err
	}
	go func() {
		err := cg.Consume(context.Background(),
			[]string{topicReadEvent},
			saramax.NewHandler[ReadEvent](hc.l, hc.Consume))
		if err != nil {
			hc.l.Error("退出了消费循环异常", logger.Error(err))
		}
	}()
	return err
}

func (hc *HistoryConsumer) Consume(
	msg *sarama.ConsumerMessage,
	evt ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return hc.repo.AddRecord(ctx, domain.HistoryRecord{
		Uid:   evt.Uid,
		Biz:   "article",
		BizId: evt.Aid,
	})
}
