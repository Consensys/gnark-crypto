// Package multisethash implements the y-increment multiset hash over kb8.
//
// Messages are 16-bit values. Each message m is mapped by scanning k in
// [0, 256) and setting y = m*256 + k in the base subfield of Fp^8. The first
// resulting point (x, y) on kb8 is used as the image of the message, and
// multiset hashing is the additive group sum of those mapped points.
package multisethash
