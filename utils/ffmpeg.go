package utils

import (
	"encoding/json"
	"strconv"
	"time"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const defaultTimeOut = 30 * time.Second

func Probe(path string) (*FFMpegProbe, error) {
	probeRaw, err := ffmpeg.Probe(path)
	if err != nil {
		return nil, err
	}
	var res FFMpegProbe
	err = json.Unmarshal([]byte(probeRaw), &res)
	return &res, err
}

func ProbeWithArgs(path string, args map[string]interface{}) (*FFMpegProbe, error) {
	jsonFormat := ffmpeg.KwArgs{
		"of": "json",
	}

	probeRaw, err := ffmpeg.ProbeWithTimeoutExec(
		path,
		30*time.Second,
		ffmpeg.MergeKwArgs([]ffmpeg.KwArgs{jsonFormat, ffmpeg.KwArgs(args)}),
	)
	if err != nil {
		return nil, err
	}
	var res FFMpegProbe
	err = json.Unmarshal([]byte(probeRaw), &res)
	return &res, err
}

func ProbeInto(path string, res interface{}, args map[string]interface{}) error {
	return ProbeIntoWithTimeout(path, res, args, defaultTimeOut)
}

func ProbeIntoWithTimeout(path string, res interface{}, args map[string]interface{}, timeout time.Duration) error {
	jsonFormat := ffmpeg.KwArgs{
		"of": "json",
	}

	probeRaw, err := ffmpeg.ProbeWithTimeoutExec(
		path,
		timeout,
		ffmpeg.MergeKwArgs([]ffmpeg.KwArgs{jsonFormat, ffmpeg.KwArgs(args)}),
	)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(probeRaw), &res)
}

type FFMpegProbe struct {
	Format struct {
		Duration       string `json:"duration"`
		Filename       string `json:"filename"`
		NbStreams      int    `json:"nb_streams"`
		NbPrograms     int    `json:"nb_programs"`
		FormatName     string `json:"format_name"`
		FormatLongName string `json:"format_long_name"`
		StartTime      string `json:"start_time"`
		Size           string `json:"size"`
		BitRate        string `json:"bit_rate"`
		ProbeScore     int    `json:"probe_score"`
	} `json:"format"`
}

func (p FFMpegProbe) Duration() (float64, error) {
	return strconv.ParseFloat(p.Format.Duration, 64)
}
