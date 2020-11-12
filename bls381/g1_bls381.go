// Copyright 2020 ConsenSys AG
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

package bls381

// ClearCofactor maps a point in E(Fp) to E(Fp)[r]
// cf https://eprint.iacr.org/2019/403.pdf, 5
func (p *g1Jac) ClearCofactor(a *g1Jac) *g1Jac {
	var res g1Jac
	res.ScalarMultiplication(a, &xGen).AddAssign(a)
	p.Set(&res)
	return p
}

// ClearCofactor maps a point in E(Fp) to E(Fp)[r]
// cf https://eprint.iacr.org/2019/403.pdf, 5
func (p *G1) ClearCofactor(a *G1) *G1 {
	var _p g1Jac
	_p.FromAffine(a)
	_p.ClearCofactor(&_p)
	p.FromJacobian(&_p)
	return p
}
