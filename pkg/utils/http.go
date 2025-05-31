// utils-------------------------------------
// @file      : http.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/5/27 22:06
// -------------------------------------------

package utils

//
//import (
//	"bytes"
//	"crypto/tls"
//	"fmt"
//	"io"
//	"net/http"
//	"strings"
//	"time"
//
//	"github.com/projectdiscovery/gologger"
//	"github.com/projectdiscovery/gologger/levels"
//)
//
//type HttpResponse struct {
//	Url           string
//	StatusCode    int
//	Headers       map[string]string
//	Body          []byte
//	ContentLength int
//	Redirect      string
//}
//
//type request struct {
//	client *http.Client
//}
//
//var HttpRequests *request
//
//func InitializeRequests() {
//	gologger.DefaultLogger.SetMaxLevel(levels.LevelWarning)
//
//	tr := &http.Transport{
//		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
//		MaxIdleConns:    100,
//		IdleConnTimeout: 10 * time.Second,
//	}
//	client := &http.Client{
//		Timeout:   10 * time.Second,
//		Transport: tr,
//	}
//	Requests = &request{client: client}
//
//	paralyze, err := wappalyzergo.New()
//	if err != nil {
//		fmt.Printf("init wappalyzer error: %v", err)
//	}
//	Wappalyzer = paralyze
//}
//
//func (r *request) HttpGet(uri string) (HttpResponse, error) {
//	req, _ := http.NewRequest(http.MethodGet, uri, nil)
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return HttpResponse{}, err
//	}
//	defer resp.Body.Close()
//
//	body, _ := io.ReadAll(resp.Body)
//	headers := make(map[string]string)
//	for k, v := range resp.Header {
//		headers[k] = strings.Join(v, ", ")
//	}
//
//	redirect := resp.Header.Get("Location")
//	if redirect == "" {
//		redirect = resp.Header.Get("location")
//	}
//
//	return HttpResponse{
//		Url:           uri,
//		StatusCode:    resp.StatusCode,
//		Headers:       headers,
//		Body:          body,
//		ContentLength: len(body),
//		Redirect:      redirect,
//	}, nil
//}
//
//func (r *request) HttpGetByte(uri string) ([]byte, error) {
//	req, _ := http.NewRequest(http.MethodGet, uri, nil)
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//	return io.ReadAll(resp.Body)
//}
//
//func (r *request) HttpPost(uri string, requestBody []byte, ct string) (error, HttpResponse) {
//	req, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(requestBody))
//	if ct == "json" {
//		req.Header.Set("Content-Type", "application/json")
//	}
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return err, HttpResponse{}
//	}
//	defer resp.Body.Close()
//
//	body, _ := io.ReadAll(resp.Body)
//	headers := make(map[string]string)
//	for k, v := range resp.Header {
//		headers[k] = strings.Join(v, ", ")
//	}
//	return nil, HttpResponse{
//		StatusCode: resp.StatusCode,
//		Headers:    headers,
//		Body:       body,
//	}, nil
//}
//
//func (r *request) HttpGetWithCustomHeader(uri string, customHeaders []string) (HttpResponse, error) {
//	req, _ := http.NewRequest(http.MethodGet, uri, nil)
//	for _, header := range customHeaders {
//		parts := strings.SplitN(header, ":", 2)
//		if len(parts) == 2 {
//			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
//		}
//	}
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return HttpResponse{}, err
//	}
//	defer resp.Body.Close()
//
//	body, _ := io.ReadAll(resp.Body)
//	const maxBodySize = 4 * 1024 * 1024
//	if len(body) > maxBodySize {
//		body = body[:maxBodySize]
//	}
//
//	headers := make(map[string]string)
//	for k, v := range resp.Header {
//		headers[k] = strings.Join(v, ", ")
//	}
//
//	redirect := resp.Header.Get("Location")
//	if redirect == "" {
//		redirect = resp.Header.Get("location")
//	}
//
//	return HttpResponse{
//		Url:           uri,
//		StatusCode:    resp.StatusCode,
//		Headers:       headers,
//		Body:          body,
//		ContentLength: len(body),
//		Redirect:      redirect,
//	}, nil
//}
//
//func (r *request) HttpGetByteWithCustomHeader(uri string, customHeaders []string) ([]byte, error) {
//	req, _ := http.NewRequest(http.MethodGet, uri, nil)
//	for _, header := range customHeaders {
//		parts := strings.SplitN(header, ":", 2)
//		if len(parts) == 2 {
//			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
//		}
//	}
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//	return io.ReadAll(resp.Body)
//}
//
//func (r *request) HttpPostWithCustomHeader(uri string, requestBody []byte, ct string, customHeaders []string) (error, HttpResponse) {
//	req, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(requestBody))
//	for _, header := range customHeaders {
//		parts := strings.SplitN(header, ":", 2)
//		if len(parts) == 2 {
//			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
//		}
//	}
//	if ct == "json" {
//		req.Header.Set("Content-Type", "application/json")
//	}
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return err, HttpResponse{}
//	}
//	defer resp.Body.Close()
//
//	body, _ := io.ReadAll(resp.Body)
//	headers := make(map[string]string)
//	for k, v := range resp.Header {
//		headers[k] = strings.Join(v, ", ")
//	}
//	return nil, HttpResponse{
//		StatusCode: resp.StatusCode,
//		Headers:    headers,
//		Body:       body,
//	}, nil
//}
//
//func (r *request) HttpGetNoRes(uri string) error {
//	req, _ := http.NewRequest(http.MethodGet, uri, nil)
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return err
//	}
//	// 不读取内容，但释放连接
//	io.Copy(io.Discard, resp.Body)
//	resp.Body.Close()
//	return nil
//}
//
//func (r *request) HttpPostNoRes(uri string, requestBody []byte, ct string) error {
//	req, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(requestBody))
//	if ct == "json" {
//		req.Header.Set("Content-Type", "application/json")
//	}
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return err
//	}
//	io.Copy(io.Discard, resp.Body)
//	resp.Body.Close()
//	return nil
//}
//
//func (r *request) HttpGetNoResWithCustomHeader(uri string, customHeaders []string) error {
//	req, _ := http.NewRequest(http.MethodGet, uri, nil)
//	for _, header := range customHeaders {
//		parts := strings.SplitN(header, ":", 2)
//		if len(parts) == 2 {
//			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
//		}
//	}
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return err
//	}
//	io.Copy(io.Discard, resp.Body)
//	resp.Body.Close()
//	return nil
//}
//
//func (r *request) HttpPostNoResWithCustomHeader(uri string, requestBody []byte, ct string, customHeaders []string) error {
//	req, _ := http.NewRequest(http.MethodPost, uri, bytes.NewReader(requestBody))
//	for _, header := range customHeaders {
//		parts := strings.SplitN(header, ":", 2)
//		if len(parts) == 2 {
//			req.Header.Set(parts[0], strings.TrimSpace(parts[1]))
//		}
//	}
//	if ct == "json" {
//		req.Header.Set("Content-Type", "application/json")
//	}
//	resp, err := r.client.Do(req)
//	if err != nil {
//		return err
//	}
//	io.Copy(io.Discard, resp.Body)
//	resp.Body.Close()
//	return nil
//}
