package hls

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	infoTag = "#EXTINF:"
)

// Get a first media manifest that is not a subtitles
func GetOneChildPlaylist(master []byte) ([]byte, error) {
	re := regexp.MustCompile(`(\S*\.m3u8$)$`)
	child := re.Find(master)
	if len(child) == 0 {
		return nil, fmt.Errorf("no child playlist found")
	}
	return child, nil
}

func CalculateM3U8Duration(child []byte) (float64, error) {
	re := regexp.MustCompile(infoTag + `([0-9(.*)\- ]+)?`)
	infos := re.FindAllString(string(child), -1)
	if len(infos) == 0 {
		return 0, fmt.Errorf("no info tags found")
	}
	var res float64
	for _, meta := range infos {
		dur, err := strconv.ParseFloat(meta[len(infoTag):], 64)
		if err != nil {
			return 0, err
		}
		res += dur
	}
	return res, nil
}

func GetOneSegment(childPlaylist []byte) ([]byte, error) {
	re := regexp.MustCompile(`(\S*\.(ts|m4a|m4s))`)
	child := re.Find(childPlaylist)
	if len(child) == 0 {
		return nil, fmt.Errorf("no chuncks found")
	}
	return child, nil
}
