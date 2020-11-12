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
// cd https://pdfs.semanticscholar.org/e305/a02d91f222de4fe62d4b5689d3b03c7db0c3.pdf, 3.1
func (p *g2Jac) ClearCofactor(a *g2Jac) *g2Jac {

	var xg, xxg, xxxg, res, t g2Jac
	xg.ScalarMultiplication(a, &xGen).Neg(&xg)
	xxg.ScalarMultiplication(&xg, &xGen).Neg(&xxg)
	xxxg.ScalarMultiplication(&xxg, &xGen).Neg(&xxxg)

	res.Set(a).
		Double(&res).
		Double(&res).
		SubAssign(&xg).
		SubAssign(&xxg).
		AddAssign(&xxxg)

	t.Set(a).
		Neg(&t).
		AddAssign(&xg).
		AddAssign(&xg).
		SubAssign(&xxg).
		psi(&t).
		AddAssign(a).
		SubAssign(&xg).
		SubAssign(&xxg).
		AddAssign(&xxxg).
		psi(&t)

	res.AddAssign(&t)
	p.Set(&res)

	return p
}

// ClearCofactor maps a point in E(Fp) to E(Fp)[r]
// cf https://eprint.iacr.org/2019/403.pdf, 5
func (p *G2) ClearCofactor(a *G2) *G2 {
	var _p g2Jac
	_p.FromAffine(a)
	_p.ClearCofactor(&_p)
	p.FromJacobian(&_p)
	return p
}
