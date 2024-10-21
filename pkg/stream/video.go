package stream

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cartersusi/bstore/pkg/cmd"
	"github.com/cartersusi/bstore/pkg/fops"
	"github.com/gin-gonic/gin"
)

type GPUType int

const (
	NoGPU GPUType = iota
	NvidiaGPU
	AppleGPU
)

const (
	DASH = iota
	HLS
	POSTER
)

const DEFAULT_BITRATE = 1000

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
	GPUType    GPUType
	Bitrate    string
}

func detectGPU() GPUType {
	if runtime.GOOS == "darwin" {
		output, err := cmd.GetCMD("sysctl", "-n", "machdep.cpu.brand_string")
		if err == nil && strings.Contains(output, "Apple") {
			return AppleGPU
		}
	}

	output, err := cmd.GetCMD("nvidia-smi")
	if err == nil && strings.Contains(output, "NVIDIA-SMI") {
		return NvidiaGPU
	}

	return NoGPU
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

func (v *VideoEncoder) getHWAccelFlags() (string, string) {
	switch v.GPUType {
	case NvidiaGPU:
		gpuInfo, err := cmd.GetCMD("nvidia-smi", "--query-gpu=gpu_name", "--format=csv,noheader")
		if err == nil && strings.Contains(strings.ToLower(gpuInfo), "40") && strings.Contains(v.Codec, "av1") {
			return "-hwaccel cuda -hwaccel_output_format cuda", "av1_nvenc"
		}
		return "-hwaccel cuda -hwaccel_output_format cuda", "h264_nvenc"
	case AppleGPU:
		return "-hwaccel videotoolbox", "h264_videotoolbox"
	default:
		return "", v.Codec
	}
}

func formatBitrate(bitrate int) string {
	if bitrate == 0 {
		return fmt.Sprintf("%dk", DEFAULT_BITRATE)
	}

	return fmt.Sprintf("%dk", bitrate)
}

func (v *VideoEncoder) VideoBuilder(method int) error {
	log.Println("Building video for", method)
	v.StreamType = method
	v.GPUType = detectGPU()

	if v.Codec == "auto" {
		fmt.Println("Auto codec")
		if v.GPUType == NvidiaGPU {
			v.Codec = "libaom-av1"
		} else {
			v.Codec = "libx264"
		}
	}

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
	fmt.Println("Command:", v.Command)
	fmt.Println("GPUType:", v.GPUType)
	fmt.Println("Bitrate:", v.Bitrate)
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
	hwaccel, encoder := v.getHWAccelFlags()
	audio_cmd := "-c:a libopus -b:a 128k"
	segment_cmd := `-dash_segment_type mp4 -adaptation_sets "id=0,streams=v id=1,streams=a"`

	if !v.Audio {
		audio_cmd = ""
		segment_cmd = `-dash_segment_type mp4 -adaptation_sets "id=0,streams=v"`
	}

	// Escape the file paths
	inputFile := shellEscape(v.InputFile)
	outputFile := shellEscape(v.OutputFile)

	v.Command = fmt.Sprintf(`ffmpeg %s -i %s \
        -map 0 -c:v %s -preset p2 -b:v %s -maxrate %s -bufsize %s \
        -keyint_min 150 -g 150 -sc_threshold 0 %s \
        -f dash -seg_duration 4 -use_template 1 -use_timeline 1 \
        -init_seg_name init-\$RepresentationID\$.m4s \
        -media_seg_name chunk-\$RepresentationID\$-\$Number\$.m4s \
        %s \
        %s`,
		hwaccel, inputFile, encoder, v.Bitrate, v.Bitrate, v.Bitrate,
		audio_cmd, segment_cmd, outputFile)
}

func (v *VideoEncoder) HLScmd() {
	hwaccel, encoder := v.getHWAccelFlags()
	audio_cmd := "-c:a aac -b:a 128k"
	segment_cmd := `-var_stream_map "v:0,a:0"`

	if !v.Audio {
		audio_cmd = ""
		segment_cmd = `-var_stream_map "v:0"`
	}

	// Escape the file paths
	inputFile := shellEscape(v.InputFile)
	outputFile := shellEscape(v.OutputFile)
	outputDir := shellEscape(v.OutputDir)

	v.Command = fmt.Sprintf(`ffmpeg %s -i %s \
        -map 0 -c:v %s -preset p2 -b:v %s -maxrate %s -bufsize %s \
        -keyint_min 150 -g 150 -sc_threshold 0 %s \
        -f hls \
        -hls_time 4 \
        -hls_playlist_type vod \
        -hls_segment_filename %s/segment_%%03d.ts \
        -master_pl_name /master.m3u8 \
        %s \
        %s`,
		hwaccel, inputFile, encoder, v.Bitrate, v.Bitrate, v.Bitrate,
		audio_cmd, outputDir, segment_cmd, outputFile)
}

func shellEscape(s string) string {
	s = strings.Replace(s, "'", "'\\''", -1)
	return "'" + s + "'"
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
