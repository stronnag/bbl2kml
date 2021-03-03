package main

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/storage"
	"fmt"
	"path/filepath"
	"os"
	"strconv"
	"strings"
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
	files, _ := options.ParseCLI(getVersion)
	if options.Dump {
		fmt.Fprintln(os.Stderr, "Dump only supported via CLI")
		return
	}
	var runbtn *widget.Button

	a := app.New()
	w := a.NewWindow("flightlog2kml")

	var logfile, missionfile string

	lfname := widget.NewLabel("")

	mfname := widget.NewLabel("")
	lbut := widget.NewButton("Log File", func() {
		filter := storage.NewExtensionFileFilter([]string{".csv", ".TXT"})
		fc := dialog.NewFileOpen(
			func(cb fyne.URIReadCloser, err error) {
				if cb != nil {
					uri := cb.URI().String()
					logfile = strings.Replace(uri, "file://", "", 1)
					lfname.SetText(cb.URI().Name())
					runbtn.Enable()
				}
			},
			w)
		fc.SetFilter(filter)
		if len(logfile) > 0 {
			df := filepath.Dir(logfile)
			uri := storage.NewURI("file://" + df)
			lsu, err := storage.ListerForURI(uri)
			if err == nil {
				fc.SetLocation(lsu)
			}
		}
		fc.Show()
	})

	mbut := widget.NewButton("Mission File", func() {
		fc := dialog.NewFileOpen(
			func(cb fyne.URIReadCloser, err error) {
				if cb != nil {
					missionfile = cb.Name()
					mfname.SetText(missionfile)
				}
			},
			w)
		fc.Show()
	})

	obut := widget.NewButton("Output ...", func() {
		fc := dialog.NewFolderOpen(
			func(cb fyne.ListableURI, err error) {
				if cb != nil {
					odir := cb.String()
					options.Outdir = strings.Replace(odir, "file://", "", 1)
				}
			},
			w)
		fc.Show()
	})

	blank := widget.NewLabel(" ")
	blankmax := fyne.NewContainerWithLayout(layout.NewMaxLayout())
	blankmax.AddObject(blank)
	lfmax := fyne.NewContainerWithLayout(layout.NewMaxLayout())
	lfmax.AddObject(lfname)
	mfmax := fyne.NewContainerWithLayout(layout.NewMaxLayout())
	mfmax.AddObject(mfname)

	idxentry := widget.NewEntry()
	idxentry.SetText("0")
	idxbox := fyne.NewContainerWithLayout(layout.NewFormLayout())
	idxbox.AddObject(widget.NewLabel("Index"))
	idxbox.AddObject(idxentry)

	grid1 := fyne.NewContainerWithLayout(layout.NewGridLayout(3))
	grid1.AddObject(lbut)
	grid1.AddObject(lfmax)
	grid1.AddObject(idxbox)
	grid1.AddObject(mbut)
	grid1.AddObject(mfname)
	grid1.AddObject(obut)

	grid2 := fyne.NewContainerWithLayout(layout.NewGridLayout(2))
	dmsck := widget.NewCheck("DMS", func(b bool) { options.Dms = b })
	extck := widget.NewCheck("Extrude", func(b bool) { options.Extrude = b })
	rssck := widget.NewCheck("RSSI is default", func(b bool) { options.Rssi = b })
	effck := widget.NewCheck("Efficiency Layer", func(b bool) { options.Efficiency = b })
	kmlck := widget.NewCheck("KML", func(b bool) { options.Kml = b })

	var gopts []string = []string{"Red", "Red/Green", "Orange/Yellow/Red"}
	gradcombo := widget.NewSelect(gopts, func(s string) {

	})

	n := 0
	switch options.Gradset {
	case "rdgn":
		n = 1
	case "yor":
		n = 2
	}
	gradcombo.SetSelected(gopts[n])
	gradbox := fyne.NewContainerWithLayout(layout.NewFormLayout())
	gradbox.AddObject(widget.NewLabel("Gradient"))
	gradbox.AddObject(gradcombo)

	grid2.AddObject(dmsck)
	grid2.AddObject(rssck)
	grid2.AddObject(extck)
	grid2.AddObject(kmlck)
	grid2.AddObject(effck)
	grid2.AddObject(gradbox)

	rssck.SetChecked(options.Rssi)
	extck.SetChecked(options.Extrude)
	effck.SetChecked(options.Efficiency)
	kmlck.SetChecked(options.Kml)
	dmsck.SetChecked(options.Dms)

	tg := widget.NewTextGrid()
	sc := widget.NewScrollContainer(tg)
	sc.SetMinSize(fyne.NewSize(640, 200))
	tg.SetText("")
	middle := fyne.NewContainerWithLayout(layout.NewMaxLayout())
	middle.AddObject(sc)
	runbtn = widget.NewButton("Run", func() {
		idx_txt := idxentry.Text
		idx, err := strconv.ParseInt(idx_txt, 10, 32)
		if err != nil || idx < 0 || idx > 128 {
			idx = 0
			idxentry.SetText("0")
		}
		options.Idx = int(idx)
		options.Dms = dmsck.Checked
		options.Rssi = rssck.Checked
		options.Extrude = extck.Checked
		options.Efficiency = effck.Checked
		geo.Frobnicate_init()
		var lfr types.FlightLog
		ftype := types.EvinceFileType(logfile)
		if ftype == types.IS_OTX {
			olfr := otx.NewOTXReader(logfile)
			lfr = &olfr
		} else if ftype == types.IS_BBL {
			blfr := bbl.NewBBLReader(logfile)
			lfr = &blfr
		} else {
			return
		}
		metas, err := lfr.GetMetas()
		if err == nil {
			for _, b := range metas {
				if (options.Idx == 0 || options.Idx == b.Index) && b.Flags&types.Is_Valid != 0 {
					for k, v := range b.Summary() {
						add_textview(tg, sc, fmt.Sprintf("%-8.8s : %s\n", k, v))
					}
					ls, res := lfr.Reader(b)
					if res {
						outfn := kmlgen.GenKmlName(b.Logname, b.Index)
						kmlgen.GenerateKML(ls.H, ls.L, outfn, b, ls.M)
					}
					for k, v := range ls.M {
						add_textview(tg, sc, fmt.Sprintf("%-8.8s : %s\n", k, v))
					}
					if s, ok := b.ShowDisarm(); ok {
						add_textview(tg, sc, fmt.Sprintf("%-8.8s : %s\n\n", "Disarm", s))
					}
					if !res {
						fmt.Fprintf(os.Stderr, "*** skipping KML/Z for log  with no valid geospatial data\n")
						dialog.ShowInformation("Error", "Failed to parse file", w)
					}
				}
			}
		} else {
			dialog.ShowInformation("Error", err.Error(), w)
		}
	})

	if len(files) > 0 {
		logfile = files[0]
		lfname.SetText(filepath.Base(logfile))
	} else {
		runbtn.Disable()
	}
	bottom := fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, nil, nil, runbtn), runbtn)
	vbox := widget.NewVBox(grid1, grid2, middle, bottom)
	w.SetContent(vbox)
	w.ShowAndRun()
}

func add_textview(tg *widget.TextGrid, sc *widget.ScrollContainer, s string) {
	t := tg.Text()
	t = t + s
	tg.SetText(t)
	sc.ScrollToBottom() // fyne 1.4.3 NEW API
}
