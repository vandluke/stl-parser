package main

import (
	"log"
	"os"

	"github.com/angl3dprinting/angl3dgo-watermark-stl/read"
)

func main() {
	// Setup log file
	logPath := "angl3dgo-stl.log"
	os.Remove(logPath)
	file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	defer file.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(file)

	// File names
	stlBinaryPath := "test/test_stl/test_binary.stl"
	stlAsciiPath := "test/test_stl/test_ascii.stl"

	binarySTLFile, err := read.OpenSTL(stlBinaryPath)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	asciiSTLFile, err := read.OpenSTL(stlAsciiPath)
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	err = binarySTLFile.ReadSTL()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	log.Println(binarySTLFile.Data.NumberOfTriangles)
	log.Println(len(binarySTLFile.Data.Triangles))
	log.Println(binarySTLFile.Data.Triangles)

	err = asciiSTLFile.ReadSTL()
	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}
	log.Println(asciiSTLFile.Data)

	binarySTLFile.Close()
	asciiSTLFile.Close()
}
