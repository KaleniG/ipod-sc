package main

import (
	"context"
	"log"
	"os"

	"ipod-sc/internal/logic"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	copy "github.com/otiai10/copy"
)

type TextViewWriter struct {
	buffer *gtk.TextBuffer
}

func (w *TextViewWriter) Write(p []byte) (int, error) {
	text := string(p)

	glib.IdleAdd(func() {
		iter := w.buffer.EndIter()
		w.buffer.Insert(iter, text)
	})

	return len(p), nil
}

func main() {
	ctx := context.Background()
	app := gtk.NewApplication("com.ipod-sc.kalenig", 0)

	// APP CREATION
	app.ConnectActivate(func() {
		browseWindowOpen := false

		// WINDOW CREATION
		win := gtk.NewApplicationWindow(app)
		win.SetTitle("iPod Song Converter")
		win.SetDefaultSize(500, 120)

		// WINDOW BOX
		box := gtk.NewBox(gtk.OrientationVertical, 10)
		box.SetMarginTop(10)
		box.SetMarginBottom(10)
		box.SetMarginStart(10)
		box.SetMarginEnd(10)

		// SCROLLABLE AREA WITH TEXTBOX AND LOG OUTPUT TO THE TEXTBOX
		scrolled := gtk.NewScrolledWindow()
		scrolled.SetVExpand(true)
		scrolled.SetHExpand(true)

		textView := gtk.NewTextView()
		textView.SetEditable(false)
		textView.SetCursorVisible(false)
		textView.SetWrapMode(gtk.WrapWord)

		buffer := textView.Buffer()

		writer := &TextViewWriter{
			buffer: buffer,
		}

		log.SetOutput(writer)

		scrolled.SetChild(textView)
		scrolled.SetVisible(false)

		// BROWSE DIRECTORY HANDLER
		openFolder := func(ctx context.Context, entry *gtk.Entry) {

			dialog := gtk.NewFileDialog()

			dialog.SelectFolder(ctx, nil, func(result gio.AsyncResulter) {
				file, err := dialog.SelectFolderFinish(result)
				if err != nil || file == nil {
					browseWindowOpen = false
					return
				}

				entry.SetText(file.Path())
				browseWindowOpen = false
			})
		}

		// SKIP ALREADY TRANSFERRED VALID FILES IN OUTPUT FOLDER CHECKBOX
		check := gtk.NewCheckButtonWithLabel("Skip files already loaded")

		// INPUT DIRECTORY (INPUT FIELD AND BROWSE BUTTON)
		sourceEntry := gtk.NewEntry()
		sourceEntry.SetHExpand(true)
		sourceEntry.SetPlaceholderText("Select source folder")

		sourceButton := gtk.NewButtonWithLabel("Browse")

		sourceButton.ConnectClicked(func() {
			if !browseWindowOpen {
				browseWindowOpen = true
				openFolder(ctx, sourceEntry)
				scrolled.SetVisible(false)
			}
		})

		sourceRow := gtk.NewBox(gtk.OrientationHorizontal, 5)
		sourceRow.Append(sourceEntry)
		sourceRow.Append(sourceButton)

		// OUTPUT DIRECTORY (INPUT FIELD AND BROWSE BUTTON)
		outputEntry := gtk.NewEntry()
		outputEntry.SetHExpand(true)
		outputEntry.SetPlaceholderText("Select output folder")

		outputButton := gtk.NewButtonWithLabel("Browse")

		outputButton.ConnectClicked(func() {
			if !browseWindowOpen {
				browseWindowOpen = true
				openFolder(ctx, outputEntry)
				scrolled.SetVisible(false)
			}
		})

		outputRow := gtk.NewBox(gtk.OrientationHorizontal, 5)
		outputRow.Append(outputEntry)
		outputRow.Append(outputButton)

		// START OPERATION BUTTON WITH SPINNER AND LABELS
		startButton := gtk.NewButtonWithLabel("Start")
		startButton.SetHAlign(gtk.AlignCenter)
		startButton.SetVAlign(gtk.AlignCenter)
		hbox := gtk.NewBox(gtk.OrientationHorizontal, 6)

		spinner := gtk.NewSpinner()
		spinner.SetSpinning(false)
		spinner.SetVisible(false)

		label := gtk.NewLabel("Start")

		hbox.Append(spinner)
		hbox.Append(label)
		hbox.SetHAlign(gtk.AlignCenter)
		hbox.SetVAlign(gtk.AlignCenter)

		startButton.SetChild(hbox)

		// START OPERATION ON CLICK HANDLER
		startButton.ConnectClicked(func() {
			spinner.SetVisible(true)
			spinner.SetSpinning(true)
			scrolled.SetVisible(true)
			buffer := textView.Buffer()
			buffer.SetText("")
			label.SetText("Processing...")

			go func() {
				// FILES AND FOLDERS PROCESSING AND COPYING TO OUPUT
				indir := sourceEntry.Text()
				outdir := outputEntry.Text()

				filesToSkip := []string{}
				if check.Active() {
					filesToSkip = logic.GetFilesToSkip(indir, outdir)
				}

				if logic.DirExists(indir) {
					log.Print("input folder [" + indir + "] exists")
				} else {
					log.Print("input folder [" + indir + "] does not exist")
				}

				if logic.DirExists(outdir) {
					log.Print("output folder [" + outdir + "] exists")
				} else {
					log.Print("output folder [" + outdir + "] does not exist")
				}

				if logic.DirExists(outdir) && logic.DirExists(indir) {
					tempDirName, err := os.MkdirTemp("", "ipod-sc-*")
					if err != nil {
						log.Print("failed to create a temporary folder, " + err.Error())
					}
					log.Print("temporary folder for copy created")
					defer os.RemoveAll(tempDirName)

					log.Print("file copying to temporary folder started")
					if err := copy.Copy(indir, tempDirName); err != nil {
						log.Print("failed to copy files into the temporary folder, " + err.Error())
					} else {
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
							log.Print("failed to copy files into the temporary folder, " + err.Error())
						} else {
							log.Print("file copying to output folder finished")
						}
					}
				}

				glib.IdleAdd(func() {
					spinner.SetSpinning(false)
					spinner.SetVisible(false)
					label.SetText("Start")
				})
			}()
		})

		// FINAL APPEND OF ALL THE ELEMENTS
		box.Append(check)
		box.Append(sourceRow)
		box.Append(outputRow)
		box.Append(startButton)
		box.Append(scrolled)

		win.SetChild(box)
		win.Present()
	})

	app.Run(nil)
}
