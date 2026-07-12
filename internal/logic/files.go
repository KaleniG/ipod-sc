package logic

import (
	"bytes"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cabbagekobe/tunetag"
	"github.com/cabbagekobe/tunetag/flac"
	"github.com/sunshineplan/imgconv"
)

func ProcessFiles(dir string, skipList []string) (int, int, int) {
	files := listFiles(dir)

	processedSongsCount := 0
	validSongsCount := 0
	totalSongsCount := 0

	for _, file := range files {
		skip := false
		for _, skipFile := range skipList {
			if file == skipFile {
				skip = true
				break
			}
		}

		if skip {
			continue
		}

		fileExt := filepath.Ext(file)
		switch fileExt {
		case ".m4a":
			totalSongsCount++

			// * GETTING COVER ART FROM M4A FILE
			fileMP4, err := tunetag.OpenMP4(file)
			if err != nil {
				log.Print(err, " failed to get m4a metadata from [", file, "]")
				errorLog.Print(err, " failed to get m4a metadata from [", file, "]")
				continue
			}

			pictures := fileMP4.Tag.Pictures()
			if len(pictures) == 0 {
				log.Print("no cover art for [", file, "]")
				errorLog.Print("no cover art for [", file, "]")
				continue
			}

			cover := pictures[0]
			img, err := imgconv.Decode(bytes.NewReader(cover.Payload))
			if err != nil {
				log.Print(err, " failed to decode cover art for [", file, "]")
				errorLog.Print(err, " failed to decode cover art for [", file, "]")
				continue
			}

			// * DATA GATHERING
			resizeNeeded := img.Bounds().Size().X != 500
			_, err = os.Stat(filepath.Dir(file) + "/cover.jpg")
			fileCoverExists := !os.IsNotExist(err)
			parentFolderValidName := filepath.Base(filepath.Dir(file)) == sanitize(filepath.Base(filepath.Dir(file)))

			if !resizeNeeded && fileCoverExists && parentFolderValidName {
				validSongsCount++
				log.Print("skipped [" + file + "], the m4a is already in a valid state")
				continue
			}

			// * COVER ART RESIZING
			if resizeNeeded {
				img = imgconv.Resize(img, &imgconv.ResizeOption{Width: 500})
			}

			// * COVER ART EXTRACTION
			var buf bytes.Buffer
			err = imgconv.Write(&buf, img, &imgconv.FormatOption{
				Format: imgconv.JPEG,
			})
			if err != nil {
				log.Print(err, " failed to write cover art into a buffer for file [", file, "]")
				errorLog.Print(err, " failed to write cover art into a buffer for file [", file, "]")
				continue
			}
			jpegBytes := buf.Bytes()

			// * COVER ART REPLACEMENT
			if resizeNeeded {
				fileMP4.Tag.Remove("covr")
				fileMP4.Tag.AddCover(jpegBytes)
			}

			// * GETTING SAVE NAMINGS
			sanitizedFileName := sanitize(strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)))

			// * MAKING A FOLDER FOR THE FILE AND COVER ART
			newFolderPath := filepath.Join(filepath.Dir(file), sanitizedFileName)
			err = os.MkdirAll(newFolderPath, 0755)
			if err != nil {
				log.Print(err, " failed to create the path [", newFolderPath, "]")
				errorLog.Print(err, " failed to create the path [", newFolderPath, "]")
				continue
			}
			log.Print("created folder [" + newFolderPath + "]")

			// * SAVING AND MOVING THE FILE IN THE NEW FOLDER
			saveFilePath := filepath.Join(newFolderPath, sanitizedFileName+fileExt)
			err = fileMP4.WriteFile(file)
			if err != nil {
				log.Print(err, " failed to save the file [", file, "]")
				errorLog.Print(err, " failed to save the file [", file, "]")
				continue
			}
			err = os.Rename(file, saveFilePath)
			if err != nil {
				log.Print(err, " failed to move the file [", file, "] to [", saveFilePath, "]")
				errorLog.Print(err, " failed to move the file [", file, "] to [", saveFilePath, "]")
				continue
			}
			log.Print("saved m4a file [" + saveFilePath + "] with resized cover art and moved to folder [" + newFolderPath + "]")

			// * SAVING THE COVER ART IN THE NEW FOLDER
			saveCoverDir := filepath.Join(newFolderPath, "cover.jpg")
			err = os.WriteFile(saveCoverDir, jpegBytes, 0644)
			if err != nil {
				log.Print(err, " failed to save the cover art file for file [", file, "]")
				errorLog.Print(err, " failed to save the cover art file for file [", file, "]")
				continue
			}
			log.Print("saved cover art [" + saveCoverDir + "] in folder [" + newFolderPath + "]")

			validSongsCount++
			processedSongsCount++
			continue
		case ".flac":
			totalSongsCount++

			// * GETTING COVER ART FROM FLAC FILE
			fileFLAC, err := tunetag.OpenFLAC(file)
			if err != nil {
				log.Print(err, " failed to get flac metadata from [", file, "]")
				errorLog.Print(err, " failed to get flac metadata from [", file, "]")
				continue
			}

			pictures := fileFLAC.Pictures()
			if len(pictures) == 0 {
				log.Print("no cover art for [", file, "]")
				errorLog.Print("no cover art for [", file, "]")
				continue
			}

			cover := pictures[0]
			img, err := imgconv.Decode(bytes.NewReader(cover.Data))
			if err != nil {
				log.Print(err, " failed to decode cover art for [", file, "]")
				errorLog.Print(err, " failed to decode cover art for [", file, "]")
				continue
			}

			validCoverArt := img.Bounds().Size().X == 500
			validFileName := file == filepath.Join(
				filepath.Dir(file),
				sanitize(strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)))+fileExt,
			)

			if validCoverArt && validFileName {
				validSongsCount++
				log.Print("skipped [" + file + "], the flac is already in a valid state")
				continue
			}

			// * COVER ART RESIZING & REPLACEMENT
			if !validCoverArt {
				img = imgconv.Resize(img, &imgconv.ResizeOption{Width: 500})

				var buf bytes.Buffer
				err = imgconv.Write(&buf, img, &imgconv.FormatOption{
					Format: imgconv.JPEG,
				})
				if err != nil {
					log.Print(err, " failed to write cover art into a buffer for file [", file, "]")
					errorLog.Print(err, " failed to write cover art into a buffer for file [", file, "]")
					continue
				}

				jpegBytes := buf.Bytes()
				fileFLAC.RemovePictures()
				fileFLAC.AddPicture(&flac.Picture{
					PictureType: 3,
					MIME:        "image/jpeg",
					Data:        jpegBytes,
				})
				log.Print("saved flac file metadata [" + file + "] with resized cover art")
			}

			if !validFileName {
				// * GETTING SAVE NAMINGS
				sanitizedFileName := filepath.Join(
					filepath.Dir(file),
					sanitize(strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)))+fileExt,
				)
				// * SAVING THE FILE
				err = fileFLAC.WriteFile(file)
				if err != nil {
					log.Print(err, " failed to save the file [", file, "]")
					errorLog.Print(err, " failed to save the file [", file, "]")
					continue
				}
				err = os.Rename(file, sanitizedFileName)
				if err != nil {
					log.Print(err, " failed to rename file [", file, "] to [", sanitizedFileName, "]")
					errorLog.Print(err, " failed to rename file [", file, "] to [", sanitizedFileName, "]")
					continue
				}
				log.Print("saved flac file as [" + sanitizedFileName + "]")
			}

			validSongsCount++
			processedSongsCount++
			continue
		case ".jpg", ".jpeg", ".png":
			if strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)) == "cover" {
				log.Print("detected cover art file [" + file + "], skipping")
			}
			continue
		default:
			log.Print("unsupported file type: ", file)
			continue
		}
	}

	return processedSongsCount, validSongsCount, totalSongsCount
}

func listFiles(dir string) []string {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		errorLog.Print(err)
		log.Fatal(err)
	}

	return files
}
