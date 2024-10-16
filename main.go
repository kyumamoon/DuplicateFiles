package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type fileHash struct {
	path           string
	hash           string
	duplicate      int
	duplicatePaths []string
	size           int64
}

func hashFiles(filePath string, ErrorLog *string) string {
	file, err := os.Open(filePath)
	if err != nil {
		createLog(ErrorLog, err.Error())
		return ""
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		createLog(ErrorLog, err.Error())
		return ""
	}

	checksum := hash.Sum(nil)
	return fmt.Sprintf("%x", checksum) // Convert checksum to hexadecimal string
}

func outputResults(hashTable []fileHash, count int, ErrorLog *string) {
	// Define the file path
	filePath := "results.txt"

	// Create or open the file
	file, err := os.Create(filePath)
	if err != nil {
		createLog(ErrorLog, fmt.Sprintf("Error creating file:"+err.Error()))
		return
	}
	defer file.Close()

	// Write number of size saved:
	var bytesSaved int64
	for _, v := range hashTable {
		if v.duplicate != 0 {
			bytesSaved += int64(v.duplicate) * v.size
		}
	}
	var mbSaved float64 = float64(bytesSaved) / 1048576
	_, err = file.WriteString(fmt.Sprintf("DUPLICATES: %v\nPOSSIBLE MB SAVED: %v\n\n", count, mbSaved))
	if err != nil {
		createLog(ErrorLog, fmt.Sprintf("Error writing to file:"+err.Error()))
		return
	}

	// Write Results
	var content string
	for k, v := range hashTable {
		content_array := ""
		for i := 0; i < len(v.duplicatePaths); i++ {
			content_array += fmt.Sprintf("[%v]: %v\n", i, v.duplicatePaths[i])
		}
		content += fmt.Sprintf("FILE %v:\nHASH: %v\nPATH: %v\nDUPLICATES: %v\nDUPLICATE PATHS:\n", k, v.hash, v.path, v.duplicate) + content_array + "\n"
	}

	// Write the results to the file
	_, err = file.WriteString(content)
	if err != nil {
		createLog(ErrorLog, fmt.Sprintf("Error writing to file:"+err.Error()))
		return
	}
	// fmt.Println("Content written to file successfully.")
}

func createLog(log *string, errs string) {
	*log += errs + "\n"
}

func createLogFile(log string) {
	// Define the file path
	filePath := "ErrorsLog.txt"

	// Create or open the file
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the results to the file
	_, err = file.WriteString(log)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func main() {
	var hashTable []fileHash
	var ErrorLog string
	fileCount := 0
	// Get all file paths.
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			createLog(&ErrorLog, err.Error())
			// Log the error and continue
			fileCount++
			fmt.Printf("\nPROCESSING FILE: %v\n || ERROR ENCOUNTERED WITH FILE", fileCount)
			return nil // Returning nil to continue the walk
		}
		if !info.IsDir() {
			// Hash File
			hashDigest := hashFiles(path, &ErrorLog)
			newFileHash := fileHash{path: path, size: info.Size(), hash: hashDigest}
			hashTable = append(hashTable, newFileHash)
			fileCount++
			fmt.Printf("\nPROCESSING FILE: %v\n", fileCount)
		}
		return nil
	})
	if err != nil {
		createLog(&ErrorLog, err.Error())
		log.Fatal(err)
	}

	// Hash all files
	//hashTable = hashFiles(hashTable, &ErrorLog)

	// Hash Map of Duplicate Indexes
	duplicateIndexes := make(map[int]int)

	// Find Duplicates in copy of hashTable
	for k, v := range hashTable {
		for i := 0; i < len(hashTable); i++ {
			// check if it is a duplicate file already scanned..
			_, exists := duplicateIndexes[k]
			if exists {
				// if it is a duplicate file that was already checked. skip
				hashTable[k].duplicatePaths = append(hashTable[k].duplicatePaths, "A DUPLICATE OF ANOTHER FILE")
				break
			} else {
				// check if it is scanning itself
				if k != i {
					_, exists := duplicateIndexes[i]
					if exists {
						// nothing
					} else {
						// if doesnt exist, compare
						// Compare hashes
						if v.hash == hashTable[i].hash {
							duplicateIndexes[i] = 0
							hashTable[k].duplicate++
							hashTable[k].duplicatePaths = append(hashTable[k].duplicatePaths, hashTable[i].path)
						}
					}
				}
			}
		}
	}

	var dupeCount int
	for _, v := range hashTable {
		dupeCount += v.duplicate
	}

	// Output Results
	outputResults(hashTable, dupeCount, &ErrorLog)
	createLogFile(ErrorLog)
}
