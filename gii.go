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

// DownloadTOCItems downloads the table of contents (TOC) items from the gesetze-im-internet.de website,  and returns them as a slice of strings.
func (giiClient *GiiClient) DownloadTOCItems(ctx context.Context, toc *TOC) ([]*Root, error) {

	if toc == nil || len(toc.Items) == 0 {
		return nil, fmt.Errorf("TOC is empty")
	}

	var roots []*Root

	for _, item := range toc.Items {
		rootItems, err := giiClient.DownloadTOCItem(ctx, item)
		if err != nil {
			return nil, err
		}

		roots = append(roots, rootItems...)
	}

	return roots, nil
}

func (giiClient *GiiClient) DownloadTOCItem(ctx context.Context, item TOCItem) ([]*Root, error) {
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
		root, err := giiClient.ReadXMLRootFile(ctx, file)
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}

	return roots, nil
}

func (giiClient *GiiClient) ReadXMLRootFile(ctx context.Context, file *zip.File) (*Root, error) {
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

// 	// work with first item
// 	item := toc.Items[0]

// 	// download the zip file from the link
// 	zipFile, err := os.Create("file.zip")
// 	if err != nil {
// 		fmt.Println("Error creating zip file:", err)
// 		return nil
// 	}
// 	defer zipFile.Close()

// 	// use http.Get to download the file
// 	resp, err := http.Get(item.Link)
// 	if err != nil {
// 		fmt.Println("Error downloading file:", err)
// 		return nil
// 	}
// 	defer resp.Body.Close()

// 	_, err = io.Copy(zipFile, resp.Body)
// 	if err != nil {
// 		fmt.Println("Error saving zip file:", err)
// 		return nil
// 	}
// 	// unzip the file and read the XML file inside
// 	Unzip("file.zip", "unzipped")

// 	// get first file in unzipped folder
// 	files, err := os.ReadDir("unzipped")
// 	if err != nil {
// 		fmt.Println("Error reading unzipped folder:", err)
// 		return nil
// 	}

// 	if len(files) == 0 {
// 		fmt.Println("No files found in unzipped folder")
// 		return nil
// 	}

// 	// get the first file in the unzipped folder
// 	firstFile := files[0]
// 	fmt.Println("First file in unzipped folder:", firstFile.Name())

// 	// unmarshal the XML file into a struct
// 	xmlFile, err := os.Open("unzipped/" + firstFile.Name())
// 	if err != nil {
// 		fmt.Println("Error opening XML file:", err)
// 		return nil
// 	}

// 	defer xmlFile.Close()

// 	xmlData, err = io.ReadAll(xmlFile)
// 	if err != nil {
// 		fmt.Println("Error reading XML file:", err)
// 		return nil
// 	}

// 	var root Root
// 	err = xml.Unmarshal(xmlData, &root)
// 	if err != nil {
// 		fmt.Println("Error unmarshalling XML:", err)
// 		return nil
// 	}

// 	var allChunks []StoredChunk
// 	chunkCfg := ChunkConfig{
// 		TargetTokens:  250,
// 		MaxTokens:     320,
// 		OverlapTokens: 40,
// 		MinTokens:     80,
// 	}

// 	for _, norm := range root.Norm {
// 		title := norm.Metadaten.Titel
// 		if title == "" {
// 			title = norm.Metadaten.Langue
// 		}
// 		if title == "" {
// 			continue
// 		}

// 		// build chunks from paragraphs
// 		chunks := BuildChunksFromParagraphs(norm.Doknr, title, norm.Textdaten.Text.Content.P, chunkCfg)
// 		if len(chunks) == 0 {
// 			continue
// 		}

// 		// embed all chunks
// 		chunks = EmbedChunks(context.Background(), embeddingClient, chunks)
// 		allChunks = append(allChunks, chunks...)
// 	}

// 	return allChunks
// }
