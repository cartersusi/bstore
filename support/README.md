## [Docker Config](./docker.conf.yml)
* Yaml file used in [deploy script](../deploy).
* Changes host, keys, public_base_path, private_base_path from [Example Config](./example.conf.yml).

## [Example Config](./example.conf.yml)
* Example of what a local development config looks like 

## [Example env](./example.keys.env)
* Example of valid `env.keys` variable names.

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