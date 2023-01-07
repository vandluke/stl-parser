# stl-parser

Basic STL parser for binary and ASCII flavors.

```cmd
go get -v github.com/vandluke/stl-parser/stl
```

Example Code

```Go
func main() {
    // Setup log file
    logPath := "stl-parser.log"
    os.Remove(logPath)
    file, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
    if err != nil {
        log.Fatalf("ERROR: %v", err)
    }
    defer file.Close()
    log.SetFlags(log.LstdFlags | log.Lshortfile)
    log.SetOutput(file)

    // STL file names
    stlBinaryPath := "in/test_binary.stl"
    stlAsciiPath := "in/test_ascii.stl"

    // Open STL files
    // Binary Flavor
    binarySTLFile, err := stl.OpenSTL(stlBinaryPath)
    if err != nil {
        log.Fatalf("ERROR: %v", err)
    }

    // ASCII Flavor
    asciiSTLFile, err := stl.OpenSTL(stlAsciiPath)
    if err != nil {
        log.Fatalf("ERROR: %v", err)
    }

    // Read STL Files
    // Binary Flavor
    err = binarySTLFile.ReadSTL()
    if err != nil {
        log.Fatalf("ERROR: %v", err)
    }

    // ASCII Flavor
    err = asciiSTLFile.ReadSTL()
    if err != nil {
        log.Fatalf("ERROR: %v", err)
    }

    // Write STL
    // Binary Flavor
    err = stl.WriteSTL("out/new_test_binary.stl", stl.STL_BINARY, &binarySTLFile.Data)
    if err != nil {
        log.Fatalf("ERROR: %v", err)
    }

    // ASCII Flavor
    err = stl.WriteSTL("out/new_test_ascii.stl", stl.STL_ASCII, &asciiSTLFile.Data)
    if err != nil {
        log.Fatalf("ERROR: %v", err)
    }

    // Close STL Files
    binarySTLFile.Close()
    asciiSTLFile.Close()
}
```

Structure for STL file from source code

```Go
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
```
