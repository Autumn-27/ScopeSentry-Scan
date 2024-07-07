// Copyright (c) 2020, Microsoft Corporation, Sean Hinchee
// Licensed under the MIT License.

package burpxml

import (
	"fmt"
)

/* Slice of string routines */

// ToStrings returns an Item as a slice of strings.
// Each element in the slice is an Item field.
func (i Item) ToStrings(noReq, noResp bool) []string {
	arr := []string{
		i.Time,
		i.Url,
		i.Host.Name,
		i.Host.Ip,
		i.Port,
		i.Protocol,
		i.Path,
		i.Extension,
	}

	if !noReq {
		arr = append(arr, i.Request.ToStrings()...)
	}

	arr = append(arr, []string{
		i.Status,
		i.ResponseLength,
		i.MimeType,
	}...)

	if !noResp {
		arr = append(arr, i.Response.ToStrings()...)
	}

	arr = append(arr, []string{
		i.Comment,
	}...)

	return arr
}

// ToStrings returns a Request as a slice of strings.
// Each element in the slice is a Request field.
// If the body has been decoded, the body exclusively will be returned.
func (r Request) ToStrings() []string {
	// We have decoded the body
	if r.Body != "" {
		return []string{r.Body}
	}

	return []string{r.Base64, r.Raw}
}

// ToStrings returns a Response as a slice of strings.
// Each element in the slice is a Response field.
// If the body has been decoded, the body exclusively will be returned.
func (r Response) ToStrings() []string {
	// We have decoded the body
	if r.Body != "" {
		return []string{r.Body}
	}

	return []string{r.Base64, r.Raw}
}

/* Flattened string formatting routines */

// FlatString returns an Item as a single-line string.
// Each element is comma-separated.
func (i Item) FlatString() string {
	return fmt.Sprintf(`%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s`,
		i.Time,
		i.Url,
		i.Host.Name,
		i.Host.Ip,
		i.Port,
		i.Protocol,
		i.Path,
		i.Extension,
		i.Request.FlatString(),
		i.Status,
		i.ResponseLength,
		i.MimeType,
		i.Response.FlatString(),
		i.Comment,
	)
}

// FlatString returns a Request as a single-line string.
// Each element is comma-separated.
// If the body has been decoded, the body exclusively will be returned.
func (r Request) FlatString() string {
	s := ""

	// We have decoded the body
	if r.Body != "" {
		s += fmt.Sprintf(`%s`, r.Body)
	} else {
		s += fmt.Sprintf(`%s,%s`, r.Base64, r.Raw)
	}

	return s
}

// FlatString returns a Response as a single-line string.
// Each element is comma-separated.
// If the body has been decoded, the body exclusively will be returned.
func (r Response) FlatString() string {
	s := ""

	// We have decoded the body
	if r.Body != "" {
		s += fmt.Sprintf(`%s`, r.Body)
	} else {
		s += fmt.Sprintf(`%s,%s`, r.Base64, r.Raw)
	}

	return s
}

/* String formatting routines */

// String returns a prettified string representation of an Item.
func (i Item) String() string {
	return fmt.Sprintf(`Item{
	Time	=	%s,
	Url		=	%s,
	Host	=	%s,
	IP		=	%s,
	Port	=	%s,
	Proto	=	%s,
	Path	=	%s,
	Ext		=	%s,
	%s,
	Status	=	%s,
	RespLen	=	%s,
	MIME	=	%s,
	%s,
	Comment	=	%s,
}`,
		i.Time,
		i.Url,
		i.Host.Name,
		i.Host.Ip,
		i.Port,
		i.Protocol,
		i.Path,
		i.Extension,
		i.Request.String(),
		i.Status,
		i.ResponseLength,
		i.MimeType,
		i.Response.String(),
		i.Comment,
	)
}

// String returns a prettified string representation of a Request.
// If the body has been decoded, the body exclusively will be returned.
func (r Request) String() string {
	s := "Request{\n"

	// We have decoded the body
	if r.Body != "" {
		s += fmt.Sprintf(`	Body = %s,\n`, r.Body)
	} else {
		s += fmt.Sprintf(`	Base64	=	%s,
Body	=	%s,\n`, r.Base64, r.Raw)
	}

	s += "}"

	return s
}

// String returns a prettified string representation of a Response.
// If the body has been decoded, the body exclusively will be returned.
func (r Response) String() string {
	s := "Response{\n"

	// We have decoded the body
	if r.Body != "" {
		s += fmt.Sprintf(`	Body = %s,\n`, r.Body)
	} else {
		s += fmt.Sprintf(`	Base64	=	%s,
Body	=	%s,\n`, r.Base64, r.Raw)
	}

	s += "}"

	return s
}
