module github.com/consensys/gurvy

go 1.15

require (
	github.com/consensys/bavard v0.0.0
	github.com/ethereum/go-ethereum v1.9.25
	github.com/kilic/bls12-381 v0.0.0-20201226121925-69dacb279461
	github.com/leanovate/gopter v0.2.9
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/sys v0.0.0-20210326220804-49726bf1d181
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/consensys/bavard => ../bavard
