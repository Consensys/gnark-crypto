package config

var Configs []Config

func init() {

	// bls12-377
	addConfig(NewConfig(
		"22",
		"8065159656716812877374967518403273466521432693661810619979959746626482506078",
		"47",
		"../../ecc/bls12-377/fr/fft",
		"github.com/consensys/gnark-crypto/ecc/bls12-377/fr",
	))

	// bls12-381
	addConfig(NewConfig(
		"7",
		"10238227357739495823651030575849232062558860180284477541189508159991286009131",
		"32",
		"../../ecc/bls12-381/fr/fft",
		"github.com/consensys/gnark-crypto/ecc/bls12-381/fr",
	))

	// bn254
	addConfig(NewConfig(
		"5",
		"19103219067921713944291392827692070036145651957329286315305642004821462161904",
		"28",
		"../../ecc/bn254/fr/fft",
		"github.com/consensys/gnark-crypto/ecc/bn254/fr",
	))

	// bw6-761
	addConfig(NewConfig(
		"15",
		"32863578547254505029601261939868325669770508939375122462904745766352256812585773382134936404344547323199885654433",
		"46",
		"../../ecc/bw6-761/fr/fft",
		"github.com/consensys/gnark-crypto/ecc/bw6-761/fr",
	))

	// bw6-633
	addConfig(NewConfig(
		"13",
		"4991787701895089137426454739366935169846548798279261157172811661565882460884369603588700158257",
		"20",
		"../../ecc/bw6-633/fr/fft",
		"github.com/consensys/gnark-crypto/ecc/bw6-633/fr",
	))

	// bls24-315
	addConfig(NewConfig(
		"7",
		"1792993287828780812362846131493071959406149719416102105453370749552622525216",
		"22",
		"../../ecc/bls24-315/fr/fft",
		"github.com/consensys/gnark-crypto/ecc/bls24-315/fr",
	))

	// bls24-317
	addConfig(NewConfig(
		"7",
		"16532287748948254263922689505213135976137839535221842169193829039521719560631",
		"60",
		"../../ecc/bls24-315/fr/fft",
		"github.com/consensys/gnark-crypto/ecc/bls24-315/fr",
	))

}

func addConfig(c Config) {
	Configs = append(Configs, c)
}
