package loadtester

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func formatStrings(packets, latency, latencyCount, dropped int64) (sLatency, sDropped string) {
	sLatency = " - "
	sDropped = " - "

	if packets > 0 {
		totalPackets := packets + dropped
		if latencyCount > 0 {
			sLatency = fmt.Sprint(time.Duration(latency / latencyCount))
		}
		sDropped = fmt.Sprintf("%d (%s%%)", dropped, formatPercentage(dropped, totalPackets))
	}

	return
}

func formatPercentage(num int64, total int64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.3f", float64(num)/float64(total)*100), "0"), ".")
}

func formatBitrate(bytes int64, elapsed time.Duration) string {
	bps := float64(bytes*8) / elapsed.Seconds()
	if bps < 1000 {
		return fmt.Sprintf("%dbps", int(bps))
	} else if bps < 1000000 {
		return fmt.Sprintf("%.1fkbps", bps/1000)
	} else {
		return fmt.Sprintf("%.1fmbps", bps/1000000)
	}
}

func getRangeFromIdentityRange(identityRange string) ([]string, error) {
	split := strings.Split(identityRange, "-")

	if len(split) != 2 {
		return nil, errors.New("identity-range must be <int>-<int>")
	}

	var first, last int
	var err error

	if first, err = strconv.Atoi(split[0]); err != nil {
		return nil, errors.New("identity range must be int")
	}

	if last, err = strconv.Atoi(split[1]); err != nil {
		return nil, errors.New("identity range must be int")
	}

	/// Inclusive range
	total := last - first + 1

	log.Printf("Generating %d identities ", total)

	identities := make([]string, total)

	index := 0

	for i := first; i<=last; i++ {
		identities[index] = strconv.Itoa(i)
		index ++
	}

	return identities, nil
}
