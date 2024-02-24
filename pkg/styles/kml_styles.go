package styles

import (
	kml "github.com/twpayne/go-kml"
	"github.com/twpayne/go-kml/icon"
	"image/color"
)

func Get_safe_styles() []kml.Element {
	return []kml.Element{
		kml.SharedStyle(
			"styleSAFEHOME",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ylw-square")),
				),
			),
		),
	}
}

func Get_approach_styles() []kml.Element {
	return []kml.Element{
		kml.SharedStyle(
			"styleFWLand",
			kml.LineStyle(
				kml.Width(4.0),
				kml.Color(color.RGBA{R: 0xfc, G: 0xac, B: 0x64, A: 0xa0}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0xfc, G: 0xac, B: 0x64, A: 0}),
			),
		),
		kml.SharedStyle(
			"styleFWApproach",
			kml.LineStyle(
				kml.Width(4.0),
				kml.Color(color.RGBA{R: 0x63, G: 0xa0, B: 0xfc, A: 0xa0}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0x63, G: 0xa0, B: 0xfc, A: 0}),
			),
		),
	}
}

func Get_zone_styles() []kml.Element {
	return []kml.Element{
		kml.SharedStyle(
			"styleINC",
			kml.IconStyle(
				kml.Scale(1.0),
				kml.Icon(
					kml.Href(icon.PaddleHref("grn-circle")),
				),
				kml.Color(color.RGBA{R: 0, G: 0xff, B: 0, A: 0xa0}),
			),
			kml.LineStyle(
				kml.Width(4.0),
				kml.Color(color.RGBA{R: 0, G: 0xff, B: 0, A: 0xa0}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0, G: 0xff, B: 0, A: 0x1a}),
			),
		),
		kml.SharedStyle(
			"styleEXC",
			kml.IconStyle(
				kml.Scale(1.0),
				kml.Icon(
					kml.Href(icon.PaddleHref("red-circle")),
				),
				kml.Color(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xa0}),
			),
			kml.LineStyle(
				kml.Width(4.0),
				kml.Color(color.RGBA{R: 0xff, G: 0, B: 0, A: 0xa0}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0xff, G: 0, B: 0, A: 0x1a}),
			),
		),
	}
}

func Get_mission_styles() []kml.Element {
	return []kml.Element{
		kml.SharedStyle(
			"styleSET_POI",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ylw-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleRTH",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("red-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleSET_HEAD",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ylw-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleWAYPOINT",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("ltblu-circle")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"stylePOSHOLD_UNLIM",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("grn-diamond")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"stylePOSHOLD_TIME",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("grn-circle")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleJUMP",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("purple-circle")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleLAND",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("pink-stars")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),

		kml.SharedStyle(
			"styleFakeHome",
			kml.IconStyle(
				kml.Scale(0.8),
				kml.Icon(
					kml.Href(icon.PaddleHref("orange-stars")),
				),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`)),
		),
		kml.SharedStyle(
			"styleWPTrack",
			kml.LineStyle(
				kml.Width(4.0),
				kml.Color(color.RGBA{R: 0xff, G: 0, B: 0, A: 0x66}),
			),
			kml.PolyStyle(
				kml.Color(color.RGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0x66}),
			),
			kml.BalloonStyle(kml.BgColor(color.RGBA{R: 0xde, G: 0xde, B: 0xde, A: 0x40}),
				kml.Text(`<b><font size="+2">$[name]</font></b><br/><br/>$[description]<br/>`),
			),
		),
	}
}
