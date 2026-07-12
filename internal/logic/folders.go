package logic

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func ProcessFolders(dir string) {
	for {
		changed := false

		list := listFolders(dir)

		for _, folder := range list {
			folderRel := filepath.Base(folder)
			sanitized := sanitize(folderRel)
			newPath := filepath.Join(filepath.Dir(folder), sanitized)

			if folder != newPath {
				err := os.Rename(folder, newPath)
				if err != nil {
					log.Print(err, " failed to rename the folder [", folder, "]")
					errorLog.Print(err, " failed to rename the folder [", folder, "]")
					continue
				}

				log.Printf("sanitized folder [%s] -> [%s]", folder, newPath)
				changed = true
				break
			}
		}

		if !changed {
			break
		}
	}
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func listFolders(dir string) []string {
	var folders []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && path != dir {
			folders = append(folders, path)
		}

		return nil
	})
	if err != nil {

		errorLog.Print(err)
		log.Fatal(err)
	}

	return folders
}
