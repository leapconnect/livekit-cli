package provider

import (
	"errors"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)


func getVideoSpecsFromRandomVideo() ([]videoSpec, error) {
	pwd, _ := os.Getwd()
	
	log.Printf("Currently working directory: %s", pwd )
	
	_files, err := ioutil.ReadDir("pkg/provider/randomVideo/")

	var files []fs.FileInfo

	for _, f := range(_files) {
		n := f.Name()
		if (n != ".gitkeep") {
			log.Println(n)
			files = append(files, f)
		}
	}

	if err != nil {
		log.Println(err)
		return nil, errors.New("failed retrieving files")
	}

	videoSpecs := make([]videoSpec, len(files))

	for i, f := range(files) {
		n := f.Name()

		if n == ".gitkeep"{
			continue
		}

		// Name for random videos is video<intsequence>_<height>_<kbps>_<fps>.h264
		split := strings.Split(n, "_")
		
		if len(split) != 4 {
			return nil, errors.New("filename not valid")
		}

		prefix := split[0]
		height, _ := strconv.Atoi(split[1])
		kbps, _ := strconv.Atoi(split[2])
		endName := split[3]

		/// end contains fps and file extension

		fps, _ := strconv.Atoi(strings.Split(endName, ".")[0]) 

		v := videoSpec{
			prefix: prefix,
			height: height,
			kbps: kbps,
			fps: fps,
		}
		videoSpecs[i] = v
	}
	return videoSpecs, nil
}

var index = 0

func GetVideoLooperForUserIdentity() (*H264VideoLooper, error){
	var specs []videoSpec
	var err error

	if specs, err = getVideoSpecsFromRandomVideo(); err != nil {
		return nil, err
	}

	totalVideo := len(specs)

	var f fs.File

	log.Printf("%d -> usedVideos %d -> totalVideo", index + 1, totalVideo)

	if index + 1 > totalVideo {	

		log.Println("Not enough video, randomly use already used videos")
		
		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)

		i := r.Int31n(int32(totalVideo - 1))

		specRandom := &specs[i]

		if f, err = os.Open(specRandom.Name()); err != nil {
			return nil, err
		}
		
		return NewH264VideoLooper(f, specRandom)
	}

	spec := &specs[index]

	if f, err = os.Open(spec.Name()); err != nil {
		return nil, err
	}

	index ++

	return NewH264VideoLooper(f, spec)
}

