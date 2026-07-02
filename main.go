package main

import (
	"ipod-sc/libs"
	"log"
	"os"
)

func main() {
	dir := os.Args[1]

	processedSongs, validSongs, totalSongs := libs.ProcessFiles(dir)
	libs.ProcessFolders(dir)

	log.Print("processed ", processedSongs, " songs, ", validSongs, " valid songs, out of ", totalSongs, " total songs")
}
