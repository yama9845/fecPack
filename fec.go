package fec

import (
    "encoding/binary"
    "fmt"
    "bytes"
)

const (
    MAGIC = 0x31454c46 // "FLE1"
    MB = 1024 * 1024   // Size of 1 MB in bytes
)

type fileData struct {
    magic    int32
    pad1     [4]byte
    size     int64
    name     [256]byte
    fileData []byte
}

type encData struct {
    size int
    fd   []*fileData
}

func combineFiles(edp *encData, fileNum int) ([]byte, error) {
    readSizes := make([]int, fileNum) // Array to track how much data has been read from each fdp
    totalSize := 0
    chunkSize := MB / fileNum // Size of data to be read at a time

    for _, fdp := range edp.fd {
        totalSize += len(fdp.fileData) // Include the size of the fdp structure itself
    }

    combinedData := make([]byte, 0, totalSize)

    for totalSize > 0 {
        for i, fdp := range edp.fd {
            // Calculate the number of bytes to read
            fdpTotalSize := len(fdp.fileData)
            bytesToRead := fdpTotalSize - readSizes[i]
            if bytesToRead > chunkSize {
                bytesToRead = chunkSize
            }
            if bytesToRead > 0 {
                // Append the chunk of data to the combinedData
                combinedData = append(combinedData, fdp.fileData[readSizes[i]:readSizes[i]+bytesToRead]...)

                // Update the number of bytes read from this fdp
                readSizes[i] += bytesToRead
                totalSize -= bytesToRead
            }
        }
    }

    return combinedData, nil
}

func createEncodingData(fileNum int, files [][]byte) (*encData, error) {
    edp := &encData{
        fd: make([]*fileData, fileNum),
    }

    // Encode each file's data and store in encData struct
    for i, file := range files {
        fileSz := int64(len(file))

        // Copy file data and metadata to encData struct
        fdp := &fileData{
            magic:    MAGIC,
            size:     fileSz,
            fileData: make([]byte, 0, len(file)+4+4+8+256+1), // Allocate space for metadata and data, size field is now 8 bytes
        }

        // Write metadata to the beginning of fileData
        buf := new(bytes.Buffer)
        binary.Write(buf, binary.LittleEndian, fdp.magic)
        binary.Write(buf, binary.LittleEndian, fdp.pad1)
        binary.Write(buf, binary.LittleEndian, fdp.size)  // This now writes an int64
        binary.Write(buf, binary.LittleEndian, fdp.name)
        fdp.fileData = append(fdp.fileData, buf.Bytes()...)

        // Append actual file data
        fdp.fileData = append(fdp.fileData, file...)

        edp.fd[i] = fdp
    }

    return edp, nil
}

func FileMerge(files [][]byte) ([]byte, error) {
    fileNum := len(files)

    if fileNum < 1 {
        fmt.Println("The num of files to be encoded is less than 1.")
        return nil, fmt.Errorf("The num of files to be encoded is less than 1.")
    }

    fmt.Println("<target files>")
    for i, file := range files {
        fmt.Printf("File%d: %v\n", i+1, file)
    }

    // Encode data to fragments (creating files for backup)
    edp, err := createEncodingData(fileNum, files)
    if err != nil {
        fmt.Printf("Failed to create encoding data: %v\n", err)
        return nil, err
    }

    combinedData, err := combineFiles(edp, fileNum)
    if err != nil {
        fmt.Printf("Failed to combine files: %v\n", err)
        return nil, err
    }
    return combinedData, nil
}
