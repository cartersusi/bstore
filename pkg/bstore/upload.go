package bstore

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cartersusi/bstore/pkg/fops"
	"github.com/cartersusi/bstore/pkg/stream"
	"github.com/gin-gonic/gin"
)

type StreamResponse struct {
	Hls    string `json:"hls_url"`
	Dash   string `json:"dash_url"`
	Poster string `json:"poster_url"`
}
type UploadRespone struct {
	Url     string         `json:"url"`
	Message string         `json:"message"`
	Stream  StreamResponse `json:"stream"`
}

func (bstore *ServerCfg) Upload(c *gin.Context) {
	log.Println("Valid Upload Request for", c.Request.URL.Path)
	validation := bstore.ValidateReq(c)
	if validation.Err != nil {
		HandleError(c, NewError(validation.HttpStatus, validation.Err.Error(), nil))
		return
	}

	fpath, err := fops.MkDirExt(validation.Fpath, validation.BasePath, bstore.Compress)
	if err != nil {
		HandleError(c, NewError(http.StatusBadRequest, "Error creating directory", err))
		return
	}
	log.Println("Creating file at", fpath)

	file, err := os.Create(fpath)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error creating file", err))
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	size, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		HandleError(c, NewError(http.StatusInternalServerError, "Error reading request body", err))
		return
	}
	if size > bstore.MaxFileSize {
		HandleError(c, NewError(http.StatusBadRequest, "File size exceeds maximum allowed size", nil))
		return
	}

	is_video := false
	stream_response := &StreamResponse{}
	stream_response.Hls = "UNAVAILABLE"
	stream_response.Dash = "UNAVAILABLE"
	stream_response.Poster = "UNAVAILABLE"
	v_fpath := strings.TrimSuffix(fpath, ".zst")
	if bstore.Streaming.Enabled {
		is_video = stream.CheckEXT(v_fpath)
	}

	if bstore.Compress {
		if is_video {
			log.Println("Video file detected, creating video stream at", v_fpath)
			err = fops.WriteNewFile(v_fpath, buf.Bytes(), false)
			if err != nil {
				HandleError(c, NewError(http.StatusInternalServerError, "Error writing data", err))
				return
			}
			err = stream.Make(v_fpath, bstore.Streaming.Codec, bstore.Compress, bstore.Encrypt, bstore.CompressionLevel)
			if err != nil {
				HandleError(c, NewError(http.StatusInternalServerError, "Error making video stream", err))
				return
			}
			_ = os.Remove(v_fpath)

			stream_response.Hls = stream.MakeUrl(c, v_fpath, stream.HLS)
			stream_response.Dash = stream.MakeUrl(c, v_fpath, stream.DASH)
			stream_response.Poster = stream.MakeUrl(c, v_fpath, stream.POSTER)
			log.Println("HLS Stream created at", stream_response.Hls)
			log.Println("DASH Stream created at", stream_response.Dash)
			log.Println("Poster created at", stream_response.Poster)
		}
		err = fops.Compress(&buf, file, bstore.CompressionLevel, bstore.Encrypt)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error writing compressed data", err))
			return
		}
	} else {
		err = fops.WriteFile(file, buf.Bytes(), bstore.Encrypt)
		if err != nil {
			HandleError(c, NewError(http.StatusInternalServerError, "Error writing data", err))
			return
		}
	}

	upload_response := &UploadRespone{
		Stream: *stream_response,
	}
	upload_response.Url = "UNAUTHORIZED"
	if bstore.GetAccess(c) != "private" {
		upload_response.Url = bstore.MakeUrl(c, validation.Fpath)
		upload_response.Message = "Public File Uploaded Successfully"
		log.Printf("Public file (%s) uploaded successfully to: %s\n", upload_response.Url, fpath)
	} else {
		upload_response.Message = "Private File Uploaded Successfully. No URL available"
		log.Printf("Private file (UNAUTHORIZED) uploaded successfully to: %s\n", fpath)
	}

	file.Sync()
	c.JSON(http.StatusOK, upload_response)
}
