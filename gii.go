package gii

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GiiClient struct {
	httpclient  *http.Client
	baseAddress string
}

func NewGiiClient(opts ...GiiClientOption) *GiiClient {
	client := &GiiClient{
		httpclient:  &http.Client{Timeout: 3 * time.Minute},
		baseAddress: "https://www.gesetze-im-internet.de",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}

	return client
}

// DownloadTOC downloads the table of contents (TOC) from the gesetze-im-internet.de website and returns it as a *TOC struct.
func (giiClient *GiiClient) DownloadTOC(ctx context.Context) (*TOC, error) {
	tocURL := giiClient.baseAddress + tocPath
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, tocURL, nil)
	resp, err := giiClient.httpclient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	xmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var toc TOC
	err = xml.Unmarshal(xmlData, &toc)

	return &toc, err
}

// DownloadTOCItems downloads the table of contents (TOC) items and returns them as a slice of *Root.
func (giiClient *GiiClient) DownloadTOCItems(ctx context.Context, toc *TOC) ([]*Root, error) {

	if toc == nil || len(toc.Items) == 0 {
		return nil, fmt.Errorf("TOC is empty")
	}

	var roots []*Root

	for _, item := range toc.Items {
		rootItems, err := giiClient.downloadTOCItem(ctx, item)
		if err != nil {
			return nil, err
		}

		roots = append(roots, rootItems...)
	}

	return roots, nil
}

func (giiClient *GiiClient) downloadTOCItem(ctx context.Context, item TOCItem) ([]*Root, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	var roots []*Root
	// 1. download the zip file from the link
	zipReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, item.Link, nil)
	zipResp, err := giiClient.httpclient.Do(zipReq)
	if err != nil {
		return nil, err
	}
	defer zipResp.Body.Close()

	files, err := Unzip(zipResp.Body)
	if err != nil {
		return nil, err
	}

	// 2. find the XML files in the zip
	for _, file := range files {
		root, err := giiClient.readXMLRootFile(ctx, file)
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}

	return roots, nil
}

func (giiClient *GiiClient) readXMLRootFile(ctx context.Context, file *zip.File) (*Root, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if file.FileInfo().IsDir() {
		return nil, fmt.Errorf("file is a directory: %s", file.Name)
	}

	if !strings.HasSuffix(file.Name, ".xml") {
		return nil, fmt.Errorf("file is not an XML file: %s", file.Name)
	}

	xmlFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	xmlData, err := io.ReadAll(xmlFile)
	if err != nil {
		return nil, err
	}

	var root Root
	err = xml.Unmarshal(xmlData, &root)
	if err != nil {
		return nil, err
	}

	return &root, nil
}
