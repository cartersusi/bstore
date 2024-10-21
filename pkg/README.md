# [Go Docs](https://pkg.go.dev/github.com/cartersusi/bstore/pkg)

## View all Source files

|pkg|file|
|-|-|
|bstore| [conf.go](./bstore/conf.go)|
|bstore| [conf.go](./bstore/conf.go)|
|bstore| [delete.go](./bstore/delete.go)|
|bstore| [download.go](./bstore/download.go)|
|bstore| [list.go](./bstore/list.go)|
|bstore| [serve.go](./bstore/serve.go)|
|bstore| [server.go](./bstore/server.go)|
|bstore| [stream.go](./bstore/stream.go)|
|bstore| [upload.go](./bstore/upload.go)|
|cmd| [cmd.go](./cmd/cmd.go)|
|fops| [enc.go](./fops/enc.go)|
|fops| [fops.go](./fops/fops.go)|
|fops| [zstd.go](./fops/zstd.go)|
|stream| [stream.go](./stream/stream.go)|
|stream| [video.go](./stream/video.go)|


## Support Encoders for Video Streaming
* `libx264` (H.264/AVC, standard)
* `libx265` (H.265/HEVC, standard)
* `libvpx-vp9` (VP9, standard)
* `libaom-av1` (AV1, standard)
* `libsvtav1` (SVT-AV1, optimized for Intel/NVIDIA)
* `h264_videotoolbox` (H.264, Apple hardware encoding)
* `hevc_videotoolbox` (H.265/HEVC, Apple hardware encoding)
* `h264_amf` (H.264, AMD hardware encoding)
* `hevc_amf` (H.265/HEVC, AMD hardware encoding)
* `h264_nvenc` (H.264, NVIDIA hardware encoding)
* `hevc_nvenc` (H.265/HEVC, NVIDIA hardware encoding)