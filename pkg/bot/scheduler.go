package bot

import (
	"fmt"
	"log/slog"

	"github.com/go-co-op/gocron/v2"

	"github.com/nint8835/scribe/pkg/config"
)

func (b *Bot) startScheduler() error {
	scheduler, err := gocron.NewScheduler(gocron.WithLocation(config.Location))
	if err != nil {
		return fmt.Errorf("error creating scheduler: %w", err)
	}

	if err = b.registerScheduledJobs(scheduler); err != nil {
		if shutdownErr := scheduler.Shutdown(); shutdownErr != nil {
			slog.Error("error shutting down scheduler", "error", shutdownErr)
		}

		return fmt.Errorf("error registering scheduled jobs: %w", err)
	}

	scheduler.Start()
	b.scheduler = scheduler

	slog.Info("Scheduler running", "job_count", len(scheduler.Jobs()))

	return nil
}

func (b *Bot) registerScheduledJobs(scheduler gocron.Scheduler) error {
	if err := b.registerOnThisDayJob(scheduler); err != nil {
		return fmt.Errorf("error registering on this day job: %w", err)
	}

	return nil
}
