<a name="v0.4.0"></a>
## [v0.4.0] - 2021-03-31

### Refactor
- gurvy -> gnark-crypto
- moved interop tests under github.com/consensys/gnark-tests
- bls381 -> bls12-381
- bls377 -> bls12-377
- bn256 -> bn254
- migrated MiMC and EdDSA from gnark into gnark-crypto
- migrated gnark/backend/fft into gnark-crypto
- migrated goff packages into ./field/...
- cleaning internal/generator pattern

### Ci
- testing with go 1.15, go 1.16 on Windows, MacOS, Linux (+arch=32bits)

### Docs
- added ecc/ecc.md and field/field.md

### Feat
- multiExp in full extended jacobian coordinates

### Fix
- handle case where numCPU < 4 in precomputeExpTable
- incorrect comment and size returned in twistededwards SetBytes fixes [#34](https://github.com/ConsenSys/gnark-crypto/issues/34)
- point.SetBytes can now be called concurently with same byte slice input



<a name="v0.3.8"></a>
## [v0.3.8] - 2021-02-01

### Bls377
- final exp hard part eprint 2020/875
- ML entirely on the twist (ABLR)

### Bls381
- final exp hard part eprint 2020/875
- ML entirely on the twist (ABLR)
- change G1 and G2 generators for interop

### Bn256
- inline lineEval() in MilleLoop
- ML entirely on the twist (ABLR)
- change G1 and G2 generators for interop

### Bw6
- add E6 and pairing tests
- correct comments in FinalExp
- fix bw6 pairing API to take slices of points and mutualize squares
- change G1 and G2 generators for interop

### Pull Requests
- Merge pull request [#29](https://github.com/ConsenSys/gnark-crypto/issues/29) from ConsenSys/youssef/bls12-finalExp
- Merge pull request [#27](https://github.com/ConsenSys/gnark-crypto/issues/27) from ConsenSys/experimental/pairing
- Merge pull request [#26](https://github.com/ConsenSys/gnark-crypto/issues/26) from ConsenSys/youssef/ML-ABLR
- Merge pull request [#25](https://github.com/ConsenSys/gnark-crypto/issues/25) from ConsenSys/csquare
- Merge pull request [#23](https://github.com/ConsenSys/gnark-crypto/issues/23) from ConsenSys/youssef/bw6-API-pairing

