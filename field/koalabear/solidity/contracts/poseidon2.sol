// SPDX-License-Identifier: Apache-2.0

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

pragma solidity ^0.8.30;

/**
 * @title Library to perform Poseidon2 hashing.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library Poseidon2 {
  /**
   * Thrown when the data is not purely in 32 byte chunks.
   */
  error DataIsNotMod32();

  // KoalaBear prime field modulus.
  uint32 private constant R_MOD = 2130706433; // 0x7f000001

  uint256 private constant DATA_IS_NOT_MOD32_SELECTOR =
    0xc2cab26c00000000000000000000000000000000000000000000000000000000; // bytes4(keccak256("DataIsNotMod32()"))

  /// @dev Keys for each round.
  uint256 private constant RK_0_0 = 52691802021506155758914962750280372212207119203515444126415105344946620971042;
  uint256 private constant RK_0_1 = 32207471970256316655474490955553459742787419335289228299095903266455798739660;
  uint256 private constant RK_1_0 = 22163791677048831312463448776400028385347383911100916908889018061663075177430;
  uint256 private constant RK_1_1 = 52176870515245694198691020358647906763460864198395806999480367588806463375956;
  uint256 private constant RK_2_0 = 20057732892593326318373784289292844513497509146673516636435948826170767804894;
  uint256 private constant RK_2_1 = 18178750968836063378915279115560606579387492034215281114134869585318875352069;
  uint256 private constant RK_3 = 271263440;
  uint256 private constant RK_4 = 648183298;
  uint256 private constant RK_5 = 1662653166;
  uint256 private constant RK_6 = 1984135584;
  uint256 private constant RK_7 = 594964655;
  uint256 private constant RK_8 = 108023522;
  uint256 private constant RK_9 = 849546096;
  uint256 private constant RK_10 = 1961993938;
  uint256 private constant RK_11 = 1546336947;
  uint256 private constant RK_12 = 1036613726;
  uint256 private constant RK_13 = 452088758;
  uint256 private constant RK_14 = 275827416;
  uint256 private constant RK_15 = 763236035;
  uint256 private constant RK_16 = 1068717067;
  uint256 private constant RK_17 = 1580958419;
  uint256 private constant RK_18 = 1376393748;
  uint256 private constant RK_19 = 892777736;
  uint256 private constant RK_20 = 1345121022;
  uint256 private constant RK_21 = 908739241;
  uint256 private constant RK_22 = 908871000;
  uint256 private constant RK_23 = 1053550888;
  uint256 private constant RK_24_0 = 7748602703960850726417234176553190502144764796409545305676706613632081374108;
  uint256 private constant RK_24_1 = 14236178139181542176197168604443439317755831764127965411495610808590768539130;
  uint256 private constant RK_25_0 = 4390545490380878999875851257118004037103299792992754784803412502652715488373;
  uint256 private constant RK_25_1 = 15895992674755709583772916119416271803214403026810424648839392700430632374893;
  uint256 private constant RK_26_0 = 37373517675827041221658956101645979913006475784844873469590649853964048342988;
  uint256 private constant RK_26_1 = 46010512812451809471058691124553676654818408969360806522307687423952321374687;

  /**
   * @notice Computes the hash of a message using a Merkle Damgard scheme, with Poseidon2 for compression.
   *
   * Tier-C gas golf (stack-safe) highlights:
   * - Keep the chaining value as 8 limbs in scratch memory across blocks; only pack once at the end.
   * - Inline S-box (x^3) everywhere it is hot (full rounds + partial rounds).
   * - Inline partial-round wrapper (ARK+S-box on state[0]) while keeping intMDS as a function to avoid stack-too-deep.
   * - Fuse extMDS: compute M4 and accumulate column sums without a second “read pass” for sums.
   *
   * @param _msg The bytes message or data to hash. Must be multiple of 32 bytes.
   * @return poseidon2Hash The Poseidon2 hash.
   */
  function hash(bytes calldata _msg) external pure returns (bytes32 poseidon2Hash) {
    assembly {
      // Require length % 32 == 0 (cheaper than mod)
      if and(_msg.length, 0x1f) {
        mstore(0x00, DATA_IS_NOT_MOD32_SELECTOR)
        revert(0x00, 0x04)
      }

      // Scratch state: 16 limbs as 16 words (32 bytes each) = 0x200 bytes.
      // Layout:
      //   st + 0x00 .. 0xE0   : state[0..7]   (first half, chaining)
      //   st + 0x100 .. 0x1E0 : state[8..15]  (second half, message / working)
      let st := mload(0x40)
      mstore(0x40, add(st, 0x200))

      // Initialize chaining value to 0 (8 limbs)
      for { let p := st } lt(p, add(st, 0x100)) { p := add(p, 0x20) } { mstore(p, 0) }

      let ptr := _msg.offset
      let end := add(ptr, _msg.length)

      for { } lt(ptr, end) { ptr := add(ptr, 0x20) } {
        // NOTE: do NOT name this variable `msg` to avoid shadowing Solidity's `msg` object.
        let blk := calldataload(ptr)

        // Store blk (packed) into second half as 8 uint32 limbs (words).
        {
          let p2 := add(st, 0x100)
          let shift := 224
          for { let i := 0 } lt(i, 8) { i := add(i, 1) } {
            mstore(p2, and(shr(shift, blk), 0xffffffff))
            p2 := add(p2, 0x20)
            shift := sub(shift, 32)
          }
        }

        // Permute in place on 16 limbs in scratch (chaining||blk)
        permuteInPlace(st)

        // Feed-forward (second half += blk limbs mod p) AND move output into first half for next block.
        // We do NOT write back to second half (it will be overwritten next iteration).
        {
          let pFirst := st
          let pSecond := add(st, 0x100)
          let shift := 224
          for { let i := 0 } lt(i, 8) { i := add(i, 1) } {
            let mi := and(shr(shift, blk), 0xffffffff)
            let bi := addmod(mload(pSecond), mi, R_MOD)
            mstore(pFirst, bi)

            pFirst := add(pFirst, 0x20)
            pSecond := add(pSecond, 0x20)
            shift := sub(shift, 32)
          }
        }
      }

      // Pack final chaining value (first half) into bytes32 output (8x uint32, MSW-first).
      poseidon2Hash := packFirstHalf(st)

      // ----------------- Utilities / permutation -----------------

      // Pack state[0..7] from scratch into uint256 (8x uint32 MSW-first).
      function packFirstHalf(st_) -> out {
        out := or(shl(224, mload(add(st_, 0x00))), shl(192, mload(add(st_, 0x20))))
        out := or(out, shl(160, mload(add(st_, 0x40))))
        out := or(out, shl(128, mload(add(st_, 0x60))))
        out := or(out, shl(96, mload(add(st_, 0x80))))
        out := or(out, shl(64, mload(add(st_, 0xa0))))
        out := or(out, shl(32, mload(add(st_, 0xc0))))
        out := or(out, mload(add(st_, 0xe0)))
      }

      // M4 multiplication on 4 limbs mod p:
      // (2 3 1 1)
      // (1 2 3 1)
      // (1 1 2 3)
      // (3 1 1 2)
      function matMulM4(a, b, c, d) -> u, v, w, x {
        let t01 := addmod(a, b, R_MOD)
        let t23 := addmod(c, d, R_MOD)
        let t0123 := addmod(t01, t23, R_MOD)
        let t01123 := addmod(t0123, b, R_MOD)
        let t01233 := addmod(t0123, d, R_MOD)

        x := addmod(addmod(a, a, R_MOD), t01233, R_MOD)
        v := addmod(addmod(c, c, R_MOD), t01123, R_MOD)
        u := addmod(t01, t01123, R_MOD)
        w := addmod(t23, t01233, R_MOD)
      }

      // External linear layer (matMulExternalInPlace), in place on the 16-limb state.
      // Fused: we accumulate t0..t3 while producing M4 outputs, saving a full read pass.
      function extMDSInPlace(st_) {
        let t0 := 0
        let t1 := 0
        let t2 := 0
        let t3 := 0

        // Block 0: indices 0..3 at base st_ + 0x00
        {
          let p := st_
          let a := mload(add(p, 0x00))
          let b := mload(add(p, 0x20))
          let c := mload(add(p, 0x40))
          let d := mload(add(p, 0x60))
          let u, v, w, x := matMulM4(a, b, c, d)
          mstore(add(p, 0x00), u)
          mstore(add(p, 0x20), v)
          mstore(add(p, 0x40), w)
          mstore(add(p, 0x60), x)
          t0 := add(t0, u)
          t1 := add(t1, v)
          t2 := add(t2, w)
          t3 := add(t3, x)
        }

        // Block 1: indices 4..7 at base st_ + 0x80
        {
          let p := add(st_, 0x80)
          let a := mload(add(p, 0x00))
          let b := mload(add(p, 0x20))
          let c := mload(add(p, 0x40))
          let d := mload(add(p, 0x60))
          let u, v, w, x := matMulM4(a, b, c, d)
          mstore(add(p, 0x00), u)
          mstore(add(p, 0x20), v)
          mstore(add(p, 0x40), w)
          mstore(add(p, 0x60), x)
          t0 := add(t0, u)
          t1 := add(t1, v)
          t2 := add(t2, w)
          t3 := add(t3, x)
        }

        // Block 2: indices 8..11 at base st_ + 0x100
        {
          let p := add(st_, 0x100)
          let a := mload(add(p, 0x00))
          let b := mload(add(p, 0x20))
          let c := mload(add(p, 0x40))
          let d := mload(add(p, 0x60))
          let u, v, w, x := matMulM4(a, b, c, d)
          mstore(add(p, 0x00), u)
          mstore(add(p, 0x20), v)
          mstore(add(p, 0x40), w)
          mstore(add(p, 0x60), x)
          t0 := add(t0, u)
          t1 := add(t1, v)
          t2 := add(t2, w)
          t3 := add(t3, x)
        }

        // Block 3: indices 12..15 at base st_ + 0x180
        {
          let p := add(st_, 0x180)
          let a := mload(add(p, 0x00))
          let b := mload(add(p, 0x20))
          let c := mload(add(p, 0x40))
          let d := mload(add(p, 0x60))
          let u, v, w, x := matMulM4(a, b, c, d)
          mstore(add(p, 0x00), u)
          mstore(add(p, 0x20), v)
          mstore(add(p, 0x40), w)
          mstore(add(p, 0x60), x)
          t0 := add(t0, u)
          t1 := add(t1, v)
          t2 := add(t2, w)
          t3 := add(t3, x)
        }

        // Reduce column sums once.
        t0 := mod(t0, R_MOD)
        t1 := mod(t1, R_MOD)
        t2 := mod(t2, R_MOD)
        t3 := mod(t3, R_MOD)

        // Add column sums back to each block element (4 blocks).
        for { let base := st_ } lt(base, add(st_, 0x200)) { base := add(base, 0x80) } {
          mstore(add(base, 0x00), addmod(mload(add(base, 0x00)), t0, R_MOD))
          mstore(add(base, 0x20), addmod(mload(add(base, 0x20)), t1, R_MOD))
          mstore(add(base, 0x40), addmod(mload(add(base, 0x40)), t2, R_MOD))
          mstore(add(base, 0x60), addmod(mload(add(base, 0x60)), t3, R_MOD))
        }
      }

      // Internal linear layer (matMulInternalInPlace) on the 16-limb state in scratch.
      function intMDSInPlace(st_) {
        // Sum 16 limbs; reduce once.
        let s := 0
        s := add(s, mload(add(st_, 0x00)))
        s := add(s, mload(add(st_, 0x20)))
        s := add(s, mload(add(st_, 0x40)))
        s := add(s, mload(add(st_, 0x60)))
        s := add(s, mload(add(st_, 0x80)))
        s := add(s, mload(add(st_, 0xa0)))
        s := add(s, mload(add(st_, 0xc0)))
        s := add(s, mload(add(st_, 0xe0)))
        s := add(s, mload(add(st_, 0x100)))
        s := add(s, mload(add(st_, 0x120)))
        s := add(s, mload(add(st_, 0x140)))
        s := add(s, mload(add(st_, 0x160)))
        s := add(s, mload(add(st_, 0x180)))
        s := add(s, mload(add(st_, 0x1a0)))
        s := add(s, mload(add(st_, 0x1c0)))
        s := add(s, mload(add(st_, 0x1e0)))
        s := mod(s, R_MOD)

        // Update first half (0..7): each output depends only on s and its own original limb, so in-place is safe.
        let x := mload(add(st_, 0x00))
        mstore(add(st_, 0x00), addmod(s, sub(R_MOD, mulmod(x, 2, R_MOD)), R_MOD))

        x := mload(add(st_, 0x20))
        mstore(add(st_, 0x20), addmod(s, x, R_MOD))

        x := mload(add(st_, 0x40))
        mstore(add(st_, 0x40), addmod(s, mulmod(x, 2, R_MOD), R_MOD))

        x := mload(add(st_, 0x60))
        mstore(add(st_, 0x60), addmod(s, mulmod(x, 1065353217, R_MOD), R_MOD))

        x := mload(add(st_, 0x80))
        mstore(add(st_, 0x80), addmod(s, mulmod(x, 3, R_MOD), R_MOD))

        x := mload(add(st_, 0xa0))
        mstore(add(st_, 0xa0), addmod(s, mulmod(x, 4, R_MOD), R_MOD))

        x := mload(add(st_, 0xc0))
        mstore(add(st_, 0xc0), addmod(s, mulmod(x, 1065353216, R_MOD), R_MOD))

        x := mload(add(st_, 0xe0))
        mstore(add(st_, 0xe0), addmod(s, sub(R_MOD, mulmod(x, 3, R_MOD)), R_MOD))

        // Update second half (8..15)
        x := mload(add(st_, 0x100))
        mstore(add(st_, 0x100), addmod(s, sub(R_MOD, mulmod(x, 4, R_MOD)), R_MOD))

        x := mload(add(st_, 0x120))
        mstore(add(st_, 0x120), addmod(s, mulmod(x, 2122383361, R_MOD), R_MOD))

        x := mload(add(st_, 0x140))
        mstore(add(st_, 0x140), addmod(s, mulmod(x, 1864368129, R_MOD), R_MOD))

        x := mload(add(st_, 0x160))
        mstore(add(st_, 0x160), addmod(s, mulmod(x, 2130706306, R_MOD), R_MOD))

        x := mload(add(st_, 0x180))
        mstore(add(st_, 0x180), addmod(s, mulmod(x, 8323072, R_MOD), R_MOD))

        x := mload(add(st_, 0x1a0))
        mstore(add(st_, 0x1a0), addmod(s, mulmod(x, 266338304, R_MOD), R_MOD))

        x := mload(add(st_, 0x1c0))
        mstore(add(st_, 0x1c0), addmod(s, mulmod(x, 133169152, R_MOD), R_MOD))

        x := mload(add(st_, 0x1e0))
        mstore(add(st_, 0x1e0), addmod(s, mulmod(x, 127, R_MOD), R_MOD))
      }

      // Full round: add round keys (packed 8xuint32 per half) + sbox on all 16 elements.
      // NOTE: S-box is inlined (x^3) here for gas.
      function fullRoundInPlace(st_, rkA, rkB) {
        // First half [0..7]
        {
          let p := st_
          let shift := 224
          for { let i := 0 } lt(i, 8) { i := add(i, 1) } {
            let k := and(shr(shift, rkA), 0xffffffff)
            let x := addmod(mload(p), k, R_MOD)
            let x2 := mulmod(x, x, R_MOD)
            x := mulmod(x2, x, R_MOD)
            mstore(p, x)

            p := add(p, 0x20)
            shift := sub(shift, 32)
          }
        }

        // Second half [8..15]
        {
          let p := add(st_, 0x100)
          let shift := 224
          for { let i := 0 } lt(i, 8) { i := add(i, 1) } {
            let k := and(shr(shift, rkB), 0xffffffff)
            let x := addmod(mload(p), k, R_MOD)
            let x2 := mulmod(x, x, R_MOD)
            x := mulmod(x2, x, R_MOD)
            mstore(p, x)

            p := add(p, 0x20)
            shift := sub(shift, 32)
          }
        }
      }

      // Poseidon2 permutation:
      // - initial ext MDS
      // - 3 full rounds (each followed by ext MDS)
      // - 21 partial rounds (ARK+S on state[0], followed by internal MDS)
      // - 3 full rounds (each followed by ext MDS)
      function permuteInPlace(st_) {
        // Initial external layer
        extMDSInPlace(st_)

        // First 3 full rounds
        fullRoundInPlace(st_, RK_0_0, RK_0_1)
        extMDSInPlace(st_)

        fullRoundInPlace(st_, RK_1_0, RK_1_1)
        extMDSInPlace(st_)

        fullRoundInPlace(st_, RK_2_0, RK_2_1)
        extMDSInPlace(st_)

        // 21 partial rounds: inline ARK+S-box on state[0] + internal MDS
        {
          let x := addmod(mload(st_), RK_3, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_4, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_5, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_6, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_7, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_8, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_9, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_10, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_11, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_12, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_13, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_14, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_15, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_16, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_17, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_18, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_19, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_20, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_21, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_22, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }
        {
          let x := addmod(mload(st_), RK_23, R_MOD)
          let x2 := mulmod(x, x, R_MOD)
          x := mulmod(x2, x, R_MOD)
          mstore(st_, x)
          intMDSInPlace(st_)
        }

        // Last 3 full rounds
        fullRoundInPlace(st_, RK_24_0, RK_24_1)
        extMDSInPlace(st_)

        fullRoundInPlace(st_, RK_25_0, RK_25_1)
        extMDSInPlace(st_)

        fullRoundInPlace(st_, RK_26_0, RK_26_1)
        extMDSInPlace(st_)
      }
    }
  }

  /**
   * @notice Pads a bytes32 input into a bytes array by splitting it into two 32-byte segments.
   * @dev Every two bytes are prepended with 2 zero bytes. E.g. 0xAAAABBBB -> 0x0000AAAA0000BBBB.
   * @param input The bytes32 input to be padded.
   * @return out A bytes array containing the two 32-byte segments.
   */
  function padBytes32(bytes32 input) external pure returns (bytes memory out) {
    assembly {
      out := mload(0x40)
      mstore(out, 0x40)

      let data := add(out, 0x20)

      // First 8 halfwords -> first 32 bytes
      {
        let w := 0
        let shIn := 240
        let shOut := 224
        for { let i := 0 } lt(i, 8) { i := add(i, 1) } {
          let v := and(shr(shIn, input), 0xFFFF)
          w := or(w, shl(shOut, v))
          shIn := sub(shIn, 16)
          shOut := sub(shOut, 32)
        }
        mstore(data, w)
      }

      // Last 8 halfwords -> second 32 bytes
      {
        let w := 0
        let shIn := 112
        let shOut := 224
        for { let i := 0 } lt(i, 8) { i := add(i, 1) } {
          let v := and(shr(shIn, input), 0xFFFF)
          w := or(w, shl(shOut, v))
          shIn := sub(shIn, 16)
          shOut := sub(shOut, 32)
        }
        mstore(add(data, 0x20), w)
      }

      mstore(0x40, add(data, 0x40))
    }
  }
}
