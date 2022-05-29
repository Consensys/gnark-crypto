package bn254

func init() {
	encodeToG1Vector = hashTestVector{
		dst: []byte("QUUX-V01-CS02-with-BN254G1_XMD:SHA-256_SW_NU_"),
		cases: []hashTestCase{
			{
				msg: "", P: point{"0x184bb665c37ff561a89ec2122dd343f20e0f4cbcaec84e3c3052ea81d1834e192c426074b02ed3dca4e7676ce4ce48ba", "0x04407b8d35af4dacc809927071fc0405218f1401a6d15af775810e4e460064bcc9468beeba82fdc751be70476c888bf3"},
				u: "0x9d0a59611de15c3378929f8b2d75b88552ddb5e33703471050f21c47712cd1f",
				Q: point{"0x1ee13e2a308654bede154182bba785c2c76265c636f80a4763f2c914fc0c8e1a", "0xdc11d8fc9ad83ae4c0c400b4dc68ad682468c6530160257a8035a23bcddf6cd"}},
			/*{"abc", "0x009769f3ab59bfd551d53a5f846b9984c59b97d6842b20a2c565baa167945e3d026a3755b6345df8ec7e6acb6868ae6d", "0x1532c00cf61aa3d0ce3e5aa20c3b531a2abd2c770a790a2613818303c6b830ffc0ecf6c357af3317b9575c567f11cd2c"},
			{"abcdef0123456789", "0x1974dbb8e6b5d20b84df7e625e2fbfecb2cdb5f77d5eae5fb2955e5ce7313cae8364bc2fff520a6c25619739c6bdcb6a", "0x15f9897e11c6441eaa676de141c8d83c37aab8667173cbe1dfd6de74d11861b961dccebcd9d289ac633455dfcc7013a3"},
			{"q128_qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq", "0x0a7a047c4a8397b3446450642c2ac64d7239b61872c9ae7a59707a8f4f950f101e766afe58223b3bff3a19a7f754027c", "0x1383aebba1e4327ccff7cf9912bda0dbc77de048b71ef8c8a81111d71dc33c5e3aa6edee9cf6f5fe525d50cc50b77cc9"},
			{
				"a512_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				"0x0e7a16a975904f131682edbb03d9560d3e48214c9986bd50417a77108d13dc957500edf96462a3d01e62dc6cd468ef11",
				"0x0ae89e677711d05c30a48d6d75e76ca9fb70fe06c6dd6ff988683d89ccde29ac7d46c53bb97a59b1901abf1db66052db",
			},*/
		},
	}

	hashToG1Vector = hashTestVector{
		dst: []byte("QUUX-V01-CS02-with-BN254G1_XMD:SHA-256_SW_RO_"),
		cases: []hashTestCase{
			{
				msg: "", P: point{"0x184bb665c37ff561a89ec2122dd343f20e0f4cbcaec84e3c3052ea81d1834e192c426074b02ed3dca4e7676ce4ce48ba", "0x04407b8d35af4dacc809927071fc0405218f1401a6d15af775810e4e460064bcc9468beeba82fdc751be70476c888bf3"},
				u: "0x2a51502312e60a33a29160c6ddadccd57dec455f95414e0dc86f1d9adabd7ca6", u1: "0x19c958178283f6be6bb4106b7c40a0f1bb7bbc13173d0545ac39f4934e4edf6e",
				Q:  point{"0x15f794421b3ffe314d212edb7852328083d1dcf3e8128034e29633104b067472", "0x3c5ac0474e1567be66444bfd181dee24653c386cef32559705fc95428d25bb7"},
				Q1: point{"0x29fdc3ba7b17a66987a5d695419fdb94f65728c8860f2e1d0023cf60ce1006da", "0x1198977ada4e04f97dfbcc1c99624e5092b61ee6be93e792acb3f3cba8576aba"},
			},
		},
	}
}
