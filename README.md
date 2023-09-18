# Go Binio
[![Go Reference](https://pkg.go.dev/badge/github.com/KlemensWinter/binio.svg)](https://pkg.go.dev/github.com/KlemensWinter/binio)

A library for decoding binary data into structs.

> [!WARNING]  
> This lib is currently not ready for production use!



## Example
```go
type Data struct {
	MyString string `bin:"size=12"` // ASCII string with a length of 12
	A uint32
    B float32
    C float64
	_ [2]byte // skip 2 bytes
	D uint64

    E []int32 `bin:"size=3"`

    MyArrayLength int32
    MyArrayData []float64 `bin:"size=%MyArrayLength"`
}
```




