// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.30;

/**
 * @title Library to perform Poseidon2 hashing.
 * @author ConsenSys Software Inc.
 * @custom:security-contact security-report@linea.build
 */
library Poseidon2 {
  /// Thrown when the data is not purely in 32 byte chunks.
  error DataIsNotMod32();

  uint32 private constant R_MOD = 2130706433;
  uint256 private constant DATA_IS_NOT_MOD32_SELECTOR =
    0xc2cab26c00000000000000000000000000000000000000000000000000000000; // bytes4(keccak256("DataIsNotMod32()"))

  /**
   * @dev Round constants for Poseidon2 permutation.
   *
   * - RK_0_* .. RK_2_* : full rounds (both state elements affected)
   * - RK_3 .. RK_23   : partial rounds (first state element only)
   * - RK_24_* .. RK_26_* : final full rounds
   */
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
   * @notice Computes Poseidon2 hash over calldata message (must be 32-byte aligned).
   */
  function hash(bytes calldata _msg) external pure returns (bytes32 poseidon2Hash) {
    assembly {
      let len := _msg.length
      if and(len, 31) {
        error_size_data()
      }

      let q := shr(5, len)
      let ptrMsg := _msg.offset

      for {
        let i := 0
      } lt(i, q) {
        i := add(i, 1)
      } {
        let tmp := calldataload(ptrMsg)
        let a := poseidon2Hash
        let b := tmp
        a, b := permutation(a, b)
        poseidon2Hash := addRoundKeyUint256(tmp, b)
        ptrMsg := add(ptrMsg, 0x20)
      }

      /**
       * @dev Poseidon2 permutation over a 2-element state (ra, rb).
       *
       * The permutation consists of:
       *  1. Initial external MDS mixing
       *  2. 3 full rounds
       *  3. 21 partial rounds
       *  4. 3 final full rounds
       *
       * Each round follows the Poseidon2 specification for t = 2.
       *
       * @param a First state element (capacity/output lane)
       * @param b Second state element (rate/input lane)
       * @return ra Updated first state element
       * @return rb Updated second state element
       */
      function permutation(a, b) -> ra, rb {
        /*-------------------------------------------------------------*
        | Initial external MDS                                         |
        |                                                              |
        | Ensures early diffusion between a and b before non-linearity |
        *-------------------------------------------------------------*/
        ra, rb := matMulExternalInPlace(a, b)

        /*--------------------------------------------------------------*
        | FULL ROUNDS (3)                                               |
        |                                                               |
        | Each full round:                                              |
        |   1. Add round key to ALL limbs                               |
        |   2. Apply S-box (x ↦ x³) to ALL limbs                        |
        |   3. Apply external MDS                                       |
        *--------------------------------------------------------------*/
        ra := addRoundKeyUint256(ra, RK_0_0)
        rb := addRoundKeyUint256(rb, RK_0_1)
        ra := sboxUint256(ra)
        rb := sboxUint256(rb)
        ra, rb := matMulExternalInPlace(ra, rb)

        ra := addRoundKeyUint256(ra, RK_1_0)
        rb := addRoundKeyUint256(rb, RK_1_1)
        ra := sboxUint256(ra)
        rb := sboxUint256(rb)
        ra, rb := matMulExternalInPlace(ra, rb)

        ra := addRoundKeyUint256(ra, RK_2_0)
        rb := addRoundKeyUint256(rb, RK_2_1)
        ra := sboxUint256(ra)
        rb := sboxUint256(rb)
        ra, rb := matMulExternalInPlace(ra, rb)

        /*--------------------------------------------------------------*
        | PARTIAL ROUNDS (21)                                           |
        |                                                               |
        | Key Poseidon2 optimization:                                   |
        |   - Only FIRST limb is non-linear                             |
        |   - Only FIRST limb gets round key                            |
        |   - Internal MDS provides diffusion                           |
        |                                                               |
        | Each partial round:                                           |
        |   1. Add round key to first limb only                         |
        |   2. Apply S-box to first limb only                           |
        |   3. Compute sum of all limbs (both state elements)           |
        |   4. Apply internal MDS using that sum                        |
        |                                                               | 
        *--------------------------------------------------------------*/

        ra := addRoundKeyFirstEntry(ra, RK_3)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_4)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_5)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_6)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_7)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_8)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_9)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_10)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_11)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_12)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_13)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_14)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_15)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_16)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_17)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_18)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_19)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_20)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_21)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_22)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        ra := addRoundKeyFirstEntry(ra, RK_23)
        ra := sboxFirstEntry(ra)
        {
          let s := computeSum(ra, rb)
          ra := matMulInternalInPlaceFirstHalf(ra, s)
          rb := matMulInternalInPlaceSecondHalf(rb, s)
        }

        // full rounds (3)
        ra := addRoundKeyUint256(ra, RK_24_0)
        rb := addRoundKeyUint256(rb, RK_24_1)
        ra := sboxUint256(ra)
        rb := sboxUint256(rb)

        ra, rb := matMulExternalInPlace(ra, rb)

        ra := addRoundKeyUint256(ra, RK_25_0)
        rb := addRoundKeyUint256(rb, RK_25_1)
        ra := sboxUint256(ra)
        rb := sboxUint256(rb)

        ra, rb := matMulExternalInPlace(ra, rb)

        ra := addRoundKeyUint256(ra, RK_26_0)
        rb := addRoundKeyUint256(rb, RK_26_1)
        ra := sboxUint256(ra)
        rb := sboxUint256(rb)

        ra, rb := matMulExternalInPlace(ra, rb)
      }

      /**
       * @dev Computes the column-wise sum of two packed state elements.
       *
       * For each limb i:
       *   s += a[i] + b[i]   (mod R_MOD)
       *
       * @notice
       * This produces a scalar used by the INTERNAL MDS matrix,
       * which relies on a sum-based mixing strategy.
       *
       * This is a Poseidon2-specific optimization that avoids
       * full matrix multiplication during partial rounds.
       */
      function computeSum(a, b) -> s {
        let M := 0xFFFFFFFF
        s := addmod(s, addmod(shr(224, a), shr(224, b), R_MOD), R_MOD)
        s := addmod(s, addmod(and(shr(192, a), M), and(shr(192, b), M), R_MOD), R_MOD)
        s := addmod(s, addmod(and(shr(160, a), M), and(shr(160, b), M), R_MOD), R_MOD)
        s := addmod(s, addmod(and(shr(128, a), M), and(shr(128, b), M), R_MOD), R_MOD)
        s := addmod(s, addmod(and(shr(96, a), M), and(shr(96, b), M), R_MOD), R_MOD)
        s := addmod(s, addmod(and(shr(64, a), M), and(shr(64, b), M), R_MOD), R_MOD)
        s := addmod(s, addmod(and(shr(32, a), M), and(shr(32, b), M), R_MOD), R_MOD)
        s := addmod(s, addmod(and(a, M), and(b, M), R_MOD), R_MOD)
      }

      /**
       * @dev Applies INTERNAL MDS mixing to the first state element.
       *
       * @notice
       * This matrix is:
       * - Dense enough to ensure diffusion
       * - Structured to avoid full matrix multiplication
       * The full matrix is filled with 1s, and the diagonal is [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
       *
       *
       * Each limb is updated as:
       *   new_limb = sum + (constant × old_limb)
       */
      function matMulInternalInPlaceFirstHalf(a, sum) -> ma {
        let M := 0xFFFFFFFF
        ma := shl(224, addmod(sum, sub(R_MOD, mulmod(shr(224, a), 2, R_MOD)), R_MOD))
        ma := or(ma, shl(192, addmod(sum, and(shr(192, a), M), R_MOD)))
        {
          let x := and(shr(160, a), M)
          ma := or(ma, shl(160, addmod(sum, addmod(x, x, R_MOD), R_MOD)))
        }
        ma := or(ma, shl(128, addmod(sum, mulmod(and(shr(128, a), M), 1065353217, R_MOD), R_MOD)))
        ma := or(ma, shl(96, addmod(sum, mulmod(and(shr(96, a), M), 3, R_MOD), R_MOD)))
        ma := or(ma, shl(64, addmod(sum, mulmod(and(shr(64, a), M), 4, R_MOD), R_MOD)))
        ma := or(ma, shl(32, addmod(sum, mulmod(and(shr(32, a), M), 1065353216, R_MOD), R_MOD)))
        ma := or(ma, addmod(sum, sub(R_MOD, mulmod(and(a, M), 3, R_MOD)), R_MOD))
      }

      /**
       * @dev Applies INTERNAL MDS mixing to the second state element.
       *
       * Uses a different set of constants than the first half,
       * ensuring full diffusion across both state elements.
       * The full matrix is filled with 1s, and the diagonal is [-2, 1, 2, 1/2, 3, 4, -1/2, -3, -4, 1/2^8, 1/8, 1/2^24, -1/2^8, -1/8, -1/16, -1/2^24]
       */
      function matMulInternalInPlaceSecondHalf(b, sum) -> mb {
        let M := 0xFFFFFFFF
        mb := shl(224, addmod(sum, sub(R_MOD, mulmod(shr(224, b), 4, R_MOD)), R_MOD))
        mb := or(mb, shl(192, addmod(sum, mulmod(and(shr(192, b), M), 2122383361, R_MOD), R_MOD)))
        mb := or(mb, shl(160, addmod(sum, mulmod(and(shr(160, b), M), 1864368129, R_MOD), R_MOD)))
        mb := or(mb, shl(128, addmod(sum, mulmod(and(shr(128, b), M), 2130706306, R_MOD), R_MOD)))
        mb := or(mb, shl(96, addmod(sum, mulmod(and(shr(96, b), M), 8323072, R_MOD), R_MOD)))
        mb := or(mb, shl(64, addmod(sum, mulmod(and(shr(64, b), M), 266338304, R_MOD), R_MOD)))
        mb := or(mb, shl(32, addmod(sum, mulmod(and(shr(32, b), M), 133169152, R_MOD), R_MOD)))
        mb := or(mb, addmod(sum, mulmod(and(b, M), 127, R_MOD), R_MOD))
      }

      /**
       * @dev Applies the FULL (external) Poseidon2 MDS matrix.
       *
       * In the following, M4 = 
       * (2 3 1 1)
       * (1 2 3 1)
       * (1 1 2 3)
       * (3 1 1 2)
       * Steps:
       * 1. Apply 4×4 MDS to each half (matMulM4uint256, where the MDS is circ(2M4,M4,..,M4))
       * 2. Compute column sums across both state elements
       * 3. Redistribute column sums to all limbs
       *
       * This ensures complete diffusion across the entire state.
       */
      function matMulExternalInPlace(a, b) -> ra, rb {
        ra := matMulM4uint256(a)
        rb := matMulM4uint256(b)

        let t0, t1, t2, t3 := sumColumns(ra, rb)

        ra := matMulExternalInPlaceFirstHalf(ra, t0, t1, t2, t3)
        rb := matMulExternalInPlaceSecondHalf(rb, t0, t1, t2, t3)
      }

      /**
       * @dev Final redistribution step of EXTERNAL MDS for first state element.
       * 
       * Here:
       * [t0, t1, t2, t3] = M4*a_hi+M4*a_lo+M4*b_hi+M4*b_lo (computed with matMulM4uint256 followed by sumColumns, see matMulExternalInPlace)
       * a = diag(M4, M4)*[a_hi, a_lo]
       *
       * This function completes the computation of circ(2M4,M4,M4,M4)*[a, b] by adding t0, t1, t2, t3] to a_hi and a_lo
       *
       */
      function matMulExternalInPlaceFirstHalf(a, t0, t1, t2, t3) -> ra {
        let M := 0xFFFFFFFF
        ra := shl(224, addmod(t0, shr(224, a), R_MOD))
        ra := or(ra, shl(192, addmod(t1, and(shr(192, a), M), R_MOD)))
        ra := or(ra, shl(160, addmod(t2, and(shr(160, a), M), R_MOD)))
        ra := or(ra, shl(128, addmod(t3, and(shr(128, a), M), R_MOD)))
        ra := or(ra, shl(96, addmod(t0, and(shr(96, a), M), R_MOD)))
        ra := or(ra, shl(64, addmod(t1, and(shr(64, a), M), R_MOD)))
        ra := or(ra, shl(32, addmod(t2, and(shr(32, a), M), R_MOD)))
        ra := or(ra, addmod(t3, and(a, M), R_MOD))
      }

      /**
       * @dev Final redistribution step of EXTERNAL MDS for second state element.
       *
       * Uses the same column sums but applies them to the second word.
       */
      function matMulExternalInPlaceSecondHalf(b, t0, t1, t2, t3) -> rb {
        let M := 0xFFFFFFFF
        rb := shl(224, addmod(t0, shr(224, b), R_MOD))
        rb := or(rb, shl(192, addmod(t1, and(shr(192, b), M), R_MOD)))
        rb := or(rb, shl(160, addmod(t2, and(shr(160, b), M), R_MOD)))
        rb := or(rb, shl(128, addmod(t3, and(shr(128, b), M), R_MOD)))
        rb := or(rb, shl(96, addmod(t0, and(shr(96, b), M), R_MOD)))
        rb := or(rb, shl(64, addmod(t1, and(shr(64, b), M), R_MOD)))
        rb := or(rb, shl(32, addmod(t2, and(shr(32, b), M), R_MOD)))
        rb := or(rb, addmod(t3, and(b, M), R_MOD))
      }

      /**
       * @dev Computes column sums for EXTERNAL MDS mixing.
       *
       * Groups limbs into columns:
       *   column 0: limbs 0 + 4
       *   column 1: limbs 1 + 5
       *   column 2: limbs 2 + 6
       *   column 3: limbs 3 + 7
       *
       * This prepares shared values reused across both state elements,
       * reducing duplicated computation.
       */
      function sumColumns(a, b) -> t0, t1, t2, t3 {
        let M := 0xFFFFFFFF
        t0 := addmod(
          addmod(shr(224, a), and(shr(96, a), M), R_MOD),
          addmod(shr(224, b), and(shr(96, b), M), R_MOD),
          R_MOD
        )
        t1 := addmod(
          addmod(and(shr(192, a), M), and(shr(64, a), M), R_MOD),
          addmod(and(shr(192, b), M), and(shr(64, b), M), R_MOD),
          R_MOD
        )
        t2 := addmod(
          addmod(and(shr(160, a), M), and(shr(32, a), M), R_MOD),
          addmod(and(shr(160, b), M), and(shr(32, b), M), R_MOD),
          R_MOD
        )
        t3 := addmod(
          addmod(and(shr(128, a), M), and(a, M), R_MOD),
          addmod(and(shr(128, b), M), and(b, M), R_MOD),
          R_MOD
        )
      }

      /**
       * @dev Applies the 4×4 MDS matrix twice to a packed uint256.
       *
       * Layout:
       *   uint256 = [v0 v1 v2 v3 | v4 v5 v6 v7]
       *
       * The function:
       * - Applies matMulM4 to the top 4 limbs
       * - Applies matMulM4 to the bottom 4 limbs
       *
       * This preserves Poseidon2’s structured MDS design while
       * keeping operations word-local and gas-efficient.
       */
      function matMulM4uint256(a) -> b {
        let M := 0xFFFFFFFF
        // Upper 4 limbs
        {
          let v0 := shr(224, a)
          let v1 := and(shr(192, a), M)
          let v2 := and(shr(160, a), M)
          let v3 := and(shr(128, a), M)
          let s0, s1, s2, s3 := matMulM4(v0, v1, v2, v3)
          b := or(shl(224, s0), shl(192, s1))
          b := or(b, shl(160, s2))
          b := or(b, shl(128, s3))
        }
        // Lower 4 limbs
        {
          let v0 := and(shr(96, a), M)
          let v1 := and(shr(64, a), M)
          let v2 := and(shr(32, a), M)
          let v3 := and(a, M)
          let s4, s5, s6, s7 := matMulM4(v0, v1, v2, v3)
          b := or(b, shl(96, s4))
          b := or(b, shl(64, s5))
          b := or(b, shl(32, s6))
          b := or(b, s7)
        }
      }

      /**
       * @dev Multiplies a 4-element vector by the fixed 4×4 Poseidon2 MDS sub-matrix.
       * The matrix is
       * (2 3 1 1)
       * (1 2 3 1)
       * (1 1 2 3)
       * (3 1 1 2)
       *
       * Input:
       *   (a, b, c, d) ∈ F^4
       *
       * Output:
       *   (u, v, w, x) = MDS4 × (a, b, c, d)
       *
       * @notice
       * This is a hand-optimized implementation that:
       * - Avoids explicit matrix constants
       * - Uses algebraic reuse to minimize mul/add operations
       *
       * This function is used as a building block for 8-limb MDS multiplication.
       */
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

      /**
       * @dev Adds a round key ONLY to the first limb (limb 0).
       *
       * This function is used during PARTIAL rounds, where:
       * - Only the first state element is non-linear
       * - Only the first limb receives a round key
       *
       * All other limbs remain unchanged.
       */
      function addRoundKeyFirstEntry(x, k) -> rx {
        // Extract first limb (bits 255..224)
        let a0 := shr(224, x)
        let t0 := addmod(k, a0, R_MOD)
        rx := sub(x, shl(224, a0))
        rx := add(rx, shl(224, t0))
      }

      /**
       * @dev Adds a full 256-bit round key to a state element.
       *
       * This function interprets both `x` and `k` as vectors of
       * eight 32-bit field elements packed into a single uint256.
       *
       * For each 32-bit limb:
       *   out[i] = (x[i] + k[i]) mod R_MOD
       *
       * This is used during FULL rounds, where round keys are
       * applied to *all* limbs of the state.
       *
       * Layout (big-endian):
       *   limb 0 : bits [255..224]
       *   limb 7 : bits [31..0]
       */
      function addRoundKeyUint256(x, k) -> rx {
        // Mask for extracting a single 32-bit limb
        let M := 0xFFFFFFFF
        // Limb 0 (highest 32 bits)
        rx := shl(224, addmod(shr(224, x), shr(224, k), R_MOD))
        // Limb 1
        rx := or(rx, shl(192, addmod(and(shr(192, x), M), and(shr(192, k), M), R_MOD)))
        // Limb 2
        rx := or(rx, shl(160, addmod(and(shr(160, x), M), and(shr(160, k), M), R_MOD)))
        // Limb 3
        rx := or(rx, shl(128, addmod(and(shr(128, x), M), and(shr(128, k), M), R_MOD)))
        // Limb 4
        rx := or(rx, shl(96, addmod(and(shr(96, x), M), and(shr(96, k), M), R_MOD)))
        // Limb 5
        rx := or(rx, shl(64, addmod(and(shr(64, x), M), and(shr(64, k), M), R_MOD)))
        // Limb 6
        rx := or(rx, shl(32, addmod(and(shr(32, x), M), and(shr(32, k), M), R_MOD)))
        // Limb 7 (lowest 32 bits)
        rx := or(rx, addmod(and(x, M), and(k, M), R_MOD))
      }

      /**
       * @dev Applies the Poseidon S-box (x ↦ x³ mod R_MOD)
       * to ALL eight 32-bit limbs of a packed uint256.
       *
       * Used during FULL rounds.
       */
      function sboxUint256(x) -> rx {
        let M := 0xFFFFFFFF
        {
          let t := shr(224, x)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := shl(224, t)
        }
        {
          let t := and(shr(192, x), M)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := or(rx, shl(192, t))
        }
        {
          let t := and(shr(160, x), M)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := or(rx, shl(160, t))
        }
        {
          let t := and(shr(128, x), M)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := or(rx, shl(128, t))
        }
        {
          let t := and(shr(96, x), M)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := or(rx, shl(96, t))
        }
        {
          let t := and(shr(64, x), M)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := or(rx, shl(64, t))
        }
        {
          let t := and(shr(32, x), M)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := or(rx, shl(32, t))
        }
        {
          let t := and(x, M)
          t := mulmod(t, mulmod(t, t, R_MOD), R_MOD)
          rx := or(rx, t)
        }
      }

      /**
       * @dev Applies the Poseidon S-box ONLY to the first limb.
       *
       * This is the key optimization of Poseidon2:
       * - Only one limb is non-linear in partial rounds
       * - Saves significant gas
       */
      function sboxFirstEntry(x) -> rx {
        // Extract first limb (bits 255..224)
        let a0 := shr(224, x)
        let t0 := mulmod(a0, mulmod(a0, a0, R_MOD), R_MOD)
        rx := sub(x, shl(224, a0))
        rx := add(rx, shl(224, t0))
      }

      function error_size_data() {
        let ptr := mload(0x40)
        mstore(ptr, DATA_IS_NOT_MOD32_SELECTOR)
        revert(ptr, 4)
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
      let w := 0

      for {
        let i := 0
      } lt(i, 0x8) {
        i := add(i, 0x1)
      } {
        let v := and(shr(mul(sub(0xF, i), 0x10), input), 0xFFFF)
        w := or(w, shl(mul(sub(0x7, i), 0x20), v))
      }

      mstore(data, w)
      w := 0

      for {
        let i := 0x8
      } lt(i, 0x10) {
        i := add(i, 0x1)
      } {
        let v := and(shr(mul(sub(0xF, i), 0x10), input), 0xFFFF)
        w := or(w, shl(mul(sub(0xF, i), 0x20), v))
      }

      mstore(add(data, 0x20), w)
      mstore(0x40, add(data, 0x40))
    }
  }
}