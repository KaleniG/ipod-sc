package logic

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/cabbagekobe/tunetag"
	"github.com/sunshineplan/imgconv"
)

func GetFilesToSkip(indir string, outdir string) []string {
	filesToSkipList := []string{}
	inFiles := listFiles(indir)

	for _, file := range inFiles {
		fileExt := filepath.Ext(file)

		foldersWithoutRoot := strings.Replace(file, indir, "", 1)
		foldersWithoutRootNoFileExt := strings.Replace(foldersWithoutRoot, fileExt, "", 1)
		foldersWithoutRootNamesArray := strings.Split(foldersWithoutRootNoFileExt, string(filepath.Separator))

		for _, dirName := range foldersWithoutRootNamesArray {
			dirName = sanitize(dirName)
		}

		switch fileExt {
		case ".m4a":
			sanitizedFileName := sanitize(strings.TrimSuffix(filepath.Base(file), filepath.Ext(file)))
			outDirExpected := outdir + strings.Join(foldersWithoutRootNamesArray, string(filepath.Separator)) + string(filepath.Separator) + sanitizedFileName + fileExt

			_, err := os.Stat(outDirExpected)
			if os.IsNotExist(err) {
				continue
			}

			// * GETTING COVER ART FROM M4A FILE
			fileMP4, err := tunetag.OpenMP4(outDirExpected)
			if err != nil {
				log.Print(err, " failed to get m4a metadata from [", outDirExpected, "]")
				errorLog.Print(err, " failed to get m4a metadata from [", outDirExpected, "]")
				continue
			}

			pictures := fileMP4.Tag.Pictures()
			if len(pictures) == 0 {
				log.Print("no cover art for [", outDirExpected, "]")
				errorLog.Print("no cover art for [", outDirExpected, "]")
				continue
			}

			cover := pictures[0]
			img, err := imgconv.Decode(bytes.NewReader(cover.Payload))
			if err != nil {
				log.Print(err, " failed to decode cover art for [", outDirExpected, "]")
				errorLog.Print(err, " failed to decode cover art for [", outDirExpected, "]")
				continue
			}

			// * CONTROLS
			resizeNeeded := img.Bounds().Size().X != 500
			_, err = os.Stat(filepath.Dir(outDirExpected) + "/cover.jpg")
			fileCoverExists := !os.IsNotExist(err)

			if !resizeNeeded && fileCoverExists {
				filesToSkipList = append(filesToSkipList, file)
				log.Print("added [" + file + "] to the list of skipped files since it is already in a valid state in the output folder")
				continue
			}

			continue
		case ".flac":
			outDirExpected := outdir + strings.Join(foldersWithoutRootNamesArray, string(filepath.Separator)) + fileExt

			_, err := os.Stat(outDirExpected)
			if os.IsNotExist(err) {
				continue
			}

			// * GETTING COVER ART FROM FLAC FILE
			fileFLAC, err := tunetag.OpenFLAC(outDirExpected)
			if err != nil {
				log.Print(err, " failed to get flac metadata from [", outDirExpected, "]")
				errorLog.Print(err, " failed to get flac metadata from [", outDirExpected, "]")
				continue
			}

			pictures := fileFLAC.Pictures()
			if len(pictures) == 0 {
				log.Print("no cover art for [", outDirExpected, "]")
				errorLog.Print("no cover art for [", outDirExpected, "]")
				continue
			}

			cover := pictures[0]
			img, err := imgconv.Decode(bytes.NewReader(cover.Data))
			if err != nil {
				log.Print(err, " failed to decode cover art for [", outDirExpected, "]")
				errorLog.Print(err, " failed to decode cover art for [", outDirExpected, "]")
				continue
			}

			if img.Bounds().Size().X == 500 {
				filesToSkipList = append(filesToSkipList, file)
				log.Print("added [" + file + "] to the list of skipped files since it is already in a valid state in the output folder")
				continue
			}

			continue
		case ".jpg", ".jpeg", ".png":
			continue
		default:
			log.Print("unsupported file type: ", file)
			continue
		}
	}

	return filesToSkipList
}

func sanitize(name string) string {
	runes := []rune(name)
	out := make([]rune, 0, len(runes))

	for i, r := range runes {
		if !strings.ContainsRune(`<>:"/\|?*`, r) {
			out = append(out, r)
			continue
		}

		var prev, next rune
		hasPrev := i > 0
		hasNext := i < len(runes)-1

		if hasPrev {
			prev = runes[i-1]
		}
		if hasNext {
			next = runes[i+1]
		}

		// Replace with '_' only if both neighbors exist and are not spaces.
		if hasPrev && hasNext && !unicode.IsSpace(prev) && !unicode.IsSpace(next) {
			out = append(out, '_')
		}
		// Otherwise, omit the illegal character.
	}

	result := string(out)

	// Remove leading/trailing spaces and dots.
	result = strings.Trim(result, " .")

	// Collapse consecutive spaces into a single space.
	result = strings.Join(strings.Fields(result), " ")

	return result
}
