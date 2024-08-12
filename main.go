package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type FileInfo struct {
	Name         string
	Size         int64
	IsDir        bool
	Permissions  os.FileMode
	Modification time.Time
}

func scanDirectory(path string, includeSubdirs bool) ([]FileInfo, error) {
	var files []FileInfo

	// Read directory contents
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, entry := range entries {
		wg.Add(1)
		go func(path string, entry os.FileInfo) {
			defer wg.Done()

			filePath := filepath.Join(path, entry.Name())

			fileInfo := FileInfo{
				Name:         entry.Name(),
				IsDir:        entry.IsDir(),
				Permissions:  entry.Mode().Perm(),
				Modification: entry.ModTime(),
			}

			if !entry.IsDir() {
				fileInfo.Size = entry.Size()
			}

			mu.Lock()
			files = append(files, fileInfo)
			mu.Unlock()

			if entry.IsDir() && includeSubdirs {
				subdirFiles, err := scanDirectory(filePath, true)
				if err != nil {
					fmt.Println("Error scanning subdirectory:", err)
					return
				}
				mu.Lock()
				files = append(files, subdirFiles...)
				mu.Unlock()
			}
		}(path, entry)
	}

	wg.Wait()

	return files, nil
}

func main() {
	var currentPath string
	var includeSubdirs bool
	var sortBy int

	// Start with the current working directory
	currentPath, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	for {
		fmt.Println("Current directory:", currentPath)

		files, err := scanDirectory(currentPath, includeSubdirs)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if len(files) == 0 {
			fmt.Println("No files found in the specified directory.")
			return
		}

		// Sort files based on the selected attribute
		switch sortBy {
		case 1:
			sort.Slice(files, func(i, j int) bool {
				return files[i].Name < files[j].Name
			})
		case 2:
			sort.Slice(files, func(i, j int) bool {
				return files[i].Size < files[j].Size
			})
		case 3:
			sort.Slice(files, func(i, j int) bool {
				return files[i].Modification.Before(files[j].Modification)
			})
		case 4:
			sort.Slice(files, func(i, j int) bool {
				return files[i].Permissions < files[j].Permissions
			})
		}

		fmt.Println("Files:")
		for i, file := range files {
			fmt.Printf("%d. %s", i+1, file.Name)
			if file.IsDir {
				fmt.Println(" (Directory)")
			} else {
				fmt.Printf(" (Size: %d bytes, Permissions: %s, Modification: %s)\n", file.Size, file.Permissions, file.Modification)
			}
		}

		fmt.Println("\nOptions:")
		fmt.Println("1. Enter directory")
		fmt.Println("2. Toggle subdirectories (Currently:", includeSubdirs, ")")
		fmt.Println("3. Sort by Name")
		fmt.Println("4. Sort by Size")
		fmt.Println("5. Sort by Modification Time")
		fmt.Println("6. Sort by Permissions")
		fmt.Println("7. Exit")

		var choice int
		fmt.Print("Enter your choice: ")
		fmt.Scanln(&choice)

		switch choice {
		case 1:
			var dirName string
			fmt.Print("Enter directory name: ")
			fmt.Scanln(&dirName)
			newPath := filepath.Join(currentPath, dirName)
			_, err := os.Stat(newPath)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				currentPath = newPath
			}
		case 2:
			includeSubdirs = !includeSubdirs
		case 3, 4, 5, 6:
			sortBy = choice
		case 7:
			return
		default:
			fmt.Println("Invalid choice. Please try again.")
		}
	}
}
