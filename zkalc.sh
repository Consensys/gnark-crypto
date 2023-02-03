
zkalc_benchmarks () {
    pushd $1

    pushd fp

    go test -run none -bench BenchmarkElementAdd
    go test -run none -bench 'BenchmarkElementMul\b'
    go test -run none -bench BenchmarkElementInverse

    go test -v -run none -bench BenchmarkElementSquare
    go test -v -run none -bench BenchmarkElementSqrt

    popd

    go test -run none -bench 'BenchmarkG1JacAdd\b'
    go test -run none -bench 'BenchmarkG1JacScalarMultiplication\b'
    go test -run none -bench 'BenchmarkMultiExpG1\b'

    go test -run none -bench BenchmarkG1JacIsInSubGroup
    go test -run none -bench BenchmarkG1AffineCofactorClearing

    go test -run none -bench 'BenchmarkG2JacAdd\b'
    go test -run none -bench 'BenchmarkG2JacScalarMultiplication\b'
    go test -run none -bench 'BenchmarkMultiExpG2\b'

    go test -run none -bench BenchmarkG2JacIsInSubGroup
    go test -run none -bench BenchmarkG2AffineCofactorClearing

    if [ -d internal/fptower ]
    then
        pushd internal/fptower

        go test -run none -bench BenchmarkE12Add
        go test -run none -bench BenchmarkE12Mul
        go test -run none -bench BenchmarkE12Cyclosquare

        popd
    fi

    go test -run none -bench BenchmarkPairing
    go test -run none -bench BenchmarkMultiPair

    popd
}


pushd ecc
zkalc_benchmarks secp256k1
zkalc_benchmarks bn254
zkalc_benchmarks bls12-381
zkalc_benchmarks bls12-377
popd