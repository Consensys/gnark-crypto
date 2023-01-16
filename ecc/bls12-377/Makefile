.PHONY: test run again

test:
	go test -run=Edwards -short && go test -run=MultiExpG1 -short
	
run:
	GOOS=linux GOARCH=arm64 go test -ldflags="-s -w" -c .
	adb push bls12-377.test /data/local/tmp/
	adb shell "/data/local/tmp/bls12-377.test -test.run=NONE -test.bench=UnifiedMixed -test.count=2"

again:
	adb shell "/data/local/tmp/bls12-377.test -test.run=NONE -test.bench=UnifiedMixed -test.count=2"

msm:
	GOOS=linux GOARCH=arm64 go test -ldflags="-s -w" -c .
	adb push bls12-377.test /data/local/tmp/
	adb shell "/data/local/tmp/bls12-377.test -test.run=NONE -test.bench=BenchmarkMultiExpG1Reference -test.count=2"

buckets:
	GOOS=linux GOARCH=arm64 go test -ldflags="-s -w" -c .
	adb push bls12-377.test /data/local/tmp/
	adb shell "/data/local/tmp/bls12-377.test -test.run=NONE -test.bench='UnifiedMixed|BenchmarkG1JacExtAddMixed' -test.count=2"