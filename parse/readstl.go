package parse

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	STL_BINARY = iota
	STL_ASCII
	BATCH_SIZE = 10000
)

type STLFile struct {
	Flavor int
	File   *os.File
	Data   STLData
}

type STLData struct {
	NumberOfTriangles int
	// m = 0 - {nx, ny, nz}    - Normal
	// m = 1 - {v1x, v1y, v1z} - Vertex1
	// m = 2 - {v2x, v2y, v2z} - Vertex2
	// m = 3 - {v3x, v3y, v3z} - Vertex3
	Triangles [][4][3]float32
}

func (stlFile *STLFile) Close() {
	stlFile.File.Close()
}

func OpenSTL(path string) (*STLFile, error) {
	// STL file
	stlFile := STLFile{}

	// Error
	var err error

	// Open STL File
	stlFile.File, err = os.Open(path)
	if err != nil {
		return &stlFile, err
	}

	// Read Header
	// Binary: UINT8[80] – Header (must not begin with "solid")
	// Ascii: solid <name>
	buf := make([]uint8, 5)
	_, err = stlFile.File.Read(buf)
	if err != nil {
		return &stlFile, err
	}

	// Get flavor of STL file
	if strings.EqualFold(string(buf), "solid") {
		stlFile.Flavor = STL_ASCII
	} else {
		stlFile.Flavor = STL_BINARY
	}

	// Reset
	stlFile.File.Seek(0, io.SeekStart)

	return &stlFile, err
}

func (stlFile *STLFile) ReadSTL() error {
	// Error
	var err error
	switch stlFile.Flavor {
	case STL_ASCII:
		stlFile.Data, err = readAsciiSTL(stlFile)
	case STL_BINARY:
		stlFile.Data, err = readBinarySTL(stlFile)
	default:
		err = fmt.Errorf("unknown stl file flavor \"%v\"", stlFile.Flavor)
	}
	return err
}

func readBinarySTL(stlFile *STLFile) (STLData, error) {
	// STL Data
	stlData := STLData{}

	// Check if proper flavor
	if stlFile.Flavor != STL_BINARY {
		return stlData, errors.New("unable to read stl file as the flavor of stl file is not binary")
	}

	buf, err := ioutil.ReadAll(stlFile.File)
	if err != nil {
		return stlData, err
	}

	var start, size, triSize int = 84, len(buf), 50
	if size > start {
		stlData.NumberOfTriangles = int(binary.LittleEndian.Uint32(buf[start-4 : start]))
		stlData.Triangles = make([][4][3]float32, stlData.NumberOfTriangles)
		batchNum := stlData.NumberOfTriangles / BATCH_SIZE
		if stlData.NumberOfTriangles%BATCH_SIZE != 0 {
			batchNum++
		}
		wg := new(sync.WaitGroup)
		wg.Add(batchNum)
		for i := 0; i < batchNum; i++ {
			lowerBound, upperBound := start+i*BATCH_SIZE*triSize, start+(i+1)*BATCH_SIZE*triSize
			if upperBound > size {
				if lowerBound >= size {
					break
				}
				upperBound = size
			}
			go func(i, lowerBound, upperBound int) {
				defer wg.Done()
				for j := lowerBound; j < upperBound; j += triSize {
					for k := 0; k < 4; k++ {
						for l := 0; l < 3; l++ {
							stlData.Triangles[(j-start)/triSize][k][l] = math.Float32frombits(binary.LittleEndian.Uint32(buf[j+12*k+4*l : j+12*k+4*(l+1)]))
						}
					}
				}
			}(i, lowerBound, upperBound)
		}
		wg.Wait()
	}
	return stlData, err
}

func readAsciiSTL(stlFile *STLFile) (STLData, error) {
	// STL Data
	stlData := STLData{}

	// Check if proper flavor
	if stlFile.Flavor != STL_ASCII {
		return stlData, errors.New("unable to read stl file as the flavor of stl file is not ascii")
	}

	buf, err := ioutil.ReadAll(stlFile.File)
	if err != nil {
		return stlData, err
	}

	// Set number of triangles read
	facets := strings.Split(string(buf), "endfacet")
	stlData.NumberOfTriangles = len(facets) - 1
	stlData.Triangles = make([][4][3]float32, stlData.NumberOfTriangles)
	batchNum := stlData.NumberOfTriangles / BATCH_SIZE
	if stlData.NumberOfTriangles%BATCH_SIZE != 0 {
		batchNum++
	}
	wg := new(sync.WaitGroup)
	wg.Add(batchNum)
	for i := 0; i < batchNum; i++ {
		lowerBound, upperBound := i*BATCH_SIZE, (i+1)*BATCH_SIZE
		if upperBound > stlData.NumberOfTriangles {
			if lowerBound >= stlData.NumberOfTriangles {
				break
			}
			upperBound = stlData.NumberOfTriangles
		}
		go func(i, lowerBound, upperBound int) {
			defer wg.Done()
			for j := lowerBound; j < upperBound; j++ {
				stlData.Triangles[j] = parseFacetAscii(facets[j])
			}
		}(i, lowerBound, upperBound)
	}

	wg.Wait()

	return stlData, err
}

func parseFacetAscii(facet string) [4][3]float32 {
	var p0, p1, max, m, n int = 0, -1, len(facet) - 1, 0, 0
	var arr [4][3]float32
	goto s0
s0:
	p1++
	if p1 > max {
		goto s2
	}
	switch facet[p1] {
	case 32:
		p0 = p1 + 1
		goto s1
	default:
		goto s0
	}
s1:
	p1++
	if p1 > max {
		goto s2
	}
	switch facet[p1] {
	case 32, 10:
		if v, err := strconv.ParseFloat(strings.TrimSpace(facet[p0:p1]), 32); err == nil {
			arr[m][n] = float32(v)
			n++
			if n >= 3 {
				n = 0
				m++
			}
			if m >= 4 {
				goto s2
			}
		}
		p1--
		goto s0
	default:
		goto s1
	}
s2:
	return arr
}
