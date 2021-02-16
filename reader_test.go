package json

import (
	"testing"
)

func BenchmarkCountWhitespace(b *testing.B) {
	for _, tc := range inputs {
		data := fixture(b, tc.path)
		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				br := byteReader{
					data: data,
				}
				got := countWhitespace(&br)
				if got != tc.whitespace {
					b.Fatalf("expected: %v, got: %v", tc.whitespace, got)
				}
			}
		})
	}
}

func countWhitespace(br *byteReader) int {
	n := 0
	w := br.window(0)
	for _, c := range w {
		if whitespace[c] {
			n++
		}
	}
	br.release(len(w))
	return n
}
