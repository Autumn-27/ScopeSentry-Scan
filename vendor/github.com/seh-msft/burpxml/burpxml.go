// Copyright (c) 2020, Microsoft Corporation, Sean Hinchee
// Licensed under the MIT License.

// Package burpxml provides functions to parse and re-format
// Burp Suite HTTP proxy history XML files
package burpxml

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Request represents an XML <request> HTTP request.
type Request struct {
	Base64 string `xml:"base64,attr"` // Is Raw base64-encoded?
	Raw    string `xml:",chardata"`   // Raw HTTP request
	Body   string // May be base64-decoded later
}

// Response represents an XML <response> HTTP response.
type Response struct {
	Base64 string `xml:"base64,attr"` // Is Raw base64-encoded?
	Raw    string `xml:",chardata"`   // Raw HTTP response
	Body   string // May be base64-decoded later
}

// Host represents an XML <host> remote host.
// In textual output, Host's elements should appear
// as though they were Item's elements.
type Host struct {
	Ip   string `xml:"ip,attr"`   // Remote IP address
	Name string `xml:",chardata"` // Remote host name
}

// Item represents an XML <item> HTTP transaction.
type Item struct {
	Time           string   `xml:"time"`           // Time the request was sent
	Url            string   `xml:"url"`            // URL requested
	Host           Host     `xml:"host"`           // Remote host name
	Port           string   `xml:"port"`           // Remote port
	Protocol       string   `xml:"protocol"`       // Protocol (such as HTTPS)
	Path           string   `xml:"path"`           // URL path requested (/foo/bar/)
	Extension      string   `xml:"extension"`      // Burp-specific
	Request        Request  `xml:"request"`        // HTTP request made
	Status         string   `xml:"status"`         // HTTP status code returned
	ResponseLength string   `xml:"responselength"` // HTTP response content length
	MimeType       string   `xml:"mimetype"`       // MIME type of HTTP response
	Response       Response `xml:"response"`       // HTTP response returned
	Comment        string   `xml:"comment"`        // Burp-specific
}

// Items represents the set of XML <items> containing many <item>'s.
type Items struct {
	Items []Item `xml:"item"`
}

// Csv emits items to CSV.
// Requests and responses may be omitted, if desired.
func (items Items) Csv(of io.Writer, noReq, noResp bool) ([][]string, error) {
	enc := csv.NewWriter(of)

	// Make items one big slice of strings
	var strings [][]string
	for _, item := range items.Items {
		strings = append(strings, item.ToStrings(noReq, noResp))
	}

	err := enc.WriteAll(strings)
	if err != nil {
		return nil, errors.New("could not convert xml to csv ⇒ " + err.Error())
	}
	enc.Flush()

	return strings, nil
}

// Json emits items to JSON.
func (items Items) Json(of io.Writer) error {
	enc := json.NewEncoder(of)
	err := enc.Encode(items)

	if err != nil {
		return errors.New("could not convert xml to json ⇒ " + err.Error())
	}

	return nil
}

// Go returns a Go-syntax representation of items.
func (items Items) Go() string {
	return fmt.Sprintf("%#v\n", items)
}

// Parse will read XML from f.
// Optionally, base64-encoded request and response bodies may be checked
// and decoded.
func Parse(f io.Reader, decode bool) (Items, error) {
	var items Items

	dec := xml.NewDecoder(f)
	err := dec.Decode(&items)
	if err != nil {
		return items, errors.New("could not parse xml ⇒ " + err.Error())
	}

	if !decode {
		return items, nil
	}

	// Decode bas64 bodies, if needed
	for n, item := range items.Items {
		encoding := base64.StdEncoding

		b, err := strconv.ParseBool(item.Request.Base64)
		if err != nil {
			return items, errors.New("could not parse request base64 bool ⇒ " + err.Error())
		}
		if b {
			decoded, err := encoding.DecodeString(item.Request.Raw)
			if err != nil {
				return items, errors.New("could not decode base64 request ⇒ " + err.Error())
			}
			items.Items[n].Request.Body = string(decoded)
		}

		b, err = strconv.ParseBool(item.Response.Base64)
		if err != nil {
			return items, errors.New("could not parse response base64 bool ⇒ " + err.Error())
		}
		if b {
			decoded, err := encoding.DecodeString(item.Response.Raw)
			if err != nil {
				return items, errors.New("could not decode base64 response ⇒ " + err.Error())
			}
			items.Items[n].Response.Body = string(decoded)
		}
	}

	return items, nil
}
