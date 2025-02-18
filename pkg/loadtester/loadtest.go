package loadtester

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	lksdk "github.com/livekit/server-sdk-go"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type LoadTest struct {
	Params     Params
	trackNames map[string]string
	lock       sync.Mutex
}

type Params struct {
	Context      context.Context
	Publishers   int
	Subscribers  int
	AudioBitrate uint32
	VideoBitrate uint32
	Duration     time.Duration
	// number of seconds to spin up per second
	NumPerSecond float64
	Simulcast    bool
	Identities 		[]string
	TesterParams
}

func NewLoadTest(params Params) *LoadTest {
	l := &LoadTest{
		Params:     params,
		trackNames: make(map[string]string),
	}

	if len(params.IdentityRange) != 0 {

		log.Printf("Found identity range %s", params.IdentityRange)

		var err error

		l.Params.Identities, err = getRangeFromIdentityRange(params.IdentityRange)
		l.Params.Publishers = len(l.Params.Identities)

		if err != nil {
			panic(err)
		}
	}

	if l.Params.NumPerSecond == 0 {
		// sane default
		l.Params.NumPerSecond = 5
	}
	if l.Params.NumPerSecond > 10 {
		l.Params.NumPerSecond = 10
	}
	if l.Params.Publishers == 0 && l.Params.Subscribers == 0 {
		l.Params.Publishers = 1
		l.Params.Subscribers = 1
	}
	return l
}

func (t *LoadTest) Run() error {

	stats, err := t.run(t.Params)
	if err != nil {
		return err
	}

	// tester results
	summaries := make(map[string]*summary)
	names := make([]string, 0, len(stats))
	
	for name := range stats {
		if strings.HasPrefix(name, "Pub") {
			continue
		}
		names = append(names, name)
	}

	sort.Strings(names)
	
	for _, name := range names {
		testerStats := stats[name]
		summaries[name] = getTesterSummary(testerStats)

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintf(w, "\n%s\t| Track\t| Kind\t| Pkts\t| Bitrate\t| Latency\t| Dropped\n", name)
		trackStatsSlice := make([]*trackStats, 0, len(testerStats.trackStats))
		for _, ts := range testerStats.trackStats {
			trackStatsSlice = append(trackStatsSlice, ts)
		}
		sort.Slice(trackStatsSlice, func(i, j int) bool {
			nameI := t.trackNames[trackStatsSlice[i].trackID]
			nameJ := t.trackNames[trackStatsSlice[j].trackID]
			return strings.Compare(nameI, nameJ) < 0
		})
		for _, trackStats := range trackStatsSlice {
			latency, dropped := formatStrings(
				trackStats.packets.Load(), trackStats.latency.Load(), trackStats.latencyCount.Load(), trackStats.dropped.Load())

			trackName := t.trackNames[trackStats.trackID]
			_, _ = fmt.Fprintf(w, "\t| %s %s\t| %s\t| %d\t| %s\t| %s\t| %s\n",
				trackName, trackStats.trackID, trackStats.kind, trackStats.packets.Load(),
				formatBitrate(trackStats.bytes.Load(), time.Since(trackStats.startedAt.Load())), latency, dropped)
		}
		_ = w.Flush()
	}

	if len(summaries) == 0 {
		return nil
	}

	// summary
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprint(w, "\nSummary\t| Tester\t| Tracks\t| Bitrate\t| Latency\t| Total Dropped\n")

	for _, name := range names {
		s := summaries[name]
		sLatency, sDropped := formatStrings(s.packets, s.latency, s.latencyCount, s.dropped)
		sBitrate := formatBitrate(s.bytes, s.elapsed)
		_, _ = fmt.Fprintf(w, "\t| %s\t| %d/%d\t| %s\t| %s\t| %s\n",
			name, s.tracks, s.expected, sBitrate, sLatency, sDropped)
	}

	s := getTestSummary(summaries)
	sLatency, sDropped := formatStrings(s.packets, s.latency, s.latencyCount, s.dropped)
	// avg bitrate per sub
	sBitrate := fmt.Sprintf("%s (%s avg)",
		formatBitrate(s.bytes, s.elapsed),
		formatBitrate(s.bytes/int64(len(summaries)), s.elapsed),
	)
	_, _ = fmt.Fprintf(w, "\t| %s\t| %d/%d\t| %s\t| %s\t| %s\n",
		"Total", s.tracks, s.expected, sBitrate, sLatency, sDropped)

	_ = w.Flush()
	return nil
}

