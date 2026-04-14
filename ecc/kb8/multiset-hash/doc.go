// Package multisethash implements a multiset hash over kb8.
//
// Messages are elements of the shared KoalaBear octic extension field. Each
// message is deterministically lifted to a point on kb8, and multiset hashing is
// the group sum of those lifted points. This matches the additive multiset-hash
// pattern used in zkVM memory arguments.
package multisethash
