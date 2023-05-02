package progress

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/brattonross/ghostedbot/internal/discord"
)

// Percentage returns the percentage of the year that has passed.
func Percentage(timestamp int64) float64 {
	t := time.Unix(timestamp, 0)
	year := t.UTC().Year()
	yearStart := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	yearEnd := time.Date(year, 12, 31, 11, 59, 59, 0, time.UTC)
	yearDuration := yearEnd.Sub(yearStart)
	elapsed := t.Sub(yearStart)
	return math.Round(elapsed.Seconds() / yearDuration.Seconds() * 100)
}

const (
	filled = "█"
	empty  = "░"
)

func ToBar(percentage float64) string {
	filledCount := int(percentage / 10)
	emptyCount := 10 - filledCount
	return fmt.Sprintf("%s%s", strings.Repeat(filled, filledCount), strings.Repeat(empty, emptyCount))
}

func PercentageHandler(ctx *discord.InteractionContext) (*discord.InteractionResponse, error) {
	percentage := Percentage(time.Now().UTC().Unix())
	return discord.MessageResponse(fmt.Sprintf("%s %v%%", ToBar(percentage), percentage)), nil
}