func (t *LoadTest) RunSuite() error {
	cases := []*struct {
		publishers  int
		subscribers int
		video       bool

		tracks  int64
		latency time.Duration
		dropped float64
	}{
		{publishers: 10, subscribers: 10, video: false},
		{publishers: 10, subscribers: 100, video: false},
		{publishers: 10, subscribers: 500, video: false},
		{publishers: 10, subscribers: 1000, video: false},
		{publishers: 50, subscribers: 50, video: false},
		{publishers: 100, subscribers: 50, video: false},

		{publishers: 10, subscribers: 10, video: true},
		{publishers: 10, subscribers: 100, video: true},
		{publishers: 10, subscribers: 500, video: true},
		{publishers: 1, subscribers: 100, video: true},
		{publishers: 1, subscribers: 1000, video: true},
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	_, _ = fmt.Fprint(w, "\nPubs\t| Subs\t| Tracks\t| Audio\t| Video\t| Packet loss\n")

	for _, c := range cases {
		caseParams := t.Params
		caseParams.Publishers = c.publishers
		caseParams.Subscribers = c.subscribers
		caseParams.Simulcast = true
		if caseParams.Duration == 0 {
			caseParams.Duration = 15 * time.Second
		}
		videoString := "Yes"
		if !c.video {
			caseParams.VideoBitrate = 0
			videoString = "No"
		}
		fmt.Printf("\nRunning test: %d pub, %d sub, video: %s\n", c.publishers, c.subscribers, videoString)

		stats, err := t.run(caseParams)
		if err != nil {
			return err
		}
		if t.Params.Context.Err() != nil {
			return err
		}

		var tracks, packets, dropped, totalLatency, latencyCount int64
		for _, testerStats := range stats {
			for _, trackStats := range testerStats.trackStats {
				tracks++
				packets += trackStats.packets.Load()
				dropped += trackStats.dropped.Load()
				totalLatency += trackStats.latency.Load()
				latencyCount += trackStats.latencyCount.Load()
			}
		}
		_, _ = fmt.Fprintf(w, "%d\t| %d\t| %d\t| Yes\t| %s\t| %.3f%%\n",
			c.publishers, c.subscribers, tracks, videoString, 100*float64(dropped)/float64(dropped+packets))
	}

	_ = w.Flush()
	return nil
}

func (t *LoadTest) run(params Params) (map[string]*testerStats, error) {
	if params.Room == "" {
		params.Room = fmt.Sprintf("testroom%d", rand.Int31n(1000))
	}

	if t.Params.Identities == nil || len(t.Params.Identities) == 0 {
		params.IdentityPrefix = randStringRunes(5)
	} else {
		params.Identities = t.Params.Identities
		params.Publishers = t.Params.Publishers
	}

	expectedTracks := params.Publishers
	if params.VideoBitrate > 0 {
		expectedTracks *= 2
	}

	log.Printf("Run test with %v identities", params.Identities)

	fmt.Printf("Starting load test with %d publishers, %d subscribers, room: %s\n",
		t.Params.Publishers, t.Params.Subscribers, t.Params.Room)

	testers := make([]*LoadTester, 0)
	group, _ := errgroup.WithContext(t.Params.Context)
	startedAt := time.Now()
	numStarted := float64(0)
	
	for i := 0; i < params.Publishers+params.Subscribers - 1; i++ {
		testerParams := params.TesterParams
		testerParams.sequence = i
		testerParams.expectedTracks = expectedTracks
		isPublisher := i < params.Publishers
		
		if isPublisher {
			if params.VideoBitrate > 0 {
				testerParams.expectedTracks -= 2
			} else {
				testerParams.expectedTracks--
			}
			
			/// If names are present (as range of ids, use that as name)
			if params.Identities != nil && len(params.Identities) != 0 {
				testerParams.name = fmt.Sprintf(params.Identities[i])
			} else {
				testerParams.name = fmt.Sprintf("Pub %d", i)
			}

		} else {
			testerParams.Subscribe = true
			testerParams.name = fmt.Sprintf("Sub %d", i-params.Publishers)
		}

		tester := NewLoadTester(testerParams)
		testers = append(testers, tester)

		group.Go(func() error {

			if err := tester.Start(); err != nil {
				return errors.Wrapf(err, "could not connect %s", testerParams.name)
			}

			if isPublisher {
				if params.AudioBitrate > 0 {
					audio, err := tester.PublishTrack("audio", lksdk.TrackKindAudio, params.AudioBitrate)
					if err != nil {
						return err
					}
					t.lock.Lock()
					t.trackNames[audio] = fmt.Sprintf("%dA", testerParams.sequence)
					t.lock.Unlock()
				}

				if params.VideoBitrate > 0 {
					var video string
					var err error
					if params.Simulcast {
						log.Println("\nVideo simulcast..")
						video, err = tester.PublishSimulcastTrack("video-simulcast", params.VideoBitrate, i)
					} else {
						video, err = tester.PublishTrack("video", lksdk.TrackKindVideo, params.VideoBitrate)
					}
					if err != nil {
						return err
					}
					t.lock.Lock()
					t.trackNames[video] = fmt.Sprintf("%dV", testerParams.sequence)
					t.lock.Unlock()
				}
			}
			return nil
		})
		numStarted++

		for {
			secondsElapsed := float64(time.Since(startedAt)) / float64(time.Second)
			startRate := numStarted / secondsElapsed
			if err := t.Params.Context.Err(); err != nil {
				return nil, err
			}
			if startRate > params.NumPerSecond {
				time.Sleep(time.Second)
			} else {
				break
			}
		}
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}

	duration := params.Duration
	if duration == 0 {
		// a really long time
		duration = 1000 * time.Hour
	}
	select {
	case <-params.Context.Done():
		// canceled
	case <-time.After(duration):
		// finished
	}

	stats := make(map[string]*testerStats)
	for _, t := range testers {
		t.Stop()
		stats[t.params.name] = t.GetStats()
	}

	return stats, nil
}
