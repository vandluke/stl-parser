package test

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/angl3dprinting/angl3dgo-watermark-stl/read"
)

type TestSTLJson struct {
	NumberOfTriangles int             `json:"NumberOfTriangles"`
	Triangles         [][4][3]float32 `json:"Triangles"`
}

type TestSTLCase struct {
	ReturnedBinary read.STLData
	ReturnedAscii  read.STLData
	Expected       TestSTLJson
}

func TestReadSTL(t *testing.T) {
	testCases := make(map[string]TestSTLCase)
	err := filepath.Walk("json/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".json") {
			var jsonFile TestSTLJson
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			buf, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			err = json.Unmarshal(buf, &jsonFile)
			if err != nil {
				return err
			}
			base := filepath.Base(path)
			fileName := base[:len(base)-len(".json")]
			v := testCases[fileName]
			v.Expected = jsonFile
			testCases[fileName] = v
		}
		return err
	})
	if err != nil {
		panic(err)
	}

	err = filepath.Walk("binary/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".stl") {
			stlFile, err := read.OpenSTL(path)
			if err != nil {
				return err
			}
			defer stlFile.Close()
			stlFile.ReadSTL()
			base := filepath.Base(path)
			fileName := base[:len(base)-len("_binary.stl")]
			v := testCases[fileName]
			v.ReturnedBinary = stlFile.Data
			testCases[fileName] = v
		}
		return err
	})
	if err != nil {
		panic(err)
	}

	err = filepath.Walk("ascii/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".stl") {
			stlFile, err := read.OpenSTL(path)
			if err != nil {
				return err
			}
			defer stlFile.Close()
			stlFile.ReadSTL()
			base := filepath.Base(path)
			fileName := base[:len(base)-len("_ascii.stl")]
			v := testCases[fileName]
			v.ReturnedAscii = stlFile.Data
			testCases[fileName] = v
		}
		return err
	})
	if err != nil {
		panic(err)
	}

	epsilon := 1e-6
	axisLabels := []string{"X", "Y", "Z"}
	labels := []string{"Normal", "Vertex1", "Vertex2", "Vertex3"}
	for fileName, tc := range testCases {
		tc := tc
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			if tc.Expected.NumberOfTriangles != tc.ReturnedAscii.NumberOfTriangles {
				t.Errorf("(ASCII) Expected Number of Triangles: %d Returned Number of Triangles: %d", tc.Expected.NumberOfTriangles, tc.ReturnedAscii.NumberOfTriangles)
			}
			if tc.Expected.NumberOfTriangles != tc.ReturnedBinary.NumberOfTriangles {
				t.Errorf("(Binary) Expected Number of Triangles: %d Returned Number of Triangles: %d", tc.Expected.NumberOfTriangles, tc.ReturnedBinary.NumberOfTriangles)
			}
			for i := range tc.Expected.Triangles {
				for j := range tc.Expected.Triangles[i] {
					for k := range tc.Expected.Triangles[i][j] {
						expected := tc.Expected.Triangles[i][j][k]
						if i < len(tc.ReturnedAscii.Triangles) {
							returned := tc.ReturnedAscii.Triangles[i][j][k]
							difference := math.Abs(float64(expected - returned))
							if difference > epsilon {
								t.Errorf("(ASCII) Expected %s %s: %v, Returned %s %s: %v, Difference: %v < %v = false", labels[j], axisLabels[k], expected, labels[j], axisLabels[k], returned, difference, epsilon)
							}
						}
						if i < len(tc.ReturnedBinary.Triangles) {
							returned := tc.ReturnedBinary.Triangles[i][j][k]
							difference := math.Abs(float64(expected - returned))
							if difference > epsilon {
								t.Errorf("(Binary) Expected %s %s: %v, Returned %s %s: %v, Difference: %v < %v = false", labels[j], axisLabels[k], expected, labels[j], axisLabels[k], returned, difference, epsilon)
							}
						}
					}
				}
			}
		})
	}
}

func BenchmarkReadBinarySTL(b *testing.B) {
	benchmarks := make(map[string]*read.STLFile)
	err := filepath.Walk("binary/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".stl") {
			stlFile, err := read.OpenSTL(path)
			if err != nil {
				return err
			}
			base := filepath.Base(path)
			fileName := base[:len(base)-len("_binary.stl")]
			benchmarks[fileName] = stlFile
		}
		return err
	})
	if err != nil {
		panic(err)
	}

	for fileName, bm := range benchmarks {
		b.Run(fileName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bm.ReadSTL()
			}
		})
		bm.Close()
	}
}

func BenchmarkReadAsciiSTL(b *testing.B) {
	benchmarks := make(map[string]*read.STLFile)
	err := filepath.Walk("ascii/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".stl") {
			stlFile, err := read.OpenSTL(path)
			if err != nil {
				return err
			}
			base := filepath.Base(path)
			fileName := base[:len(base)-len("_ascii.stl")]
			benchmarks[fileName] = stlFile
		}
		return err
	})
	if err != nil {
		panic(err)
	}

	for fileName, bm := range benchmarks {
		b.Run(fileName, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bm.ReadSTL()
			}
		})
		bm.Close()
	}
}
