package gii

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

type TOC struct {
	Items []Item `xml:"item"`
}

type Root struct {
	Builddate int    `xml:"builddate,attr"` // XML attribute
	Doknr     string `xml:"doknr,attr"`     // XML attribute
	Norm      []Norm `xml:"norm"`           // XML element (array)
}
type Norm struct {
	Builddate int       `xml:"builddate,attr"` // XML attribute
	Doknr     string    `xml:"doknr,attr"`     // XML attribute
	Metadaten Metadaten `xml:"metadaten"`      // XML element
	Textdaten Textdaten `xml:"textdaten"`      // XML element
}
type Textdaten struct {
	Text      Text      `xml:"text"`      // XML element
	Fussnoten Fussnoten `xml:"fussnoten"` // XML element
}
type Fussnoten struct {
	Content Content `xml:"Content"` // XML element
}
type Text struct {
	Format  string  `xml:"format,attr"` // XML attribute
	Content Content `xml:"Content"`     // XML element
}
type Content struct {
	P []string `xml:"P"` // XML element
}
type Metadaten struct {
	Jurabk            string     `xml:"jurabk"`             // XML element
	AusfertigungDatum string     `xml:"ausfertigung-datum"` // XML element
	Fundstelle        Fundstelle `xml:"fundstelle"`         // XML element
	Langue            string     `xml:"langue"`             // XML element
	Enbez             string     `xml:"enbez"`              // XML element
	Titel             string     `xml:"titel"`              // XML element
}
type Fundstelle struct {
	Typ        string `xml:"typ,attr"`   // XML attribute
	Periodikum string `xml:"periodikum"` // XML element
	Zitstelle  string `xml:"zitstelle"`  // XML element
}

type TOCEntry struct {
	Title string
	Link  string
}

type GiiClient struct {
	httpclient  *http.Client
	baseAddress string
}

type GiiClientOption func(*GiiClient)

func NewGiiClient(httpClient *http.Client) *GiiClient {
	c := httpClient
	if c == nil {
		c = &http.Client{Timeout: 3 * time.Minute}
	}
	return &GiiClient{
		httpclient:  c,
		baseAddress: "",
	}
}

func DownloadTOC() []TOCEntry {
	xmlData, err := os.ReadFile("gii-toc.xml")
	if err != nil {
		fmt.Println("Error reading XML file:", err)
		return nil
	}

	// unmarshal the XML data into a slice of Item structs
	var toc TOC
	err = xml.Unmarshal(xmlData, &toc)
	if err != nil {
		fmt.Println("Error unmarshalling XML:", err)
		return nil
	}

	// work with first item
	item := toc.Items[0]

	// download the zip file from the link
	zipFile, err := os.Create("file.zip")
	if err != nil {
		fmt.Println("Error creating zip file:", err)
		return nil
	}
	defer zipFile.Close()

	// use http.Get to download the file
	resp, err := http.Get(item.Link)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return nil
	}
	defer resp.Body.Close()

	_, err = io.Copy(zipFile, resp.Body)
	if err != nil {
		fmt.Println("Error saving zip file:", err)
		return nil
	}
	// unzip the file and read the XML file inside
	Unzip("file.zip", "unzipped")

	// get first file in unzipped folder
	files, err := os.ReadDir("unzipped")
	if err != nil {
		fmt.Println("Error reading unzipped folder:", err)
		return nil
	}

	if len(files) == 0 {
		fmt.Println("No files found in unzipped folder")
		return nil
	}

	// get the first file in the unzipped folder
	firstFile := files[0]
	fmt.Println("First file in unzipped folder:", firstFile.Name())

	// unmarshal the XML file into a struct
	xmlFile, err := os.Open("unzipped/" + firstFile.Name())
	if err != nil {
		fmt.Println("Error opening XML file:", err)
		return nil
	}

	defer xmlFile.Close()

	xmlData, err = io.ReadAll(xmlFile)
	if err != nil {
		fmt.Println("Error reading XML file:", err)
		return nil
	}

	var root Root
	err = xml.Unmarshal(xmlData, &root)
	if err != nil {
		fmt.Println("Error unmarshalling XML:", err)
		return nil
	}

	var allChunks []StoredChunk
	chunkCfg := ChunkConfig{
		TargetTokens:  250,
		MaxTokens:     320,
		OverlapTokens: 40,
		MinTokens:     80,
	}

	for _, norm := range root.Norm {
		title := norm.Metadaten.Titel
		if title == "" {
			title = norm.Metadaten.Langue
		}
		if title == "" {
			continue
		}

		// build chunks from paragraphs
		chunks := BuildChunksFromParagraphs(norm.Doknr, title, norm.Textdaten.Text.Content.P, chunkCfg)
		if len(chunks) == 0 {
			continue
		}

		// embed all chunks
		chunks = EmbedChunks(context.Background(), embeddingClient, chunks)
		allChunks = append(allChunks, chunks...)
	}

	return allChunks
}
