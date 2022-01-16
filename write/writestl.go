package write

import (
	"fmt"
	"os"

	"github.com/angl3dprinting/angl3dgo-watermark-stl/read"
)

func WriteSTL(path string, flavor int, stlData read.STLData) error {
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

func writeBinarySTL(file *os.File, stlData read.STLData) error {
	var err error
	return err
}

func writeAsciiSTL(file *os.File, stlData read.STLData) error {
	var err error
	return err
}
