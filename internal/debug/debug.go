package debug

import (
	"fmt"
	"log"
	"strconv"
	"time"
)

// variables that get substituted by the deployment script (-ldflags).
var (
	BuildHash = "dev"
	BuildDate = fmt.Sprintf("%d", time.Now().Unix())
)

func FormattedBuildDate() string {
	epoch, err := strconv.ParseInt(BuildDate, 10, 64)
	if err != nil {
		log.Fatalf("failed to parse build date: %s\n", err)
	}

	return time.Unix(epoch, 0).Format(time.DateTime)
}
