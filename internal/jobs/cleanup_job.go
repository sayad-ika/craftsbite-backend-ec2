package jobs

import (
	"craftsbite-backend/internal/repository"
	"craftsbite-backend/pkg/logger"
	"fmt"

	"github.com/robfig/cron/v3"
)

// CleanupJob handles automatic cleanup of old history records
type CleanupJob struct {
	historyRepo     repository.HistoryRepository
	retentionMonths int
}

// NewCleanupJob creates a new cleanup job
func NewCleanupJob(historyRepo repository.HistoryRepository, retentionMonths int) *CleanupJob {
	return &CleanupJob{
		historyRepo:     historyRepo,
		retentionMonths: retentionMonths,
	}
}

// Run executes the cleanup job
func (j *CleanupJob) Run() {
	logger.Info(fmt.Sprintf("Starting history cleanup job (retention: %d months)", j.retentionMonths))

	deleted, err := j.historyRepo.DeleteOlderThan(j.retentionMonths)
	if err != nil {
		logger.Error(fmt.Sprintf("History cleanup failed: %v", err))
		return
	}

	logger.Info(fmt.Sprintf("History cleanup completed: %d records deleted", deleted))
}

// StartScheduler starts the cron scheduler for the cleanup job
func (j *CleanupJob) StartScheduler(cronSchedule string) (*cron.Cron, error) {
	c := cron.New()

	_, err := c.AddFunc(cronSchedule, j.Run)
	if err != nil {
		return nil, fmt.Errorf("failed to schedule cleanup job: %w", err)
	}

	c.Start()
	logger.Info(fmt.Sprintf("Cleanup job scheduler started (schedule: %s)", cronSchedule))

	return c, nil
}

// StopScheduler stops the cron scheduler gracefully
func StopScheduler(c *cron.Cron) {
	if c != nil {
		ctx := c.Stop()
		<-ctx.Done()
		logger.Info("Cleanup job scheduler stopped")
	}
}
