package json

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var inputs = []struct {
	path       string
	tokens     int // decoded tokens
	alltokens  int // raw tokens, includes : and ,
	whitespace int // number of whitespace chars
}{
	// from https://github.com/miloyip/nativejson-benchmark
	{"canada", 223236, 334373, 33},
	{"citm_catalog", 85035, 135990, 1227563},
	{"twitter", 29573, 55263, 167931},
	{"code", 217707, 396293, 3},

	// from https://raw.githubusercontent.com/mailru/easyjson/master/benchmark/example.json
	{"example", 710, 1297, 4246},

	// from https://github.com/ultrajson/ultrajson/blob/master/tests/sample.json
	{"sample", 5276, 8677, 518549},
}

func BenchmarkScanner(b *testing.B) {
	for _, tc := range inputs {
		data := fixture(b, tc.path)
		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sc := &Scanner{
					br: byteReader{
						data: data,
					},
				}
				n := 0
				for len(sc.Next()) > 0 {
					n++
				}
				if n != tc.alltokens {
					b.Fatalf("expected %v tokens, got %v", tc.alltokens, n)
				}
			}
		})
	}
}

func BenchmarkBufferSize(b *testing.B) {
	b.Skip()
	for _, tc := range inputs {
		data := fixture(b, tc.path)
		b.Run(tc.path, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sc := &Scanner{
					br: byteReader{
						data: data,
					},
				}
				for len(sc.Next()) > 0 {
				}
			}
		})
	}
}

// fixture returns a *bytes.Reader for the contents of path.
func fixture(tb testing.TB, path string) []byte {
	f, err := os.Open(filepath.Join("testdata", path+".json.gz"))
	check(tb, err)
	defer f.Close()
	gz, err := gzip.NewReader(f)
	check(tb, err)
	buf, err := ioutil.ReadAll(gz)
	check(tb, err)
	return buf
}

func check(tb testing.TB, err error) {
	if err != nil {
		tb.Helper()
		tb.Fatal(err)
	}
}
