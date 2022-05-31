package bn254

func init() {
	encodeToG1Vector = encodeTestVector{
		dst: []byte("QUUX-V01-CS02-with-BN254G1_XMD:SHA-256_SVDW_NU_"),
		cases: []encodeTestCase{
			{
				msg: "", P: point{"0x1bb8810e2ceaf04786d4efd216fc2820ddd9363712efc736ada11049d8af5925", "0x1168559d94d0c7c3eb6fc17f1af2b54e3b885d68eab42f324902e3c4471e6e36"},
				Q: point{"0x1bb8810e2ceaf04786d4efd216fc2820ddd9363712efc736ada11049d8af5925", "0x1168559d94d0c7c3eb6fc17f1af2b54e3b885d68eab42f324902e3c4471e6e36"},
				u: "0xcb81538a98a2e3580076eed495256611813f6dae9e16d3d4f8de7af0e9833e1",
			}, {
				msg: "abc", P: point{"0xda4a96147df1f35b0f820bd35c6fac3b80e8e320de7c536b1e054667b22c332", "0x17c87a76e14cd7b5aaeb024133eb90ccb33497c005ec3c51c8e3609350e4c190"},
				Q: point{"0xda4a96147df1f35b0f820bd35c6fac3b80e8e320de7c536b1e054667b22c332", "0x17c87a76e14cd7b5aaeb024133eb90ccb33497c005ec3c51c8e3609350e4c190"},
				u: "0xba35e127276e9000b33011860904ddee28f1d48ddd3577e2a797ef4a5e62319",
			}, {
				msg: "abcdef0123456789", P: point{"0x2ff727cfaaadb3acab713fa22d91f5fddab3ed77948f3ef6233d7ea9b03f4da1", "0x23cdfc515ea7af332eefff5987daa54c684c4fd16cc39d09b1b8126b6bbb9d"},
				Q: point{"0x2ff727cfaaadb3acab713fa22d91f5fddab3ed77948f3ef6233d7ea9b03f4da1", "0x23cdfc515ea7af332eefff5987daa54c684c4fd16cc39d09b1b8126b6bbb9d"},
				u: "0x11852286660cd970e9d7f46f99c7cca2b75554245e91b9b19d537aa6147c28fc",
			}, {
				msg: "q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq", P: point{"0x11a2eaa8e3e89de056d1b3a288a7f733c8a1282efa41d28e71af065ab245df9b", "0x60f37c447ac29fd97b9bb83be98ddccf15e34831a9cdf5493b7fede0777ae06"},
				Q: point{"0x11a2eaa8e3e89de056d1b3a288a7f733c8a1282efa41d28e71af065ab245df9b", "0x60f37c447ac29fd97b9bb83be98ddccf15e34831a9cdf5493b7fede0777ae06"},
				u: "0x174d1c85d8a690a876cc1deba0166d30569fafdb49cb3ed28405bd1c5357a1cc",
			}, {
				msg: "a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", P: point{"0x27409dccc6ee4ce90e24744fda8d72c0bc64e79766f778da0c1c0ef1c186ea84", "0x15a24ccd9e32d613d0d1157f73dfd50c49b7cd075d45b71fd732a8ba07c7df87"},
				Q: point{"0x27409dccc6ee4ce90e24744fda8d72c0bc64e79766f778da0c1c0ef1c186ea84", "0x15a24ccd9e32d613d0d1157f73dfd50c49b7cd075d45b71fd732a8ba07c7df87"},
				u: "0x73b81432b4cf3a8a9076201500d1b94159539f052a6e0928db7f2df74bff672",
			},
		}}

	hashToG1Vector = hashTestVector{
		dst: []byte("QUUX-V01-CS02-with-BN254G1_XMD:SHA-256_SVDW_RO_"),
		cases: []hashTestCase{
			{
				msg: "", P: point{"0xa976ab906170db1f9638d376514dbf8c42aef256a54bbd48521f20749e59e86", "0x2925ead66b9e68bfc309b014398640ab55f6619ab59bc1fab2210ad4c4d53d5"},
				Q0: point{"0xe449b959abbd0e5ab4c873eaeb1ccd887f1d9ad6cd671fd72cb8d77fb651892", "0x29ff1e36867c60374695ee0c298fcbef2af16f8f97ed356fa75e61a797ebb265"},
				Q1: point{"0x19388d9112a306fba595c3a8c63daa8f04205ad9581f7cf105c63c442d7c6511", "0x182da356478aa7776d1de8377a18b41e933036d0b71ab03f17114e4e673ad6e4"},
				u0: "0x2f87b81d9d6ef05ad4d249737498cc27e1bd485dca804487844feb3c67c1a9b5", u1: "0x6de2d0d7c0d9c7a5a6c0b74675e7543f5b98186b5dbf831067449000b2b1f8e",
			}, {
				msg: "abc", P: point{"0x1250ed73eeaf132a6f0afc18eec66b45563aaa5505998802cc086a2dab68180a", "0x5ed828e0931606df6b82ddc86af34b5dcffa8136bca43e2dbef39938b36bdcc"},
				Q0: point{"0x1452c8cc24f8dedc25b24d89b87b64e25488191cecc78464fea84077dd156f8d", "0xfc8183f90d5f6d2c281f81f36facaa424c878dacb0e087468bff2ab09bb5006"},
				Q1: point{"0x4e8357c98524e6208ae2b771e370f0c449e839003988c2e4ce1eaf8d632559f", "0x4396ec43dd8ec8f2b4a705090b5892219759da30154c39490fc4d59d51bb817"},
				u0: "0x11945105b5e3d3b9392b5a2318409cbc28b7246aa47fa30da5739907737799a9", u1: "0x1255fc9ad5a6e0fb440916f091229bda611c41be2f2283c3d8f98c596be4c8c9",
			}, {
				msg: "abcdef0123456789", P: point{"0x196bca8afbc25fde75b96691bfc62bcb06c9b7b15bf34390540afc2df4b8c10", "0x115baf8ca3ac9340fafd7300deb8392357cb92e8e324a67bfc822c0c13342110"},
				Q0: point{"0x28d01790d2a1cc4832296774438acd46c2ce162d03099926478cf52319daba8d", "0x10227ab2707fd65fb45e87f0a48cfe3556f04113d27b1da9a7ae1709007355e1"},
				Q1: point{"0x7dc256c7aadac1b4e1d23b3b2bbb5e2ffd9c753b9073d8d952ead8f812ce1b3", "0xadb4de7b31bc375e6e36996945b23e6177fb8ded9671659472a9fb0a2b79f29"},
				u0: "0x2f7993a6b43a8dbb37060e790011a888157f456b895b925c3568690685f4983d", u1: "0x2677d0532b47a4cead2488845e7df7ebc16c0b8a2cd8a6b7f4ce99f51659794e",
			}, {
				msg: "q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq", P: point{"0x277769c98b482661e852e15c7c84fd4b463210a246a407024672d5fef17fa042", "0x13d9a11333bb7204668c6fa9bf6f1138a1c7307ceefc071fdd253d0c85f5dfa7"},
				Q0: point{"0x1c53b05f2fce15ba0b9100650c0fb46de1fb62f1d0968b69151151bd25dfefa4", "0x107cca77ec73e2b00138cd68bbeb3f4d29369dc184bbf11834871979531e81c6"},
				Q1: point{"0x214a4e6e97adda47558f80088460eabd71ed35bc8ceafb99a493dd6f4e2b3f0a", "0xfaaeb29cc23f9d09b187a99741613aed84443e7c35736258f57982d336d13bd"},
				u0: "0x2a50be15282ee276b76db1dab761f75401cdc8bd9fff81fcf4d428db16092a7b", u1: "0x23b41953676183c30aca54b5c8bd3ffe3535a6238c39f6b15487a5467d5d20eb",
			}, {
				msg: "a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", P: point{"0x27b2c37af3e558e3e72ce59443342624f22f39d6dd6679b432eceee5f0d87125", "0x497ae8461d62ef3a393ce6053fc326b37b265d18ea692a1e53922b5c48d47b6"},
				Q0: point{"0x2298ba379768da62495af6bb390ffca9156fde1dc167235b89c6dd008d2f2f3b", "0x660564cf6fce5cdea4780f5976dd0932559336fd072b4ddd83ec37f00fc7699"},
				Q1: point{"0x2811dea430f7a1f6c8c941ecdf0e1e725b8ad1801ad15e832654bd8f10b62f16", "0xb30bd85917e01d0f543a17deed654352f346f60ae92c4911885d6e3eddcd903"},
				u0: "0x48527470f534978bae262c0f3ba8380d7f560916af58af9ad7dcb6a4238e633", u1: "0x19a6d8be25702820b9b11eada2d42f425343889637a01ecd7672fbcf590d9ffe",
			},
		}}
}
