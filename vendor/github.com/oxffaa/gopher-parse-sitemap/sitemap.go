// Package sitemap provides primitives for high effective parsing of huge
// sitemap files.
package sitemap

import (
	"encoding/xml"
	"io"
	"net/http"
	"os"
	"time"
)

// Frequency is a type alias for change frequency.
type Frequency = string

// Change frequency constants set describes how frequently a page is changed.
const (
	Always  Frequency = "always"  // A page is changed always
	Hourly  Frequency = "hourly"  // A page is changed every hour
	Daily   Frequency = "daily"   // A page is changed every day
	Weekly  Frequency = "weekly"  // A page is changed every week
	Monthly Frequency = "monthly" // A page is changed every month
	Yearly  Frequency = "yearly"  // A page is changed every year
	Never   Frequency = "never"   // A page is changed never
)

// Entry is an interface describes an element \ an URL in the sitemap file.
// Keep in mind. It is implemented by a totally immutable entity so you should
// minimize calls count because it can produce additional memory allocations.
//
// GetLocation returns URL of the page.
// GetLocation must return a non-nil and not empty string value.
//
// GetLastModified parses and returns date and time of last modification of the page.
// GetLastModified can return nil or a valid time.Time instance.
// Be careful. Each call return new time.Time instance.
//
// GetChangeFrequency returns string value indicates how frequent the page is changed.
// GetChangeFrequency returns non-nil string value. See Frequency consts set.
//
// GetPriority return priority of the page.
// The valid value is between 0.0 and 1.0, the default value is 0.5.
//
// You shouldn't implement this interface in your types.
type Entry interface {
	GetLocation() string
	GetLastModified() *time.Time
	GetChangeFrequency() Frequency
	GetPriority() float32
}

// IndexEntry is an interface describes an element \ an URL in a sitemap index file.
// Keep in mind. It is implemented by a totally immutable entity so you should
// minimize calls count because it can produce additional memory allocations.
//
// GetLocation returns URL of a sitemap file.
// GetLocation must return a non-nil and not empty string value.
//
// GetLastModified parses and returns date and time of last modification of sitemap.
// GetLastModified can return nil or a valid time.Time instance.
// Be careful. Each call return new time.Time instance.
//
// You shouldn't implement this interface in your types.
type IndexEntry interface {
	GetLocation() string
	GetLastModified() *time.Time
}

// EntryConsumer is a type represents consumer of parsed sitemaps entries
type EntryConsumer func(Entry) error

// Parse parses data which provides by the reader and for each sitemap
// entry calls the consumer's function.
func Parse(reader io.Reader, consumer EntryConsumer) error {
	return parseLoop(reader, func(d *xml.Decoder, se *xml.StartElement) error {
		return entryParser(d, se, consumer)
	})
}

// ParseFromFile reads sitemap from a file, parses it and for each sitemap
// entry calls the consumer's function.
func ParseFromFile(sitemapPath string, consumer EntryConsumer) error {
	sitemapFile, err := os.OpenFile(sitemapPath, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer sitemapFile.Close()

	return Parse(sitemapFile, consumer)
}

// ParseFromSite downloads sitemap from a site, parses it and for each sitemap
// entry calls the consumer's function.
func ParseFromSite(url string, consumer EntryConsumer) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return Parse(res.Body, consumer)
}

// IndexEntryConsumer is a type represents consumer of parsed sitemaps indexes entries
type IndexEntryConsumer func(IndexEntry) error

// ParseIndex parses data which provides by the reader and for each sitemap index
// entry calls the consumer's function.
func ParseIndex(reader io.Reader, consumer IndexEntryConsumer) error {
	return parseLoop(reader, func(d *xml.Decoder, se *xml.StartElement) error {
		return indexEntryParser(d, se, consumer)
	})
}

// ParseIndexFromFile reads sitemap index from a file, parses it and for each sitemap
// index entry calls the consumer's function.
func ParseIndexFromFile(sitemapPath string, consumer IndexEntryConsumer) error {
	sitemapFile, err := os.OpenFile(sitemapPath, os.O_RDONLY, os.ModeExclusive)
	if err != nil {
		return err
	}
	defer sitemapFile.Close()

	return ParseIndex(sitemapFile, consumer)
}

// ParseIndexFromSite downloads sitemap index from a site, parses it and for each sitemap
// index entry calls the consumer's function.
func ParseIndexFromSite(sitemapURL string, consumer IndexEntryConsumer) error {
	res, err := http.Get(sitemapURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return ParseIndex(res.Body, consumer)
}
