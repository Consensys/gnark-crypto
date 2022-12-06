// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iop

import "errors"

var (
	ErrInconsistantSize           = errors.New("the sizes of the polynomial must be the same as the size of the domain")
	ErrNumberPolynomials          = errors.New("the number of polynomials in the denominator and the numerator must be the same")
	ErrSizeNotPowerOfTwo          = errors.New("the size of the polynomials must be a power of two")
	ErrInconsistantSizeDomain     = errors.New("the size of the domain must be consistant with the size of the polynomials")
	ErrIncorrectNumberOfVariables = errors.New("the number of variables is incorrect")
)
