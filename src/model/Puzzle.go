package model

const (
	ACROSS = 'A'
	DOWN = 'D'
)

type Circle struct {
	Row int `xml:"Row,attr"`
	Col int `xml:"Col,attr"`
}

type Shade struct {
	Row   int   `xml:"Row,attr"`
	Col   int   `xml:"Col,attr"`
	Color string    `xml:"Color,attr"`
}

type Rebus struct {
	Row      int    `xml:"Row,attr"`
	Col      int    `xml:"Col,attr"`
	Short    string `xml:"Short,attr"`
	Expanded string `xml:",chardata"`
}

type Clue struct {
	Row       int    `xml:"Row,attr"`
	Col       int    `xml:"Col,attr"`
	Number    int    `xml:"Num,attr"`
	Direction string `xml:"Dir,attr"`
	Answer    string `xml:"Ans,attr"`
	Text      string `xml:",chardata"`
}

type Puzzle struct {
	Id        string `xml:"-"`
	Type      string
	Title     string
	Author    string
	Editor    string
	Copyright string
	Publisher string
	Date      string

	Width  int `xml:"Size>Cols"`
	Height int `xml:"Size>Rows"`

	Grid         []string `xml:"Grid>Row"`
	Circles      []Circle `xml:"Circles>Circle"`
	RebusEntries []Rebus  `xml:"RebusEntries>Rebus"`
	Shades       []Shade  `xml:"Shades>Shade"`
	Clues        []Clue   `xml:"Clues>Clue"`

	Notepad string
}

