package forms

type Form struct {
	ID       string    `json:"id"`
	Header             //Header    `json:"header"`
	Sections []Section `json:"sections,omitempty" doc:"Each section displays as another tab/page to be filled and user can navigate to next/prev."`
}

//todo: add validation methods so users can download, edit and update their own forms without a graphical editor

type Section struct {
	Header       //Header `json:"header"`
	FirstSection bool
	Name         string `json:"name"`
	Items        []Item `json:"items"`
}

type Item struct {
	//exactly one of the following must be defined
	Header *Header `json:"header,omitempty" doc:"A header to seperate what's above from what below"`
	Image  *Image  `json:"image,omitempty"`
	Field  *Field  `json:"field,omitempty"`
	// List   *List   `json:"list" doc:"A list with one field, which can be repeated"`
	Table *Table `json:"table,omitempty" doc:"A list with multiple fields on each line displayed as columns"`
	Sub   *Sub   `json:"sub,omitempty" doc:"A sub section is another header with fields enclosed in a block. It supports ability for user to add more instances of the secion, e.g. if each section describes a person with several fields, of which one of more fields' values must be unique from other instances to create a unique key."`
}

type Header struct {
	Title       string `json:"title,omitempty" doc:"Title is printed bigger than description"`
	Description string `json:"description,omitempty" doc:"Use markdown to style"`
}

type Image struct {
	Header
	ID string `json:"id" doc:"Reference to uploaded image to display"`
	//todo: basic display options
}

//List commented out - see if can do with table having one field...
// type List struct {
// 	Header
// 	Name  string `json:"name" doc:"Value is stored as this name which is unique in this form"`
// 	Min   int    `json:"min"`
// 	Max   int    `json:"max"`
// 	Field Field  `json:"field"`
// }

type Table struct {
	Header
	Name   string   `json:"name" doc:"Value is stored as this name which is unique in this form"`
	Min    int      `json:"min" doc:"Minimum nr of values required"`
	Max    int      `json:"max" doc:"Maximum nr of values required"`
	Uniq   []string `json:"uniq" doc:"Names of fields that must be unique"`
	Fields []Field  `json:"fields" doc:"One of more fields"`
}

type Sub struct {
	Header
	Name    string  `json:"name" doc:"Value is stored as this name which is unique in this form"`
	Min     int     `json:"min" doc:"Minimum nr of values required"`
	Max     int     `json:"max" doc:"Maximum nr of values required"`
	Section Section `json:"section"`
}

type Field struct {
	Header
	Name string `json:"name" doc:"Value is stored as this name which is unique in this form"`
	//one of the following must be defined
	Short     *Short     `json:"short" doc:"Enter a short answer in one line"`
	Integer   *Integer   `json:"integer" doc:"Integer value displayed as a slider or a up-down toggle or type it"`
	Number    *Number    `json:"number" doc:"Enter a number which could have fractions"`
	Text      *Text      `json:"text" doc:"Enter a multi-line response"`
	Date      *Date      `json:"date" doc:"Enter/select a date in your local time zone"`
	Time      *Time      `json:"time" doc:"Enter/select a time of day"`
	Duration  *Duration  `json:"duration" doc:"Enter/select a duration of time"`
	Choice    *Choice    `json:"choice" doc:"Select one from a list. Display as radio button or drop down"`
	Selection *Selection `json:"selection" doc:"Select multiple options. Displayed as check boxes"`
	// Grid coice (choices repeats for each row)
	// Grid check (check repeats for each row)
	// ...
}

// todo: add validation and display options to each of these
type Short struct {
}

type Integer struct {
}

type Number struct {
}

type Text struct {
}

type Date struct{}

type Time struct{}

type Duration struct{}

type Choice struct {
	Options []Option `json:"options"`
}

type Selection struct {
	Options []Option `json:"options"`
}

type Option struct {
	Title string `json:"title"`
	Value string `json:"value" doc:"Stored value when selected"`
}
