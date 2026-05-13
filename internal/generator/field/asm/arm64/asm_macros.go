// Copyright 2020-2025 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package arm64

import (
	"github.com/consensys/bavard/arm64"
)

func (f *FFArm64) Loop(n arm64.Register, fn func()) {
	lblLoop := f.NewLabel("loop")
	lblDone := f.NewLabel("done")

	// while n > 0, do:
	f.LABEL(lblLoop)
	f.CBZ(n, lblDone)
	f.SUB(1, n, n)

	fn()

	f.JMP(lblLoop)
	f.LABEL(lblDone)
}
