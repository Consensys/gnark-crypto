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

package bw761

import "math/big"

// ClearCofactor maps a point in E(Fp) to E(Fp2-)[r]
// https://eprint.iacr.org/2020/351.pdf
func (p *g2Jac) ClearCofactor(a *g2Jac) *g2Jac {

	var points [4]g2Jac
	points[0].Set(a)
	points[1].ScalarMultiplication(a, &xGen)
	points[2].ScalarMultiplication(&points[1], &xGen)
	points[3].ScalarMultiplication(&points[2], &xGen)

	var scalars [7]big.Int
	scalars[0].SetInt64(103)
	scalars[1].SetInt64(83)
	scalars[2].SetInt64(143)
	scalars[3].SetInt64(27)

	scalars[4].SetInt64(7)
	scalars[5].SetInt64(117)
	scalars[6].SetInt64(109)

	var p1, p2, tmp g2Jac
	p1.ScalarMultiplication(&points[3], &scalars[0])
	tmp.ScalarMultiplication(&points[2], &scalars[1]).Neg(&tmp)
	p1.AddAssign(&tmp)
	tmp.ScalarMultiplication(&points[1], &scalars[2]).Neg(&tmp)
	p1.AddAssign(&tmp)
	tmp.ScalarMultiplication(&points[0], &scalars[3])
	p1.AddAssign(&tmp)

	p2.ScalarMultiplication(&points[2], &scalars[4])
	tmp.ScalarMultiplication(&points[1], &scalars[5]).Neg(&tmp)
	p2.AddAssign(&tmp)
	tmp.ScalarMultiplication(&points[0], &scalars[6]).Neg(&tmp)
	p2.AddAssign(&tmp)
	p2.phi(&p2).phi(&p2)

	p.Set(&p1).AddAssign(&p2)

	return p
}

// ClearCofactor maps a point in E(Fp) to E(Fp)[r]
// cf https://eprint.iacr.org/2019/403.pdf, 5
func (p *G2Affine) ClearCofactor(a *G2Affine) *G2Affine {
	var _p g2Jac
	_p.FromAffine(a)
	_p.ClearCofactor(&_p)
	p.FromJacobian(&_p)
	return p
}
