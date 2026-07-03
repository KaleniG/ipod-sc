package main

import (
	"ipod-sc/libs"
	"log"
	"os"

	copy "github.com/otiai10/copy"
)

func main() {
	indir := os.Args[1]
	outdir := os.Args[2]

	if libs.DirExists(indir) {
		log.Print("input folder [" + indir + "] exists")
	} else {
		log.Panic("input folder [" + indir + "] does not exist")
	}

	if libs.DirExists(outdir) {
		log.Print("output folder [" + outdir + "] exists")
	} else {
		log.Panic("output folder [" + outdir + "] does not exist")
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
	processedSongs, validSongs, totalSongs := libs.ProcessFiles(tempDirName)
	log.Print("processing files finished")

	log.Print("processing folders started")
	libs.ProcessFolders(tempDirName)
	log.Print("processing folders finished")

	log.Print("processed ", processedSongs, " songs, ", validSongs, " valid songs, out of ", totalSongs, " total songs")

	log.Print("file copying to output folder started")
	if err := copy.Copy(tempDirName, outdir); err != nil {
		log.Panic("failed to copy files into the temporary folder, " + err.Error())
	}
	log.Print("file copying to output folder finished")
}
