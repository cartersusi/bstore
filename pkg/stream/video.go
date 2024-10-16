package stream

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cartersusi/bstore/pkg/cmd"
	"github.com/cartersusi/bstore/pkg/fops"
	"github.com/gin-gonic/gin"
)

const (
	DASH = iota
	HLS
	POSTER
)

var MethodFMap = map[int]string{
	DASH:   "index.mpd",
	HLS:    "index.m3u8",
	POSTER: "index.jpg",
}
var VidEXT = []string{".mp4", ".webm", ".ogg", ".wmv", ".mov", ".avchd", ".av1"}

type VideoEncoder struct {
	InputFile  string
	OutputDir  string
	OutputFile string
	StreamType int
	Codec      string
	Audio      bool
	Command    string
}

func MakeUrl(c *gin.Context, fpath string, method int) string {
	tls := "https://"
	is_https := c.Request.TLS
	if is_https == nil {
		tls = "http://"
	}

	_, fname := filepath.Split(fpath)
	dir_name := strings.TrimSuffix(fname, filepath.Ext(fname))
	stream_fname := MethodFMap[method]

	return fmt.Sprintf("%s%s/bstore/%s/%s", tls, c.Request.Host, dir_name, stream_fname)
}

func (v *VideoEncoder) VideoBuilder(method int) error {
	log.Println("Building video for", method)
	v.StreamType = method

	err := v.CheckAll()
	if err != nil {
		return err
	}

	v.SetOutput()
	v.SetCommand()
	v.Print()

	return nil
}

func (v *VideoEncoder) Print() {
	fmt.Println("InputFile:", v.InputFile)
	fmt.Println("OutputDir:", v.OutputDir)
	fmt.Println("OutputFile:", v.OutputFile)
	fmt.Println("Codec:", v.Codec)
	fmt.Println("Audio:", v.Audio)
}

func (v *VideoEncoder) SetOutput() {
	v.SetOutputDir()
	v.SetOutputFile()
	v.CheckAudio()
}

func (v *VideoEncoder) SetOutputDir() {
	dname, fname := filepath.Split(v.InputFile)
	ext := filepath.Ext(fname)
	fname = strings.TrimSuffix(fname, ext)
	v.OutputDir = filepath.Join(dname, fname)
	os.MkdirAll(v.OutputDir, os.ModePerm)
}
func (v *VideoEncoder) SetOutputFile() {
	v.OutputFile = fmt.Sprintf("%s/%s", v.OutputDir, MethodFMap[v.StreamType])
}

func (v *VideoEncoder) SetCommand() {
	switch v.StreamType {
	case DASH:
		v.DASHcmd()
	case HLS:
		v.HLScmd()
	}
}

func (v *VideoEncoder) CheckAll() error {
	_, fname := filepath.Split(v.InputFile)
	if !CheckEXT(fname) {
		return errors.New("Invalid file extension")
	}

	if !v.CheckCodec() {
		return errors.New("Invalid codec")
	}

	return nil
}

func (v *VideoEncoder) CheckCodec() bool {
	output, err := cmd.GetCMD("ffmpeg", "-h", "encoder="+v.Codec)
	if err != nil || output == "" {
		return false
	}

	if fops.Contains(output, "is not recognized by FFmpeg") {
		return false
	}

	return true
}

func (v *VideoEncoder) CheckAudio() {
	has_audio, err := cmd.GetCMD("ffprobe", "-i", v.InputFile, "-show_streams", "-select_streams", "a", "-loglevel", "error")
	if err != nil || has_audio == "" {
		v.Audio = false
		return
	}

	v.Audio = true
}

func (v *VideoEncoder) DASHcmd() {
	audio_cmd := "-c:a libopus -b:a 128k"
	segment_cmd := `-dash_segment_type mp4 -adaptation_sets "id=0,streams=v id=1,streams=a"`
	if !v.Audio {
		audio_cmd = ""
		segment_cmd = `-dash_segment_type mp4 -adaptation_sets "id=0,streams=v"`
	}
	v.Command = fmt.Sprintf(`ffmpeg -i %s \
  -map 0 -c:v %s -b:v 1000k -keyint_min 150 -g 150 -sc_threshold 0 %s \
  -f dash -seg_duration 4 -use_template 1 -use_timeline 1 -init_seg_name 'init-$RepresentationID$.m4s' \
  -media_seg_name 'chunk-$RepresentationID$-$Number$.m4s' \
  %s \
  %s`,
		v.InputFile, v.Codec, audio_cmd, segment_cmd, v.OutputFile)
}

func (v *VideoEncoder) HLScmd() {
	audio_cmd := "-c:a aac -b:a 128k"
	segment_cmd := `-var_stream_map "v:0,a:0"`
	if !v.Audio {
		audio_cmd = ""
		segment_cmd = `-var_stream_map "v:0"`
	}
	v.Command = fmt.Sprintf(`ffmpeg -i %s \
  -map 0 -c:v %s -b:v 1000k -keyint_min 150 -g 150 -sc_threshold 0 %s \
  -f hls \
  -hls_time 4 \
  -hls_playlist_type vod \
  -hls_segment_filename %s/segment_%%03d.ts \
  -master_pl_name /master.m3u8 \
  %s \
  %s`,
		v.InputFile, v.Codec, audio_cmd, v.OutputDir, segment_cmd, v.OutputFile)
}

func CheckEXT(fname string) bool {
	ext := filepath.Ext(fname)
	for _, e := range VidEXT {
		if e == ext {
			return true
		}
	}

	return false
}
