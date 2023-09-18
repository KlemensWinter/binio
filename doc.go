/*
Read / Write .dat files

# Types

Signed- and Unsigned-Ints are stores as Little Endian values.

Floats are stored as IEEE7xxx bytes.

Static Arrays of size N are stored as a sequence of their encoded types.

Static Strings are stored as ASCII.

"Holey Arrays" are Arrays with empty values.

Dynamic Arrays / Strings are stored with a `count` before the data.
*/

// Supported Types
//   - dynarray
//   - dynstring
//   - holeyarray

// dynarray(size_type)
// dynstring(size_type)
// holeyArray(size, ptrs)

package binio
