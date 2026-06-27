package gii

// TOCItem represents a single item in the table of contents (TOC) of the GII XML data. The link field contains
// the URL to the corresponding zip file for that item.
type TOCItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

// TOC represents the table of contents structure for the GII XML data.
type TOC struct {
	Items []TOCItem `xml:"item"`
}

// Root represents the root structure of the GII XML data, containing metadata and text data for a specific legal norm within a singular XML file.
// There can be multiple XML files in a zip file.
type Root struct {
	Builddate int    `xml:"builddate,attr"`
	Doknr     string `xml:"doknr,attr"`
	Norm      []Norm `xml:"norm"`
}

type Norm struct {
	Builddate int       `xml:"builddate,attr"`
	Doknr     string    `xml:"doknr,attr"`
	Metadaten Metadaten `xml:"metadaten"`
	Textdaten Textdaten `xml:"textdaten"`
}

type Textdaten struct {
	Text      Text      `xml:"text"`
	Fussnoten Fussnoten `xml:"fussnoten"`
}

type Fussnoten struct {
	Content Content `xml:"Content"`
}

type Text struct {
	Format  string  `xml:"format,attr"`
	Content Content `xml:"Content"`
}

type Content struct {
	P []string `xml:"P"`
}

type Metadaten struct {
	Jurabk            string     `xml:"jurabk"`
	AusfertigungDatum string     `xml:"ausfertigung-datum"`
	Fundstelle        Fundstelle `xml:"fundstelle"`
	Langue            string     `xml:"langue"`
	Enbez             string     `xml:"enbez"`
	Titel             string     `xml:"titel"`
}

type Fundstelle struct {
	Typ        string `xml:"typ,attr"`
	Periodikum string `xml:"periodikum"`
	Zitstelle  string `xml:"zitstelle"`
}
