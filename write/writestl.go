package write

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/angl3dprinting/angl3dgo-watermark-stl/read"
)

func WriteSTL(path string, flavor int, stlData *read.STLData) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	switch flavor {
	case read.STL_ASCII:
		return writeAsciiSTL(file, stlData)
	case read.STL_BINARY:
		return writeBinarySTL(file, stlData)
	default:
		return fmt.Errorf("unknown flavor \"%d\"", flavor)
	}
}

func writeBinarySTL(file *os.File, stlData *read.STLData) error {
	// Initilize with empty header
	buf := make([]uint8, 80)
	// Add number of triangles
	numberOfTriangles := make([]uint8, 4)
	binary.LittleEndian.PutUint32(numberOfTriangles, uint32(stlData.NumberOfTriangles))
	buf = append(buf, numberOfTriangles...)

	batchNum := stlData.NumberOfTriangles / read.BATCH_SIZE
	if stlData.NumberOfTriangles%read.BATCH_SIZE != 0 {
		batchNum++
	}
	batches := make([][]uint8, batchNum)
	wg := new(sync.WaitGroup)
	wg.Add(batchNum)
	for i := 0; i < batchNum; i++ {
		lowerBound, upperBound := i*read.BATCH_SIZE, (i+1)*read.BATCH_SIZE
		if upperBound > stlData.NumberOfTriangles {
			if lowerBound >= stlData.NumberOfTriangles {
				break
			}
			upperBound = stlData.NumberOfTriangles
		}
		go func(i, lowerBound, upperBound int) {
			defer wg.Done()
			for j := lowerBound; j < upperBound; j++ {
				// Triangle
				for k := 0; k < len(stlData.Triangles[j]); k++ {
					for l := 0; l < len(stlData.Triangles[j][k]); l++ {
						real32Value := make([]uint8, 4)
						binary.LittleEndian.PutUint32(real32Value, math.Float32bits(stlData.Triangles[j][k][l]))
						batches[i] = append(batches[i], real32Value...)
					}
				}
				// Standard attribute byte count set to 0
				var byteCount uint16
				uint16Value := make([]uint8, 2)
				binary.LittleEndian.PutUint16(uint16Value, byteCount)
				batches[i] = append(batches[i], uint16Value...)
			}
		}(i, lowerBound, upperBound)
	}
	wg.Wait()

	for i := 0; i < batchNum; i++ {
		buf = append(buf, batches[i]...)
	}

	_, err := file.Write(buf)
	if err != nil {
		return err
	}

	return err
}

func writeAsciiSTL(file *os.File, stlData *read.STLData) error {
	// Initilize with header
	_, err := file.WriteString("solid Exported from angl3dgo-watermark-stl\n")
	if err != nil {
		return err
	}

	batchNum := stlData.NumberOfTriangles / read.BATCH_SIZE
	if stlData.NumberOfTriangles%read.BATCH_SIZE != 0 {
		batchNum++
	}
	batches := make([]strings.Builder, batchNum)
	wg := new(sync.WaitGroup)
	wg.Add(batchNum)
	for i := 0; i < batchNum; i++ {
		lowerBound, upperBound := i*read.BATCH_SIZE, (i+1)*read.BATCH_SIZE
		if upperBound > stlData.NumberOfTriangles {
			if lowerBound >= stlData.NumberOfTriangles {
				break
			}
			upperBound = stlData.NumberOfTriangles
		}
		go func(i, lowerBound, upperBound int) {
			defer wg.Done()
			for j := lowerBound; j < upperBound; j++ {
				// Triangle
				for k := 0; k < len(stlData.Triangles[j]); k++ {
					var formattedVector string
					for l := 0; l < len(stlData.Triangles[j][k]); l++ {
						formattedVector += fmt.Sprintf(" %f", stlData.Triangles[j][k][l])
					}
					if k == 0 {
						_, err := batches[i].WriteString(fmt.Sprintf("facet normal%s\nouter loop\n", formattedVector))
						if err != nil {
							panic(err)
						}
					} else {
						_, err := batches[i].WriteString(fmt.Sprintf("vertex %s\n", formattedVector))
						if err != nil {
							panic(err)
						}
					}
				}
				_, err := batches[i].WriteString("endloop\nendfacet\n")
				if err != nil {
					panic(err)
				}
			}
		}(i, lowerBound, upperBound)
	}
	wg.Wait()

	for i := 0; i < batchNum; i++ {
		_, err = file.WriteString(batches[i].String())
		if err != nil {
			return err
		}
	}

	_, err = file.WriteString("endsolid Exported from angl3dgo-watermark-stl\n")
	if err != nil {
		return err
	}

	return err
}
