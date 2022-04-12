package provider

import (
	"embed"
	"fmt"
	"os"
	"strconv"

	"github.com/livekit/protocol/livekit"
)

type videoSpec struct {
	prefix string
	height int
	kbps   int
	fps    int
}

func videoToPublishSpec(height, kbps, fps int) *videoSpec {
	return &videoSpec{
		prefix: "video",
		height: height,
		kbps:   kbps,
		fps:    fps,
	}
}

func beachSpec(height, kbps, fps int) *videoSpec {
	return &videoSpec{
		prefix: "beach",
		height: height,
		kbps:   kbps,
		fps:    fps,
	}
}

func (v *videoSpec) Name() string {
	// pwd, _ := os.Getwd()
	// pwd = strings.TrimSpace(pwd)
	height := strconv.Itoa(v.height)
	kbps := strconv.Itoa(v.kbps)
	fps := strconv.Itoa(v.fps)
	return fmt.Sprintf("pkg/provider/randomVideo/%s_%s_%s_%s.h264" , v.prefix, height, kbps, fps)
}

func (v *videoSpec) ToVideoLayer(quality livekit.VideoQuality) *livekit.VideoLayer {
	return &livekit.VideoLayer{
		Quality: quality,
		Height:  uint32(v.height),
		Width:   uint32(v.height * 16 / 9),
		Bitrate: v.bitrate(),
	}
}

func (v *videoSpec) bitrate() uint32 {
	return uint32(v.kbps * 1000)
}

var (
	//go:embed resources
	res embed.FS

	// map of key => bitrate
	videosToPublish = []*videoSpec{
		videoToPublishSpec(180, 150, 15),
		videoToPublishSpec(360, 400, 20),
		videoToPublishSpec(540, 800, 25),
		videoToPublishSpec(720, 2000, 30),
		videoToPublishSpec(1080, 3000, 30),
	}

	beachFiles = []*videoSpec{
		beachSpec(180, 150, 15),
		beachSpec(360, 400, 20),
		beachSpec(540, 800, 25),
		beachSpec(720, 2000, 30),
		beachSpec(1080, 3000, 30),
	}

)

func ButterflyLooper(height int) (*H264VideoLooper, error) {
	var spec *videoSpec

	for _, s := range videosToPublish {
		if s.height == height {
			spec = s
			break
		}
	}

	if spec == nil {
		return nil, os.ErrNotExist
	}

	f, err := res.Open(spec.Name())
	
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return NewH264VideoLooper(f, spec)
}

func ButterflyLooperForBitrate(bitrate uint32) (*H264VideoLooper, error) {
	var spec *videoSpec
	//for _, s := range butterflyFiles {
	for _, s := range beachFiles {		
		spec = s
		if s.bitrate() >= bitrate {
			break
		}
	}
	if spec == nil {
		return nil, os.ErrNotExist
	}
	f, err := res.Open(spec.Name())
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewH264VideoLooper(f, spec)
}
