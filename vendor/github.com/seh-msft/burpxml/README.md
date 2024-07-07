# burpxml

[![PkgGoDev](https://pkg.go.dev/badge/github.com/seh-msft/burpxml)](https://pkg.go.dev/github.com/seh-msft/burpxml)

Go module to parse Burp Suite HTTP proxy history XML files. 

Output formats include CSV, JSON, and Go syntax. 

Options are provided for controlling base64 decoding for request/response bodies. 

Requests/responses may be optionally omitted for the CSV format output. 

See also: [bx](https://github.com/seh-msft/bx)

## Install

	go get github.com/seh-msft/burpxml

## Documentation

```
package burpxml

Package burpxml provides functions to parse and re-format Burp Suite HTTP
proxy history XML files

TYPES

type Host struct {
        Ip   string `xml:"ip,attr"`   // Remote IP address
        Name string `xml:",chardata"` // Remote host name
}
    Host represents an XML <host> remote host. In textual output, Host's
    elements should appear as though they were Item's elements.

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
    Item represents an XML <item> HTTP transaction.

func (i Item) FlatString() string
    FlatString returns an Item as a single-line string. Each element is
    comma-separated.

func (i Item) String() string
    String returns a prettified string representation of an Item.

func (i Item) ToStrings(noReq, noResp bool) []string
    ToStrings returns an Item as a slice of strings. Each element in the slice
    is an Item field.

type Items struct {
        Items []Item `xml:"item"`
}
    Items represents the set of XML <items> containing many <item>'s.

func Parse(f io.Reader, decode bool) (Items, error)
    Parse will read XML from f. Optionally, base64-encoded request and response
    bodies may be checked and decoded.

func (items Items) Csv(of io.Writer, noReq, noResp bool) ([][]string, error)
    Csv emits items to CSV. Requests and responses may be omitted, if desired.

func (items Items) Go() string
    Go returns a Go-syntax representation of items.

func (items Items) Json(of io.Writer) error
    Json emits items to JSON.

type Request struct {
        Base64 string `xml:"base64,attr"` // Is Raw base64-encoded?
        Raw    string `xml:",chardata"`   // Raw HTTP request
        Body   string // May be base64-decoded later
}
    Request represents an XML <request> HTTP request.

func (r Request) FlatString() string
    FlatString returns a Request as a single-line string. Each element is
    comma-separated. If the body has been decoded, the body exclusively will be
    returned.

func (r Request) String() string
    String returns a prettified string representation of a Request. If the body
    has been decoded, the body exclusively will be returned.

func (r Request) ToStrings() []string
    ToStrings returns a Request as a slice of strings. Each element in the slice
    is a Request field. If the body has been decoded, the body exclusively will
    be returned.

type Response struct {
        Base64 string `xml:"base64,attr"` // Is Raw base64-encoded?
        Raw    string `xml:",chardata"`   // Raw HTTP response
        Body   string // May be base64-decoded later
}
    Response represents an XML <response> HTTP response.

func (r Response) FlatString() string
    FlatString returns a Response as a single-line string. Each element is
    comma-separated. If the body has been decoded, the body exclusively will be
    returned.

func (r Response) String() string
    String returns a prettified string representation of a Response. If the body
    has been decoded, the body exclusively will be returned.

func (r Response) ToStrings() []string
    ToStrings returns a Response as a slice of strings. Each element in the
    slice is a Response field. If the body has been decoded, the body
    exclusively will be returned.
```
