# zkAlc benchmarks

## BLS12-381

* Fp:
In `ecc/bls12-381/fp`
- Add: `go test -run none -bench BenchmarkElementAdd`
- Mul: `go test -run none -bench 'BenchmarkElementMul\b'`
- Inv: `go test -run none -bench BenchmarkElementInverse`

- Square: `go test -v -run none -bench BenchmarkElementSquare`
- Sqrt: `go test -v -run none -bench BenchmarkElementSqrt`

----
* G1:
In `ecc/bls12-381`
- Add(Jac+Jac): `go test -run none -bench 'BenchmarkG1JacAdd\b'`
- ScalarMul(Jac): `go test -run none -bench 'BenchmarkG1JacScalarMultiplication\b'`
- MSM(2^1--2^21): `go test -run none -bench 'BenchmarkMultiExpG1\b'`

- IsInSubGroup: `go test -run none -bench BenchmarkG1JacIsInSubGroup`
- CofactorClearing: `go test -run none -bench BenchmarkG1AffineCofactorClearing`

----
* G2:
In `ecc/bls12-381`
- Add(Jac+Jac): `go test -run none -bench 'BenchmarkG2JacAdd\b'`
- ScalarMul(Jac): `go test -run none -bench 'BenchmarkG2JacScalarMultiplication\b'`
- MSM(2^1--2^21): `go test -run none -bench 'BenchmarkMultiExpG2\b'`

- IsInSubGroup: `go test -run none -bench BenchmarkG2JacIsInSubGroup`
- CofactorClearing: `go test -run none -bench BenchmarkG2AffineCofactorClearing`

----
* Gt:
In `ecc/bls12-381/internal/fptower`
- Add: `go test -run none -bench BenchmarkE12Add`
- Mul: `go test -run none -bench BenchmarkE12Mul`
- Square: `go test -run none -bench BenchmarkE12Cyclosquare`

----
* Pairings:
In `ecc/bls12-381`
- 1 Pairing: `go test -run none -bench BenchmarkPairing`
- Multi-Pairing(2^4--2^10): `go test -run none -bench BenchmarkMultiPair`
