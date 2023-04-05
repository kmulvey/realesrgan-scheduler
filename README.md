# realesrgan-scheduler

[![realesrgan-scheduler](https://github.com/kmulvey/realesrgan-scheduler/actions/workflows/release_build.yml/badge.svg)](https://github.com/kmulvey/realesrgan-scheduler/actions/workflows/release_build.yml) [![codecov](https://codecov.io/gh/kmulvey/realesrgan-scheduler/branch/main/graph/badge.svg?token=XpJ5kCJzsn)](https://codecov.io/gh/kmulvey/realesrgan-scheduler) [![Go Report Card](https://goreportcard.com/badge/github.com/kmulvey/realesrgan-scheduler)](https://goreportcard.com/report/github.com/kmulvey/realesrgan-scheduler) [![Go Reference](https://pkg.go.dev/badge/github.com/kmulvey/realesrgan-scheduler.svg)](https://pkg.go.dev/github.com/kmulvey/realesrgan-scheduler)


```
curl --request POST \
  --url http://localhost:3000/upsize/ \
  --header 'Authorization: Basic *****' \
  --header 'content-type: multipart/form-data' \
  --form sha512=950ebadec9ea32a1851b806df10be8d4403ea15adaf0bf24b3bda2ea0d467161b26e12cb4f6a68cd6613acf27e9cd29cf416b67c30bf0d04b0b15210425ee1d1 \
  --form image=@image
```
