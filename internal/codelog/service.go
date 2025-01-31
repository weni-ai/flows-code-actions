package codelog

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/weni-ai/flows-code-actions/config"
)

type Service struct {
	repo Repository
}

func NewCodeLogService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, codelog *CodeLog) (*CodeLog, error) {
	return s.repo.Create(ctx, codelog)
}

func (s *Service) GetByID(ctx context.Context, id string) (*CodeLog, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListRunLogs(ctx context.Context, id string) ([]CodeLog, error) {
	return s.repo.ListRunLogs(ctx, id)
}

func (s *Service) Update(ctx context.Context, id string, content string) (*CodeLog, error) {
	return s.repo.Update(ctx, id, content)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) StartCodeLogCleaner(cfg *config.Config) error {
	scheduleTime := cfg.Cleaner.ScheduleTime // default is "01:00"
	layout := "15:05"
	retentionPeriod, _ := strconv.Atoi(cfg.Cleaner.RetentionPeriod)
	go func() {
		ticker := time.NewTicker(time.Minute * 2)
		for {
			<-ticker.C
			now := time.Now()
			t, _ := time.Parse(layout, scheduleTime)
			startTime := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
			if now.After(startTime) && now.Before(startTime.Add(time.Hour)) {
				ctx := context.Background()
				retentionPeriod := 24 * time.Duration(retentionPeriod) * time.Hour // 30 days is the default
				currentTime := time.Now()
				retentionLimit := currentTime.Add(-retentionPeriod)
				deletedCount, err := s.repo.DeleteOlder(ctx, retentionLimit, 1000)
				if err != nil {
					logrus.Error("error on running codelog cleaner", err.Error())
				} else {
					logrus.Info(fmt.Sprintf("deleted %d logs from codelog\n", deletedCount))
				}
			}
		}
	}()
	return nil
}
