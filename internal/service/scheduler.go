package service

import (
	"log"
	"time"

	"github.com/spf13/viper"

	"wechatoarss/internal/store"
)

type SchedulerService struct {
	fetcherSvc *FetcherService
	stopChan   chan bool
}

func NewSchedulerService(fetcherSvc *FetcherService) *SchedulerService {
	return &SchedulerService{
		fetcherSvc: fetcherSvc,
		stopChan:   make(chan bool),
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start() error {
	times := viper.GetStringSlice("scheduler.times")
	if len(times) == 0 {
		times = []string{"07:00", "12:00", "20:00"}
	}

	log.Printf("Scheduler started with times: %v", times)

	// Start ticker
	go s.run(times)

	return nil
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	log.Println("Stopping scheduler...")
	s.stopChan <- true
}

func (s *SchedulerService) run(times []string) {
	// Parse times
	var hourMins []struct {
		hour int
		min  int
	}

	for _, t := range times {
		parsed, err := time.Parse("15:04", t)
		if err != nil {
			log.Printf("Invalid time format: %s", t)
			continue
		}
		hourMins = append(hourMins, struct {
			hour int
			min  int
		}{parsed.Hour(), parsed.Minute()})
	}

	if len(hourMins) == 0 {
		log.Println("No valid times configured, using default")
		hourMins = []struct {
			hour int
			min  int
		}{{7, 0}, {12, 0}, {20, 0}}
	}

	// Run immediately on start
	s.runFetch()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for _, hm := range hourMins {
				if now.Hour() == hm.hour && now.Minute() == hm.min {
					log.Printf("Scheduled fetch triggered at %02d:%02d", hm.hour, hm.min)
					s.runFetch()
				}
			}
		case <-s.stopChan:
			log.Println("Scheduler stopped")
			return
		}
	}
}

func (s *SchedulerService) runFetch() {
	// Run fetch in background
	go func() {
		log.Println("Starting scheduled fetch...")
		if err := s.fetcherSvc.FetchAll(); err != nil {
			log.Printf("Scheduled fetch failed: %v", err)
		} else {
			log.Println("Scheduled fetch completed")
		}
	}()
}

// TriggerManualFetch triggers a manual fetch
func (s *SchedulerService) TriggerManualFetch(bizID string) error {
	if bizID != "" {
		return s.fetcherSvc.FetchChannel(bizID)
	}
	return s.fetcherSvc.FetchAll()
}

// GetSchedulerStatus returns scheduler status
func (s *SchedulerService) GetSchedulerStatus() map[string]interface{} {
	times := viper.GetStringSlice("scheduler.times")
	if len(times) == 0 {
		times = []string{"07:00", "12:00", "20:00"}
	}

	return map[string]interface{}{
		"enabled":   true,
		"times":     times,
		"lastRun":   time.Now().Format("2006-01-02 15:04:05"),
		"nextRun":   "N/A",
	}
}

// UpdateSchedulerTimes updates scheduler times
func (s *SchedulerService) UpdateSchedulerTimes(times []string) error {
	viper.Set("scheduler.times", times)
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	log.Printf("Scheduler times updated to: %v", times)
	return nil
}
