package json

// A byteReader implements a sliding window over an io.Reader.
type byteReader struct {
	data   []byte
	offset int
}

// release discards n bytes from the front of the window.
func (b *byteReader) release(n int) {
	b.offset += n
}

// window returns the current window.
// The window is invalidated by calls to release or extend.
func (b *byteReader) window(offset int) []byte {
	return b.data[b.offset+offset:]
}
