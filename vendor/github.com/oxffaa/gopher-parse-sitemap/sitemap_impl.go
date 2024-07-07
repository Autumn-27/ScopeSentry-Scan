package sitemap

import (
	"encoding/xml"
	"io"
)

func entryParser(decoder *xml.Decoder, se *xml.StartElement, consume EntryConsumer) error {
	if se.Name.Local == "url" {
		entry := newSitemapEntry()

		decodeError := decoder.DecodeElement(entry, se)
		if decodeError != nil {
			return decodeError
		}

		consumerError := consume(entry)
		if consumerError != nil {
			return consumerError
		}
	}

	return nil
}

func indexEntryParser(decoder *xml.Decoder, se *xml.StartElement, consume IndexEntryConsumer) error {
	if se.Name.Local == "sitemap" {
		entry := new(sitemapIndexEntry)

		decodeError := decoder.DecodeElement(entry, se)
		if decodeError != nil {
			return decodeError
		}

		consumerError := consume(entry)
		if consumerError != nil {
			return consumerError
		}
	}

	return nil
}

type elementParser func(*xml.Decoder, *xml.StartElement) error

func parseLoop(reader io.Reader, parser elementParser) error {
	decoder := xml.NewDecoder(reader)

	for {
		t, tokenError := decoder.Token()

		if tokenError == io.EOF {
			break
		} else if tokenError != nil {
			return tokenError
		}

		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}

		parserError := parser(decoder, &se)
		if parserError != nil {
			return parserError
		}
	}

	return nil
}
