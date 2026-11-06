package scheduler

import (
	"context"
	"log"
	"time"

	"example.com/hris/backend/internal/usecase"
)

type GpsCleanerScheduler struct {
	usecase  usecase.GpsUsecase
	interval time.Duration
}

func NewGpsCleanerScheduler(u usecase.GpsUsecase, interval time.Duration) *GpsCleanerScheduler {
	return &GpsCleanerScheduler{
		usecase:  u,
		interval: interval,
	}
}

func (s *GpsCleanerScheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("[Scheduler] Running GPS locations cleaner...")
				if err := s.usecase.ClearOldLocations(); err != nil {
					log.Printf("[Scheduler] Error clearing old GPS locations: %v", err)
				} else {
					log.Println("[Scheduler] Successfully cleared old GPS locations")
				}
			case <-ctx.Done():
				ticker.Stop()
				log.Println("[Scheduler] GPS cleaner stopped")
				return
			}
		}
	}()
}
