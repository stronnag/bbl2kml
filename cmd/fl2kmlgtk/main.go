package main

// #cgo CFLAGS: -O2 -Wall
// #cgo pkg-config: gio-2.0
// #include "resources.h"
import "C"

import (
	"log"
	"os"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"strconv"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	options "github.com/stronnag/bbl2kml/pkg/options"
	otx "github.com/stronnag/bbl2kml/pkg/otx"
	bbl "github.com/stronnag/bbl2kml/pkg/bbl"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	kmlgen "github.com/stronnag/bbl2kml/pkg/kmlgen"
)

var GitCommit = "local"
var GitTag = "0.0.0"

func getVersion() string {
	return fmt.Sprintf("%s %s, commit: %s", filepath.Base(os.Args[0]), GitTag, GitCommit)
}


func main() {

	gtk.Init(nil)

	C.resources_register_resource()

	files, _ := options.ParseCLI(getVersion)
	if options.Dump {
		fmt.Fprintln(os.Stderr, "Dump only supported via CLI")
		return
	}

	b, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal("builder:", err)
	}
	img, err := gtk.ImageNewFromResource("/org/bbl2kml/fl2kmlgtk/logo.svg")
	if err != nil {
		log.Fatal("pix lookup:", err)
	}
	pbuf := img.GetPixbuf()

	err = b.AddFromResource("/org/bbl2kml/fl2kmlgtk/logkml.ui")
	if err != nil {
		log.Fatal("glade ui:", err)
	}

	obj, err := b.GetObject("appwin")
	if err != nil {
		log.Fatal("lookup:", err)
	}
	win := obj.(*gtk.Window)
	win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	win.SetIcon(pbuf)

	obj, err = b.GetObject("runbtn")
	runbtn := obj.(*gtk.Button)
	/*
		logchooser, _ := gtk.FileChooserDialogNewWith2Buttons(
			"Log file", win, gtk.FILE_CHOOSER_ACTION_OPEN,
			"OK", gtk.RESPONSE_OK, "Cancel", gtk.RESPONSE_CANCEL)
	*/
	logchooser, _ := gtk.FileChooserNativeDialogNew(
		"Log file", win, gtk.FILE_CHOOSER_ACTION_OPEN,
		"OK", "Cancel")
	filter, err := gtk.FileFilterNew()
	filter.SetName("All Logs")
	filter.AddPattern("*.bbl")
	filter.AddPattern("*.BBL")
	filter.AddPattern("*.TXT")
	filter.AddPattern("*.txt")
	filter.AddPattern("*.csv")
	filter.AddPattern("*.CSV")
	logchooser.AddFilter(filter)
	filter, err = gtk.FileFilterNew()
	filter.SetName("BBox Logs")
	filter.AddPattern("*.bbl")
	filter.AddPattern("*.BBL")
	filter.AddPattern("*.TXT")
	filter.AddPattern("*.txt")
	logchooser.AddFilter(filter)
	filter, err = gtk.FileFilterNew()
	filter.SetName("OpenTX Logs")
	filter.AddPattern("*.csv")
	filter.AddPattern("*.CSV")
	logchooser.AddFilter(filter)
	filter, err = gtk.FileFilterNew()
	filter.SetName("All files")
	filter.AddPattern("*")
	logchooser.AddFilter(filter)
	logchooser.SetSelectMultiple(true)

	obj, err = b.GetObject("log_label")
	loglbl := obj.(*gtk.Entry)

	obj, err = b.GetObject("mission_label")
	missionlbl := obj.(*gtk.Entry)

	obj, err = b.GetObject("log_btn")
	logbtn := obj.(*gtk.Button)
	logbtn.Connect("clicked", func() {
		id := logchooser.Run()
		if id == int(gtk.RESPONSE_OK) || id == int(gtk.RESPONSE_ACCEPT) {
			fs, err := logchooser.GetFilenames()
			if err == nil {
				files = fs
				var sb strings.Builder
				for k, s := range fs {
					bn := filepath.Base(s)
					if k != 0 {
						sb.WriteByte(',')
					}
					sb.WriteString(bn)
				}
				loglbl.SetText(sb.String())
				runbtn.SetSensitive(true)
			}
		}
		logchooser.Hide()
	})

	missionchooser, _ := gtk.FileChooserNativeDialogNew(
		"Mission file", win, gtk.FILE_CHOOSER_ACTION_OPEN,
		"OK", "Cancel")

	filter, err = gtk.FileFilterNew()
	filter.SetName("All inav Missions")
	filter.AddPattern("*.mission")
	filter.AddPattern("*.json")
	missionchooser.AddFilter(filter)
	filter, err = gtk.FileFilterNew()
	filter.SetName("All files")
	filter.AddPattern("*")
	missionchooser.AddFilter(filter)

	obj, err = b.GetObject("mission_btn")
	missionbtn := obj.(*gtk.Button)
	missionbtn.Connect("clicked", func() {
		id := missionchooser.Run()
		if id == int(gtk.RESPONSE_OK) || id == int(gtk.RESPONSE_ACCEPT) {
			options.Mission = missionchooser.GetFilename()
			missionlbl.SetText(filepath.Base(options.Mission))
		}
		missionchooser.Hide()
	})

	/*
		outchooser, _ := gtk.FileChooserDialogNewWith2Buttons(
		"Output Directory", win, gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER,
		"OK", gtk.RESPONSE_OK,
		"Cancel", gtk.RESPONSE_CANCEL)
	*/
	outchooser, _ := gtk.FileChooserNativeDialogNew(
		"Output Directory", win, gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER,
		"OK", "Cancel")

	if len(options.Outdir) > 0 {
		outchooser.SetCurrentFolder(options.Outdir)
	}
	obj, err = b.GetObject("out_btn")
	outbtn := obj.(*gtk.Button)
	outbtn.Connect("clicked", func() {
		id := outchooser.Run()
		if id == int(gtk.RESPONSE_OK) || id == int(gtk.RESPONSE_ACCEPT) {
			options.Outdir = outchooser.GetFilename()
		}
		outchooser.Hide()
	})

	runbtn.SetSensitive(false)

	if len(files) > 0 {
		fn := filepath.Base(files[0])
		loglbl.SetText(fn)
		dir := filepath.Dir(files[0])
		logchooser.SetCurrentFolder(dir)
		runbtn.SetSensitive(true)
	}

	obj, err = b.GetObject("textview")
	textview := obj.(*gtk.TextView)

	obj, err = b.GetObject("dms_check")
	dms_check := obj.(*gtk.CheckButton)
	obj, err = b.GetObject("rssi_check")
	rssi_check := obj.(*gtk.CheckButton)
	obj, err = b.GetObject("extrude_check")
	extrude_check := obj.(*gtk.CheckButton)
	obj, err = b.GetObject("effic_check")
	effic_check := obj.(*gtk.CheckButton)
	obj, err = b.GetObject("kml_check")
	kml_check := obj.(*gtk.CheckButton)
	obj, err = b.GetObject("grad_combo")
	grad_combo := obj.(*gtk.ComboBoxText)

	dms_check.SetActive(options.Dms)
	kml_check.SetActive(options.Kml)
	rssi_check.SetActive(options.Rssi)
	extrude_check.SetActive(options.Extrude)
	effic_check.SetActive(options.Efficiency)
	gradopts := []string{"red", "rdgn", "yor"}

	n := 0
	switch options.Gradset {
	case gradopts[1]:
		n = 1
	case gradopts[2]:
		n = 2
	}
	grad_combo.SetActive(n)
	obj, err = b.GetObject("idx_entry")
	idx_entry := obj.(*gtk.Entry)

	runbtn.Connect("clicked", func() {
		runbtn.SetSensitive(false)

		idx_txt, err := idx_entry.GetText()
		idx := int64(0)
		if err == nil {
			idx, err = strconv.ParseInt(idx_txt, 10, 32)
		}
		if err != nil || idx < 0 || idx > 128 {
			idx = 0
			idx_entry.SetText("0")
		}
		options.Idx = int(idx)
		options.Dms = dms_check.GetActive()
		options.Kml = kml_check.GetActive()
		options.Rssi = rssi_check.GetActive()
		options.Extrude = extrude_check.GetActive()
		options.Efficiency = effic_check.GetActive()
		n := grad_combo.GetActive()
		options.Gradset = gradopts[n]

		//		pdl, _ := gtk.DialogNew()
		pdl := gtk.MessageDialogNew(win,
			gtk.DIALOG_MODAL|gtk.DIALOG_DESTROY_WITH_PARENT,
			gtk.MESSAGE_INFO, gtk.BUTTONS_NONE,
			"Processing")

		pbar, _ := gtk.ProgressBarNew()
		bx, _ := pdl.GetContentArea()
		bx.PackStart(pbar, false, false, 2)
		pdl.SetTransientFor(win)
		pdl.ShowAll()
		working := true
		glib.TimeoutAdd(50, func() bool {
			if working {
				pbar.Pulse()
			} else {
				pbar.Destroy()
				pdl.Destroy()
				runbtn.SetSensitive(true)
			}
			return working
		})

		go func() {
			geo.Frobnicate_init()
			var lfr types.FlightLog
			for _, fn := range files {
				ftype := types.EvinceFileType(fn)
				if ftype == types.IS_OTX {
					olfr := otx.NewOTXReader(fn)
					lfr = &olfr
				} else if ftype == types.IS_BBL {
					blfr := bbl.NewBBLReader(fn)
					lfr = &blfr
				} else {
					continue
				}
				metas, err := lfr.GetMetas()
				if err == nil {
					for _, b := range metas {
						if (options.Idx == 0 || options.Idx == b.Index) && b.Flags&types.Is_Valid != 0 {
							for k, v := range b.Summary() {
								add_textview(textview, fmt.Sprintf("%-8.8s : %s\n", k, v))
							}
							ls, res := lfr.Reader(b)
							if res {
								outfn := kmlgen.GenKmlName(b.Logname, b.Index)
								kmlgen.GenerateKML(ls.H, ls.L, outfn, b, ls.M)
							}
							for k, v := range ls.M {
								add_textview(textview, fmt.Sprintf("%-8.8s : %s\n", k, v))
							}
							if s, ok := b.ShowDisarm(); ok {
								add_textview(textview, fmt.Sprintf("%-8.8s : %s\n\n", "Disarm", s))
							}
							if !res {
								msg := "Skipping KML/Z for log  with no valid geospatial data"
								if lfr.LogType() == 'B' {
									msg += "\nMaybe blackbox_decode is too old?"
								}
								show_dialogue(win, msg)
							}
						}
					}
				} else {
					show_dialogue(win, err.Error())
				}
			}
			working = false
		}()
	})

	_, err = exec.LookPath(options.Blackbox_decode)
	if err != nil {
		glib.IdleAdd(func() {})
		show_dialogue(win, "Missing blackbox_decode")
	} else {
		win.ShowAll()
		gtk.Main()
	}
}

func show_dialogue(win *gtk.Window, msg string) {
	dlg := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL|gtk.DIALOG_DESTROY_WITH_PARENT, gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, msg)
	dlg.Run()
	dlg.Destroy()
}

func add_textview(textview *gtk.TextView, s string) {
	glib.IdleAdd(func() {
		textbuf, _ := textview.GetBuffer()
		iter := textbuf.GetEndIter()
		textbuf.Insert(iter, s)
		iter = textbuf.GetEndIter()
		textview.ScrollToIter(iter, 0.0, true, 0.0, 1.0)
	})
}
