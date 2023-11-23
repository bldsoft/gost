package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
)

// const defaultTimeOut = 30 * time.Second

func probeCall(ctx context.Context, filename string, args args) (string, error) {
	args["of"] = "json"
	cmdArgs := args.toCmdArgs()
	cmdArgs = append(cmdArgs, filename)
	cmd := exec.CommandContext(ctx, "ffprobe", cmdArgs...)
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func Probe(ctx context.Context, path string, args map[string]interface{}) (*FFMpegProbe, error) {
	probeRaw, err := probeCall(ctx, path, args)
	if err != nil {
		return nil, err
	}
	var res FFMpegProbe
	err = json.Unmarshal([]byte(probeRaw), &res)
	return &res, err
}

func ProbeInto(ctx context.Context, path string, res interface{}, args map[string]interface{}) error {
	probeRaw, err := probeCall(ctx, path, args)
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

type args map[string]interface{}

func (a args) toCmdArgs() []string {
	var keys, args []string
	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := a[k]
		switch a := v.(type) {
		case string:
			args = append(args, fmt.Sprintf("-%s", k))
			if a != "" {
				args = append(args, a)
			}
		case []string:
			for _, r := range a {
				args = append(args, fmt.Sprintf("-%s", k))
				if r != "" {
					args = append(args, r)
				}
			}
		case []int:
			for _, r := range a {
				args = append(args, fmt.Sprintf("-%s", k))
				args = append(args, strconv.Itoa(r))
			}
		case int:
			args = append(args, fmt.Sprintf("-%s", k))
			args = append(args, strconv.Itoa(a))
		default:
			args = append(args, fmt.Sprintf("-%s", k))
			args = append(args, fmt.Sprintf("%v", a))
		}
	}
	return args
}
