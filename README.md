# gurvy

[![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)  [![Go Report Card](https://goreportcard.com/badge/github.com/consensys/gurvy)](https://goreportcard.com/badge/github.com/consensys/gurvy) [![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/consensys/gurvy)](https://pkg.go.dev/mod/github.com/consensys/gurvy)


`gurvy` implements Elliptic Curve Cryptography (+Pairing) for BLS381, BLS377 and BN256. 

It is actively developed and maintained by the team (zkteam@consensys.net) behind:
* [`gnark`: a framework to execute (and verify) algorithms in zero-knowledge](https://github.com/consensys/gnark) 
* [`goff`: fast finite field arithmetic in Golang](https://github.com/consensys/goff)


## Warning
**`gurvy` has not been audited and is provided as-is, use at your own risk. In particular, `gurvy` makes no security guarantees such as constant time implementation or side-channel attack resistance.**

`gurvy` is optimized for 64bits architectures (x86 `amd64`) and tested on Unix (Linux / macOS).

## Curves supported

* BLS12-381 (Zcash)
* BN256 (Ethereum)
* BLS377 (ZEXE)
* BW6-761 (EC supporting pairing on BLS377 field of definition)


## Getting started

### Go version

`gurvy` is tested with the last 2 major releases of Go (1.14 and 1.15).

### Install `gurvy` 

```bash
go get github.com/consensys/gurvy
```

Note if that if you use go modules, in `go.mod` the module path is case sensitive (use `consensys` and not `ConsenSys`).

### Documentation
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/consensys/gurvy)](https://pkg.go.dev/mod/github.com/consensys/gurvy)

The APIs are consistent accross the curves. For example, [here is `bn256` godoc](https://pkg.go.dev/github.com/consensys/gurvy/bn256#pkg-overview).

## Benchmarks

Here are our measurements comparing `gurvy` (and [`goff` our finite field library](https://github.com/consensys/gurvy)) with [`mcl`](https://github.com/herumi/mcl).

These benchmarks ran on a AWS z1d.3xlarge instance, with hyperthreading disabled. 

|bn256|mcl(ns/op)|gurvy & goff (ns/op)|
| -------- | -------- | -------- |
|Fp::Add	|3.32|	3.44|
|Fp::Mul	|18.43|	16.1|
|Fp::Square	|18.64|	15.1|
|Fp::Inv	|690.55	|2080*|
|Fp::Pow	|6485|	7440*|
|G1::ScalarMul|	41394|	56900|
|G1::Add	|213|	224|
|G1::Double	|155|	178|
|G2::ScalarMul|	88423|	141000|
|G2::Add	|598|	871|
|G2::Double	|371|	386|
|Pairing	|478244	|606000|


----


|bls381|mcl(ns/op)|gurvy & goff (ns/op)|
| -------- | -------- | -------- |
|Fp::Add	|5.42|	4.6|
|Fp::Mul	|33.63|	29.3|
|Fp::Square	|33.86|	27|
|Fp::Inv	|1536	|4390*|
|Fp::Pow	|18039|	18300*|
|G1::ScalarMul|	76799|	91500|
|G1::Add	|424|	389|
|G1::Double	|308|	301|
|G2::ScalarMul|	159068|	273000|
|G2::Add	|1162|	1240|
|G2::Double	|727|	799|
|Pairing	|676513	|949000|

*note that some routines don't have assembly implementation in `goff` yet.


## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/consensys/gurvy/tags). 


## License

This project is licensed under the Apache 2 License - see the [LICENSE](LICENSE) file for details
