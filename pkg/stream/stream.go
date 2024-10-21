package stream

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cartersusi/bstore/pkg/cmd"
	"github.com/cartersusi/bstore/pkg/fops"
)

type VideoEncoderRequest struct {
	InputPath   string
	Codec       string
	Bitrate     int
	Compress    bool
	Encrypt     bool
	CompressLvl int
}

func Make(vreq VideoEncoderRequest) error {
	dash := &VideoEncoder{
		InputFile: vreq.InputPath,
		Codec:     vreq.Codec,
		Bitrate:   formatBitrate(vreq.Bitrate),
	}
	dash.VideoBuilder(DASH)

	thumbnail_cmd := fmt.Sprintf("ffmpeg -i %s -vf \"select=eq(n\\,0)\" -frames:v 1 -update 1 %s", vreq.InputPath, filepath.Join(dash.OutputDir, "index.jpg"))
	_ = cmd.RunCMD_fs(thumbnail_cmd)

	hls := &VideoEncoder{
		InputFile: vreq.InputPath,
		Codec:     vreq.Codec,
		Bitrate:   formatBitrate(vreq.Bitrate),
	}
	hls.VideoBuilder(HLS)

	// TODO: Removed for concurrent task execution, needs testing
	//dash_err := cmd.RunCMD_fs(dash.Command)
	//hls_err := cmd.RunCMD_fs(hls.Command)

	var dash_err, hls_err error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		dash_err = cmd.RunCMD_fs(dash.Command)
		if dash_err != nil {
			log.Println("Error with DASH:", dash_err)
		}
	}()

	go func() {
		defer wg.Done()
		hls_err = cmd.RunCMD_fs(hls.Command)
		if hls_err != nil {
			log.Println("Error with HLS:", hls_err)
		}
	}()

	wg.Wait()

	// If there is an error, still run the cleanup
	carry_over_errormsg := ""
	if dash_err != nil {
		carry_over_errormsg += fmt.Sprintf("DASH: %s\n", dash_err)
		log.Println("Error with DASH")
	}
	if hls_err != nil {
		carry_over_errormsg += fmt.Sprintf("HLS: %s\n", hls_err)
		log.Println("Error with HLS")
	}

	if dash.OutputDir == hls.OutputDir {
		err := CleanUp(vreq.Compress, vreq.Encrypt, vreq.CompressLvl, dash.OutputDir)
		if err != nil {
			return errors.New(fmt.Sprintf("%sVideo file is stream compatible but not able to compress/encrypt. %s", carry_over_errormsg, err))
		}
	} else {
		return errors.New(fmt.Sprintf("%sVideo file is stream compatible but not able to compress/encrypt. Output directories do not match", carry_over_errormsg))
	}
	return nil
}

func CleanUp(compress, encrypt bool, compress_lvl int, output_dir string) error {
	if !compress && !encrypt { // 0,0
		log.Println("No compression or encryption needed")
		return nil
	}
	files, err := fops.ListDir(output_dir)
	if err != nil {
		return errors.New("Error listing files")
	}

	log.Println(fmt.Sprintf("Compression: %t, Encrypting: %t | %d files", compress, encrypt, len(files)))
	for _, f := range files {

		// TODO: TEMP, if this folder already exists with .zst files, it should have never been processed
		if strings.HasSuffix(f, ".zst") {
			continue
		}
		fpath := filepath.Join(output_dir, f)
		fdata, err := fops.ReadFile(fpath, false)
		if err != nil {
			return errors.New("Error reading file")
		}

		if compress {
			fpath += ".zst"
		}

		file, err := os.Create(fpath)
		if err != nil {
			return errors.New("Error creating file")
		}
		defer file.Close()

		if compress { // 1,0 & 1,1
			err = fops.CompressData(fdata, file, compress_lvl, encrypt)
			if err != nil {
				return errors.New("Error compressing data")
			}
			_ = os.Remove(strings.TrimSuffix(fpath, ".zst"))
		} else { // 0,1
			// check here again later
			_ = os.Remove(fpath)
			data_enc, err := fops.Encrypt(fdata)
			if err != nil {
				return err
			}
			_, err = file.Write(data_enc)
			if err != nil {
				return errors.New("Error writing data")
			}
		}
	}

	return nil
}
