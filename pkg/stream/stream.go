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
)

func Make(input_path string, codec string, compress bool, encrypt bool, compress_lvl int) error {
	dash := &VideoEncoder{
		InputFile: input_path,
		Codec:     codec,
	}
	dash.VideoBuilder(DASH)

	hls := &VideoEncoder{
		InputFile: input_path,
		Codec:     codec,
	}
	hls.VideoBuilder(HLS)

	cmd.RunCMD_fs(dash.Command)
	cmd.RunCMD_fs(hls.Command)

	if dash.OutputDir == hls.OutputDir {
		err := CleanUp(compress, encrypt, compress_lvl, dash.OutputDir)
		if err != nil {
			return errors.New(fmt.Sprintf("Video file is stream compatible but not able to compress/encrypt. %s", err))
		}
	} else {
		return errors.New("Video file is stream compatible but not able to compress/encrypt. Output directories do not match")
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

		// TEMP, if this folder already exists with .zst files, it should have never been processed
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
