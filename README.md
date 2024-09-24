# gnark-crypto

[![Twitter URL](https://img.shields.io/twitter/url/https/twitter.com/gnark_team.svg?style=social&label=Follow%20%40gnark_team)](https://twitter.com/gnark_team) [![License](https://img.shields.io/badge/license-Apache%202-blue)](LICENSE)  [![Go Report Card](https://goreportcard.com/badge/github.com/ConsenSys/gnark-crypto)](https://goreportcard.com/badge/github.com/ConsenSys/gnark-crypto) [![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/consensys/gnark-crypto)](https://pkg.go.dev/mod/github.com/consensys/gnark-crypto) [![DOI](https://zenodo.org/badge/DOI/10.5281/zenodo.5815453.svg)](https://doi.org/10.5281/zenodo.5815453)

`gnark-crypto` provides efficient cryptographic primitives, in Go:

* Elliptic curve cryptography & **Pairing** on:
  * [`bn254`] ([audit report](https://github.com/Consensys/gnark/blob/master/audits/2022-10%20-%20Kudelski%20-%20gnark-crypto.pdf))
  * [`bls12-381`] ([audit report](https://github.com/Consensys/gnark/blob/master/audits/2022-10%20-%20Kudelski%20-%20gnark-crypto.pdf))
  * [`bls24-317`]
  * [`bls12-377`] / [`bw6-761`]
  * [`bls24-315`] / [`bw6-633`]
  * Each of these curves has a [`twistededwards`] sub-package with its companion curve which allow efficient elliptic curve cryptography inside zkSNARK circuits.
* [`field/goff`] - Finite field arithmetic code generator (blazingly fast big.Int)
* [`fft`] - Fast Fourier Transform
* [`fri`] - FRI (multiplicative) commitment scheme
* [`fiatshamir`] - Fiat-Shamir transcript builder
* [`mimc`] - MiMC hash function using Miyaguchi-Preneel construction
* [`kzg`] - KZG commitment scheme
* [`permutation`] - Permutation proofs
* [`plookup`] - Plookup proofs
* [`eddsa`] - EdDSA signatures (on the companion [`twistededwards`] curves)

`gnark-crypto` is actively developed and maintained by the team (gnark@consensys.net | [HackMD](https://hackmd.io/@gnark)) behind:

* [`gnark`: a framework to execute (and verify) algorithms in zero-knowledge](https://github.com/consensys/gnark)

## Warning

**`gnark-crypto` is not fully audited and is provided as-is, use at your own risk. In particular, `gnark-crypto` makes no security guarantees such as constant time implementation or side-channel attack resistance.**

**To report a security bug, please refer to [`gnark` Security Policy](https://github.com/ConsenSys/gnark/blob/master/SECURITY.md).**

`gnark-crypto` packages are optimized for 64bits architectures (x86 `amd64`) and tested on Unix (Linux / macOS).

## Audits

See [list of audits for `gnark` and `gnark-crypto`](https://github.com/consensys/gnark?tab=readme-ov-file#audits)

## Getting started

### Go version

`gnark-crypto` is tested with the last 2 major releases of Go (currently 1.19 and 1.20).

### Install `gnark-crypto`

```bash
go get github.com/consensys/gnark-crypto
```

Note that if you use go modules, in `go.mod` the module path is case sensitive (use `consensys` and not `ConsenSys`).

### Development

Most (but not all) of the code is generated from the templates in `internal/generator`.

The generated code contains little to no interfaces and is strongly typed with a field (generated by the `gnark-crypto/field` package). The two main factors driving this design choice are:

1. Performance: `gnark-crypto` algorithms manipulate millions (if not billions) of field elements. Interface indirection at this level, plus garbage collection indexing takes a heavy toll on perf.
2. Need to derive (mostly) identical code for various moduli and curves, with consistent APIs. Generics introduce significant performance overhead and are not yet suited for high performance computing.

To regenerate the files, see `internal/generator/main.go`. Run:

```bash
go generate ./...
```

## Benchmarks

[Benchmarking pairing-friendly elliptic curves libraries](https://hackmd.io/@gnark/eccbench)

>The libraries are implemented in different languages and some use more assembly code than others. Besides the different algorithmic and software optimizations used across, it should be noted also that some libraries target constant-time implementation for some operations making it de facto slower. However, it can be clear that consensys/gnark-crypto is one of the fastest pairing-friendly elliptic curve libraries to be used in zkp projects with different curves.

## Citing

If you use `gnark-crypto` in your research a citation would be appreciated.
Please use the following BibTeX to cite the most recent release.

```bib
@software{gnark-crypto-v0.14.0,
  author       = {Gautam Botrel and
                  Thomas Piellard and
                  Youssef El Housni and
                  Arya Tabaie and
                  Gus Gutoski and
                  Ivo Kubjas},
  title        = {Consensys/gnark-crypto: v0.14.0},
  month        = sep,
  year         = 2024,
  publisher    = {Zenodo},
  version      = {v0.14.0},
  doi          = {10.5281/zenodo.5815453},
  url          = {https://doi.org/10.5281/zenodo.5815453}
}
```

## Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://github.com/consensys/gnark-crypto/tags).

## License

This project is licensed under the Apache 2 License - see the [LICENSE](LICENSE) file for details.

[`field/goff`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/field/goff
[`bn254`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254
[`bls12-381`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bls12-381
[`bls24-317`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bls24-317
[`bls12-377`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bls12-377
[`bls24-315`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bls24-315
[`bw6-761`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bw6-761
[`bw6-633`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bw6-633
[`twistededwards`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/twistededwards
[`eddsa`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa
[`fft`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/fr/fft
[`fri`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/fr/fri
[`mimc`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc
[`kzg`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/fr/kzg
[`plookup`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/fr/plookup
[`permutation`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/ecc/bn254/fr/permutation
[`fiatshamir`]: https://pkg.go.dev/github.com/consensys/gnark-crypto/fiat-shamir
