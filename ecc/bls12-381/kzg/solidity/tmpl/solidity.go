package tmpl

const SolidityKzg = `// SPDX-License-Identifier: Apache-2.0

// Copyright 2023 Consensys Software Inc.
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

// Code generated by gnark DO NOT EDIT

pragma solidity ^0.8.20;

contract KzgVerifier {

  uint256 private constant R_MOD = 21888242871839275222246405745257275088548364400416034343698204186575808495617;
  uint256 private constant R_MOD_MINUS_ONE = 21888242871839275222246405745257275088548364400416034343698204186575808495616;
  
  {{ range $index, $element := .Vk.G2 }}
  uint256 private constant G2_SRS_{{ $index }}_X_0_MSB = {{ (fpstrMSB $element.X.A1) }};
  uint256 private constant G2_SRS_{{ $index }}_X_0_LSB = {{ (fpstrLSB $element.X.A1) }};
  uint256 private constant G2_SRS_{{ $index }}_X_1_MSB = {{ (fpstrMSB $element.X.A0) }};
  uint256 private constant G2_SRS_{{ $index }}_X_1_LSB = {{ (fpstrLSB $element.X.A0) }};
  uint256 private constant G2_SRS_{{ $index }}_Y_0_MSB = {{ (fpstrMSB $element.Y.A1) }};
  uint256 private constant G2_SRS_{{ $index }}_Y_0_LSB = {{ (fpstrLSB $element.Y.A1) }};
  uint256 private constant G2_SRS_{{ $index }}_Y_1_MSB = {{ (fpstrMSB $element.Y.A0) }};
  uint256 private constant G2_SRS_{{ $index }}_Y_1_LSB = {{ (fpstrLSB $element.Y.A0) }};
  {{ end }}
  uint256 private constant G1_SRS_X_MSB = {{ fpstrMSB .Vk.G1.X }};
  uint256 private constant G1_SRS_X_LSB = {{ fpstrLSB .Vk.G1.X }};
  uint256 private constant G1_SRS_Y_MSB = {{ fpstrMSB .Vk.G1.Y }};
  uint256 private constant G1_SRS_Y_LSB = {{ fpstrLSB .Vk.G1.Y }};
  // uint256 private constant G1_SRS_Y_NEG = {{ neg .Vk.G1.Y }};


  // -------- proofs layout
  uint256 private constant SINGLE_PROOF_POINT = 0x0;
  uint256 private constant SINGLE_PROOF_CLAIMED_VALUE = 0x20;
  uint256 private constant SINGLE_PROOF_COMMITMENT = 0x40;
  uint256 private constant SINGLE_PROOF_QUOTIENT = 0x80;

  // -------- errors
  uint256 private constant ERROR_STRING_ID = 0x08c379a000000000000000000000000000000000000000000000000000000000; // selector for function Error(string)

  // -------- useful constants
  uint256 private constant SIZE_SCALAR_FIELD = 0x20;
  uint256 private constant SIZE_BASE_FIELD = 0x20;
  uint256 private constant SIZE_POINT = 0x40;
  uint256 private constant GAMMA = 0x67616d6d61; // "gamma" in ascii

  // -------- precompiles
  uint8 private constant SHA256 = 0x2;
  uint8 private constant MOD_EXP = 0x5;
  
  uint8 private constant EC_ADD = 0x0b;
  uint8 private constant EC_MSM = 0x0c;
  uint8 private constant EC_PAIR = 0x0f;
    
event PrintUint256(uint256 test);

  /// @notice verifies a batched opening proof at a single point of a list of polynomials.
  /// @dev Reverts if the proof or the public inputs are malformed.
  /// @param batchOpeningProof serialised KZG proof -> [ point || digests || claimed values || proof ], proof is the quotient
  /// @return success true if the proof passes false otherwise
  function BatchVerifySinglePoint(bytes calldata batchOpeningProof) 

  public returns (bool) {

    bool res;

    uint256  test;

    assembly {

      let nb_digests := sub(batchOpeningProof.length, SIZE_BASE_FIELD)
      nb_digests := sub(batchOpeningProof.length, SIZE_POINT)
      nb_digests := div(nb_digests, add(SIZE_POINT, SIZE_SCALAR_FIELD))

      test := nb_digests
      let free_mem := mload(0x40)
      let gamma := derive_challenge(batchOpeningProof.offset, nb_digests, free_mem)
      // fold_proof(batchOpeningProof.offset, nb_digests, gamma, add(free_mem, add(SIZE_POINT, SIZE_SCALAR_FIELD)), free_mem)

      // let H := add(batchOpeningProof.offset, sub(batchOpeningProof.length, SIZE_POINT))
      
      // res := verify(free_mem, H, batchOpeningProof.offset, add(free_mem, add(SIZE_POINT, SIZE_BASE_FIELD)))
  
      /// @notice error returned when SHA256 failed
      function error_sha256() {
        let ptError := mload(0x40)
        mstore(ptError, ERROR_STRING_ID) // selector for function Error(string)
        mstore(add(ptError, 0x4), 0x20)
        mstore(add(ptError, 0x24), 0xc)
        mstore(add(ptError, 0x44), "error sha256")
        revert(ptError, 0x64)
      }

      /// Called when an operation on Bn254 fails
      /// @dev for instance when calling EcMul on a point not on Bn254.
      function error_math_op() {
        let ptError := mload(0x40)
        mstore(ptError, ERROR_STRING_ID) // selector for function Error(string)
        mstore(add(ptError, 0x4), 0x20)
        mstore(add(ptError, 0x24), 0x14)
        mstore(add(ptError, 0x44), "error math operation")
        revert(ptError, 0x64)
      }

      /// @dev dst &lt;- dst + [s]dst
      /// @param dst pointer storing the result
      /// @param pt calldata pointer to an EC point
      /// @param s scalar
      /// @param mPtr free memory
      // function point_acc_mul_calldata(dst, pt, s, mPtr) {
      //   mstore(mPtr, calldataload(pt))
      //   mstore(add(mPtr, 0x20), calldataload(add(pt, 0x20)))
      //   mstore(add(mPtr, 0x40), s)
      //   let l_success := staticcall(gas(), EC_MUL, mPtr, 0x60, mPtr, 0x40)
      //   if iszero(l_success) {
      //     error_math_op()
      //   }
      //   // TODO change offset for bls12
      //   mstore(add(mPtr, 0x40), mload(dst))
      //   mstore(add(mPtr, 0x60), mload(add(dst, 0x20)))
      //   l_success := staticcall(gas(), EC_ADD, mPtr, 0x80, dst, 0x40)
      //   if iszero(l_success) {
      //     error_math_op()
      //   }
      // }

      /// @dev dst &lt;- dst + [s]pt
      /// @param dst pointer storing the result
      /// @param pt pointer to an EC point
      /// @param s scalar
      /// @param mPtr free memory
      // function point_acc_mul(dst, pt, s, mPtr) {
      //   mstore(mPtr, mload(pt))
      //   mstore(add(mPtr, SIZE_BASE_FIELD), mload(add(pt, SIZE_BASE_FIELD)))
      //   mstore(add(mPtr, SIZE_POINT), s)
      //   let l_success := staticcall(gas(), EC_MUL, mPtr, add(SIZE_POINT, SIZE_SCALAR_FIELD), mPtr, SIZE_POINT)
      //   if iszero(l_success) {
      //     error_math_op()
      //   }
      //   mstore(add(mPtr, SIZE_POINT), mload(dst))
      //   mstore(add(mPtr, add(SIZE_POINT, SIZE_BASE_FIELD)), mload(add(dst, SIZE_BASE_FIELD)))
      //   l_success := staticcall(gas(), EC_ADD, mPtr, add(SIZE_POINT, SIZE_POINT), dst, SIZE_POINT)
      //   if iszero(l_success) {
      //     error_math_op()
      //   }
      // }

      /// @dev dst &lt;- [s]dst + pt
      /// @param dst pointer storing the result
      /// @param pt calldata pointer to an EC point
      /// @param s scalar
      /// @param mPtr free memory
      // function point_mul_add_calldata(dst, pt, s, mPtr) {

      //   mstore(mPtr, mload(dst))
      //   mstore(add(mPtr, SIZE_BASE_FIELD), mload(add(dst, SIZE_BASE_FIELD)))
      //   mstore(add(mPtr, SIZE_POINT), s)
      //   let l_success := staticcall(gas(), EC_MUL, mPtr, add(SIZE_POINT, SIZE_SCALAR_FIELD), mPtr, SIZE_POINT)
      //   if iszero(l_success) {
      //     error_math_op()
      //   }

      //   calldatacopy(add(mPtr, SIZE_POINT), pt, SIZE_POINT)
      //   l_success := staticcall(gas(), EC_ADD, mPtr, add(SIZE_POINT, SIZE_POINT), dst, SIZE_POINT)
      //   if iszero(l_success) {
      //     error_math_op()
      //   }
      // }

      /// @dev dst &lt;- pt + s*dst [R_MOD]
      /// @param dst pointer storing the result
      /// @param pt calldata pointer to a scalar
      /// @param s scalar
      /// @param mPtr free memory
      function scalar_mul_add_calldata(dst, pt, s) {
        let tmp := mulmod(mload(dst), s, R_MOD)
        tmp := addmod(calldataload(pt), tmp, R_MOD)
        mstore(dst, tmp)
      }

      /// @notice verifies a folded proof at a single point
      /// @param folded_digest_and_claimed_values pointer to the digests folded, and the claimed values folded
      /// @param quotient calldata pointer to the quotient of the batch opening proof
      /// @param point calldata pointer to the point at which the proofs are opened
      /// @param mPtr pointer to free memory
      // function verify(folded_digest_and_claimed_values, quotient, point, mPtr)->res_pairing {

      //   let _mPtr := add(mPtr, SIZE_POINT)

      //   // folded_digest + [z]quotient
      //   mstore(mPtr, mload(folded_digest_and_claimed_values))
      //   mstore(add(mPtr, SIZE_BASE_FIELD), mload(add(folded_digest_and_claimed_values, SIZE_BASE_FIELD)))
      //   point_acc_mul_calldata(mPtr, quotient, calldataload(point), _mPtr)
        
      //   // folded_digest + [z]quotient - [folded_claimed_values]G
      //   mstore(_mPtr, G1_SRS_X)
      //   mstore(add(_mPtr, SIZE_BASE_FIELD), G1_SRS_Y_NEG)
      //   let g1_ptr := _mPtr
      //   _mPtr := add(_mPtr, SIZE_POINT)
      //   point_acc_mul(mPtr, g1_ptr, mload(add(folded_digest_and_claimed_values, SIZE_POINT)), add(_mPtr, SIZE_POINT))
      //   let tmp := mload(add(mPtr, SIZE_BASE_FIELD))
      //   tmp := sub(P_MOD, tmp)

      //   // - [ folded_digest + [z]quotient - [folded_claimed_values]G ]
      //   mstore(add(mPtr, SIZE_BASE_FIELD), tmp)

      //   // check e(- [ f(\alpha) + [z]H(\alpha) - [f(z)]G ], G2).e(H(\alpha)G1, [\alpha]G2)==1
      //   mstore(add(mPtr, 0x40), G2_SRS_0_X_0)
      //   mstore(add(mPtr, 0x60), G2_SRS_0_X_1)
      //   mstore(add(mPtr, 0x80), G2_SRS_0_Y_0)
      //   mstore(add(mPtr, 0xa0), G2_SRS_0_Y_1)
      //   mstore(add(mPtr, 0xc0), calldataload(quotient))
      //   mstore(add(mPtr, 0xe0), calldataload(add(quotient, SIZE_BASE_FIELD)))
      //   mstore(add(mPtr, 0x100), G2_SRS_1_X_0)
      //   mstore(add(mPtr, 0x120), G2_SRS_1_X_1)
      //   mstore(add(mPtr, 0x140), G2_SRS_1_Y_0)
      //   mstore(add(mPtr, 0x160), G2_SRS_1_Y_1)
      //   let pairing_op := staticcall(gas(), EC_PAIR, mPtr, 0x180, mPtr, 0x20)
      //   if iszero(pairing_op) {
      //     error_math_op()
      //   }
      //   res_pairing := mload(mPtr)
      // }

      /// @notice compute the challenge for kzg 
      /// @param proof calldata pointer to the proof, [ point || digests || claimed values || proof ]
      /// @param nbDigests number of proofs to fold
      /// @param _gamma challenge for folding the proofs
      /// @param mPtr free memory
      /// @param dst pointer where the result is stored. The result is [folded_digests, folded_claimed_values]
      // function fold_proof(proof, nbDigests, _gamma, mPtr, dst) {

      //   let current_claimed_value := add(SIZE_SCALAR_FIELD, mul(nbDigests, add(SIZE_BASE_FIELD, SIZE_POINT)))
      //   let current_digest := add(SIZE_SCALAR_FIELD, mul(nbDigests, SIZE_POINT))
        
      //   current_claimed_value := sub(current_claimed_value, SIZE_SCALAR_FIELD)
      //   current_digest := sub(current_digest, SIZE_POINT)
      
      //   calldatacopy(dst, add(proof, current_digest), SIZE_POINT)
      //   mstore(add(dst, SIZE_POINT), calldataload(add(proof, current_claimed_value)))

      //   for {let i := 0} lt(i, sub(nbDigests,1)) {i:=add(i,1)}{
          
      //     current_claimed_value := sub(current_claimed_value, SIZE_SCALAR_FIELD)
      //     scalar_mul_add_calldata(add(dst, SIZE_POINT), add(proof, current_claimed_value), _gamma)
          
      //     current_digest := sub(current_digest, SIZE_POINT)
      //     point_mul_add_calldata(dst, add(proof, current_digest), _gamma, mPtr)
      //   }

      // }

      /// @notice verify the folded proof
      /// @param proof calldata pointer to the proof
      /// @param nbDigests number of digests
      /// @param mPtr free memory
      function derive_challenge(proof, nbDigests, mPtr)->_gamma {

        let total_size_data

        // load gamma
        mstore(mPtr, GAMMA)
        let _mPtr := add(mPtr, 0x20)

        // load the point
        mstore(_mPtr, calldataload(proof))
        total_size_data := SIZE_SCALAR_FIELD

        // load the digests
        let size_data :=  mul(nbDigests, SIZE_POINT)
        calldatacopy(add(_mPtr, total_size_data), add(proof, total_size_data), size_data)
        total_size_data := add(total_size_data, size_data)

        // load the claimed values
        size_data := mul(nbDigests, SIZE_SCALAR_FIELD)
        calldatacopy(add(_mPtr, total_size_data), add(proof, total_size_data), size_data)
        total_size_data := add(total_size_data, size_data)

        // hash
        total_size_data := add(total_size_data, 5)
        let check_staticcall := staticcall(gas(), SHA256, add(mPtr,0x1b), total_size_data, mPtr, 0x20)
        if iszero(check_staticcall) {
          error_sha256()
        }

        // reduce
        _gamma := mod(mload(mPtr), R_MOD)

      }
    }
    emit PrintUint256(test);
    return res;
  
  }
}
`
