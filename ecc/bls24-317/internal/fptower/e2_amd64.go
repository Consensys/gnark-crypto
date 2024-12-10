// Copyright 2020 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fptower

//go:noescape
func addE2(res, x, y *E2)

//go:noescape
func subE2(res, x, y *E2)

//go:noescape
func doubleE2(res, x *E2)

//go:noescape
func negE2(res, x *E2)
