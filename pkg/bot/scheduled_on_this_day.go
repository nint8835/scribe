package bot

import (
	"fmt"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/go-co-op/gocron/v2"

	"github.com/nint8835/scribe/pkg/config"
)

func (b *Bot) registerOnThisDayJob(scheduler gocron.Scheduler) error {
	if config.Instance.OnThisDayChannelId == "" {
		slog.Warn("On this day job disabled: SCRIBE_ON_THIS_DAY_CHANNEL_ID is not set")
		return nil
	}

	_, err := scheduler.NewJob(
		gocron.CronJob(config.Instance.OnThisDayCron, false),
		gocron.NewTask(func() {
			if err := b.postOnThisDay(); err != nil {
				slog.Error("error posting on this day quote", "error", err)
			}
		}),
	)
	if err != nil {
		return fmt.Errorf("error scheduling on this day quote: %w", err)
	}

	slog.Info(
		"On this day job registered",
		"cron", config.Instance.OnThisDayCron,
		"channel_id", config.Instance.OnThisDayChannelId,
	)

	return nil
}

func (b *Bot) postOnThisDay() error {
	embed, err := b.makeOnThisDayEmbed(config.Instance.GuildId)
	if err != nil {
		return fmt.Errorf("error generating on this day embed: %w", err)
	}

	_, err = b.Session.ChannelMessageSendComplex(config.Instance.OnThisDayChannelId, &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	if err != nil {
		return fmt.Errorf("error sending on this day message: %w", err)
	}

	return nil
}
