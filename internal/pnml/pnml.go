// Copyright (c) 2021 Silvano DAL ZILIO
//
// GNU Affero GPL v3

package pnml

import (
	"encoding/xml"
	"fmt"
	"io"
)

const (
	// DOCTYPE for the generated PNML file
	DOCTYPE = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

// PT is the type of PNML for a P/T net without graphical information
type PT struct {
	XMLName xml.Name `xml:"http://www.pnml.org/version-2009/grammar/pnml pnml"`
	WNET    Net      `xml:"net"`
}

// Net is the type of PNML net, without graphical information, where all
// information is written in a single page.
type Net struct {
	Thetype string `xml:"type,attr"`
	ID      string `xml:"id,attr"`
	NAME    string `xml:"name>text"`
	PAGE    Page   `xml:"page"`
}

// Page is the unit for defining a P/T net inside a PNML file.
type Page struct {
	ID     string  `xml:"id,attr"`
	PLACES []Place `xml:"place"`
	TRANS  []Trans `xml:"transition"`
}

// Place is the type used to marshal places.
type Place struct {
	Name  string
	Label string
	Init  int
}

// Trans is the type used to marshal transitions. We keep a pointer to the net
// so that we can find references to the arcs. We do not support inhibitor arcs.
type Trans struct {
	Name    string
	Label   string
	In, Out []Arc
}

// Arc is a pair of a place and a multiplicity. This is used to build arcs in
// the unfolding of a hlnet.
type Arc struct {
	Place *Place
	Mult  int
}

// MarshalXML encodes the receiver as zero or more XML elements. This makes
// Place a xml.Marshaller
func (v Place) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Attr = []xml.Attr{{Name: xml.Name{Local: "id"}, Value: "pl_" + v.Name}}
	e.EncodeToken(start)
	e.EncodeToken(xml.StartElement{Name: xml.Name{Local: "name"}})
	if v.Label != "" {
		e.EncodeElement(v.Name+": "+v.Label, xml.StartElement{Name: xml.Name{Local: "text"}})
	} else {
		e.EncodeElement(v.Name, xml.StartElement{Name: xml.Name{Local: "text"}})

	}
	e.EncodeToken(xml.EndElement{Name: xml.Name{Local: "name"}})
	if v.Init != 0 {
		e.EncodeToken(xml.StartElement{Name: xml.Name{Local: "initialMarking"}})
		e.EncodeElement(v.Init, xml.StartElement{Name: xml.Name{Local: "text"}})
		e.EncodeToken(xml.EndElement{Name: xml.Name{Local: "initialMarking"}})
	}
	e.EncodeToken(xml.EndElement{Name: start.Name})
	return nil
}

// MarshalXML encodes the receiver as zero or more XML elements. This makes
// Trans a xml.Marshaller
func (v Trans) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Attr = []xml.Attr{{Name: xml.Name{Local: "id"}, Value: "tr_" + v.Name}}
	e.EncodeToken(start)
	e.EncodeToken(xml.StartElement{Name: xml.Name{Local: "name"}})
	if v.Label != "" {
		e.EncodeElement(v.Name+": "+v.Label, xml.StartElement{Name: xml.Name{Local: "text"}})
	} else {
		e.EncodeElement(v.Name, xml.StartElement{Name: xml.Name{Local: "text"}})

	}
	e.EncodeToken(xml.EndElement{Name: xml.Name{Local: "name"}})
	e.EncodeToken(xml.EndElement{Name: start.Name})

	for _, c := range v.In {
		encodeArc(e, fmt.Sprintf("p2t-%s-%s", c.Place.Name, v.Name), "pl_"+c.Place.Name, "tr_"+v.Name, c.Mult)
	}
	for _, c := range v.Out {
		encodeArc(e, fmt.Sprintf("t2p-%s-%s", v.Name, c.Place.Name), "tr_"+v.Name, "pl_"+c.Place.Name, c.Mult)
	}

	return nil
}

func encodeArc(e *xml.Encoder, id, src, tgt string, weight int) {
	arc := xml.StartElement{
		Name: xml.Name{Local: "arc"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "id"}, Value: id},
			{Name: xml.Name{Local: "source"}, Value: src},
			{Name: xml.Name{Local: "target"}, Value: tgt},
		},
	}
	e.EncodeToken(arc)
	if weight != 1 {
		e.EncodeToken(xml.StartElement{Name: xml.Name{Local: "inscription"}})
		e.EncodeElement(weight, xml.StartElement{Name: xml.Name{Local: "text"}})
		e.EncodeToken(xml.EndElement{Name: xml.Name{Local: "inscription"}})
	}
	e.EncodeToken(xml.EndElement{Name: xml.Name{Local: "arc"}})
}

// Write prints a P/T net in PNML format on an io.Writer
func Write(w io.Writer, name string, pl []Place, tr []Trans) error {
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")

	// Now we output the file on the io.Writer
	wpnml := PT{
		WNET: Net{
			Thetype: "http://www.pnml.org/version-2009/grammar/ptnet",
			ID:      name,
			NAME:    name,
			PAGE: Page{
				ID:     "page",
				PLACES: pl,
				TRANS:  tr,
			},
		},
	}
	w.Write([]byte(DOCTYPE))
	return encoder.Encode(wpnml)
}
