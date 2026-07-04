package main

import (
	"flag"
	"fmt"
	"ipod-sc/internal/logic"
	"log"
	"os"

	copy "github.com/otiai10/copy"
)

func main() {
	skipExistingFiles := flag.Bool("s", false, "Skip existing files")
	skipExistingFilesLong := flag.Bool("skip-existing", false, "Skip existing files (long flag)")

	flag.Usage = func() {
		fmt.Println("Invalid usage of ipod-sc, missing parameters")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  ipod-sc [options] <input_dir> <output_dir>")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  -s, --skip-existing   Skip existing files")
		fmt.Println("")
		fmt.Println("Arguments:")
		fmt.Println("  input_dir   Folder containing music files")
		fmt.Println("  output_dir  Folder where processed files will be written")
		fmt.Println("")
		fmt.Println("Example:")
		fmt.Println("  ipod-sc -s ./input ./output")
	}

	flag.Parse()

	skipFlag := *skipExistingFiles || *skipExistingFilesLong

	args := flag.Args()

	if len(args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	indir := args[0]
	outdir := args[1]

	if logic.DirExists(indir) {
		log.Print("input folder [" + indir + "] exists")
	} else {
		log.Panic("input folder [" + indir + "] does not exist")
	}

	if logic.DirExists(outdir) {
		log.Print("output folder [" + outdir + "] exists")
	} else {
		log.Panic("output folder [" + outdir + "] does not exist")
	}

	filesToSkip := []string{}

	if skipFlag {
		filesToSkip = logic.GetFilesToSkip(indir, outdir)
	}

	tempDirName, err := os.MkdirTemp("", "ipod-sc-*")
	if err != nil {
		log.Panic("failed to create a temporary folder, " + err.Error())
	}
	log.Print("temporary folder for copy created")
	defer os.RemoveAll(tempDirName)

	log.Print("file copying to temporary folder started")
	if err := copy.Copy(indir, tempDirName); err != nil {
		log.Panic("failed to copy files into the temporary folder, " + err.Error())
	}
	log.Print("file copying to temporary folder finished")

	log.Print("processing files started")
	processedSongs, validSongs, totalSongs := logic.ProcessFiles(tempDirName, filesToSkip)
	log.Print("processing files finished")

	log.Print("processing folders started")
	logic.ProcessFolders(tempDirName)
	log.Print("processing folders finished")

	log.Print("processed ", processedSongs, " songs, ", validSongs, " valid songs, out of ", totalSongs, " total songs")

	log.Print("file copying to output folder started")
	if err := copy.Copy(tempDirName, outdir); err != nil {
		log.Panic("failed to copy files into the temporary folder, " + err.Error())
	}
	log.Print("file copying to output folder finished")
}
