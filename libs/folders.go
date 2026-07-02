package libs

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
					log.Print(err, " ", folder)
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
		log.Fatal(err)
	}

	return folders
}
