# go-ahrs

## Installation 
```sh
go get github.com/trakctum/go-ahrs
```

## Usage
See example in [ahrs_test.go](./ahrs_test.go#l80)

## Benchmark
Benchmark ran in my Dell `Intel® Core™ i3-5010U CPU @ 2.10GHz × 4` (2015 i3 laptop), with Arch Linux OS:
```
goos: linux
goarch: amd64
pkg: github.com/tracktum/go-ahrs
BenchmarkMadgwick-4     78186508               152 ns/op
BenchmarkMahony-4       100000000              117 ns/op
PASS
ok      github.com/tracktum/go-ahrs     25.318s
```
