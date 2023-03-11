package forms

import (
	"regexp"
	"time"

	"github.com/go-msvc/errors"
)

type Form struct {
	ID        string    `json:"id,omitempty" doc:"Unique ID assigned when the form is created"`
	Rev       int       `json:"rev,omitempty" doc:"Revision count form updates 1,2,3,..."`
	Timestamp time.Time `json:"timestamp" doc:"Time when the form revision was created"`
	Header
	Sections []Section `json:"sections,omitempty" doc:"Each section displays as another tab/page to be filled and user can navigate to next/prev."`
}

func (f *Form) Validate() error {
	if f.Rev < 0 {
		return errors.Errorf("negative rev:%d", f.Rev)
	}
	if err := f.Header.Validate(); err != nil {
		return errors.Wrapf(err, "invalid header")
	}
	for i, s := range f.Sections {
		if err := s.Validate(); err != nil {
			return errors.Wrapf(err, "invalid section[%d]", i)
		}
		s.FirstSection = false
	}
	if len(f.Sections) < 1 {
		return errors.Errorf("missing sections")
	}
	f.Sections[0].FirstSection = true
	return nil
} //Form.Validate()

//todo: add validation methods so users can download, edit and update their own forms without a graphical editor

type Section struct {
	Header
	FirstSection bool   `json:"first_section,omitempty"` //defined when validated - will override whatever you specified
	Name         string `json:"name"`
	Items        []Item `json:"items"`
}

func (s Section) Validate() error {
	if err := s.Header.Validate(); err != nil {
		return errors.Wrapf(err, "invalid header")
	}
	if s.Name == "" {
		return errors.Errorf("missing name")
	}
	if len(s.Items) < 1 {
		return errors.Errorf("missing items")
	}
	for i, item := range s.Items {
		if err := item.Validate(); err != nil {
			return errors.Wrapf(err, "invalid item[%d]", i)
		}
	}
	return nil
} //Section.Validate()

type Item struct {
	//exactly one of the following must be defined
	Header *Header `json:"header,omitempty" doc:"A header to seperate what's above from what below"`
	Image  *Image  `json:"image,omitempty"`
	Field  *Field  `json:"field,omitempty"`
	// List   *List   `json:"list" doc:"A list with one field, which can be repeated"`
	Table *Table `json:"table,omitempty" doc:"A list with multiple fields on each line displayed as columns"`
	Sub   *Sub   `json:"sub,omitempty" doc:"A sub section is another header with fields enclosed in a block. It supports ability for user to add more instances of the secion, e.g. if each section describes a person with several fields, of which one of more fields' values must be unique from other instances to create a unique key."`
}

func (i *Item) Validate() error {
	count := 0
	if i.Header != nil {
		count++
		if err := i.Header.Validate(); err != nil {
			return errors.Wrapf(err, "invalid header")
		}
	}
	if i.Image != nil {
		count++
		if err := i.Image.Validate(); err != nil {
			return errors.Wrapf(err, "invalid image")
		}
	}
	if i.Field != nil {
		count++
		if err := i.Field.Validate(); err != nil {
			return errors.Wrapf(err, "invalid field")
		}
	}
	if i.Table != nil {
		count++
		if err := i.Table.Validate(); err != nil {
			return errors.Wrapf(err, "invalid table")
		}
	}
	if i.Sub != nil {
		count++
		if err := i.Sub.Validate(); err != nil {
			return errors.Wrapf(err, "invalid sub")
		}
	}
	if count != 1 {
		return errors.Errorf("has %d of header|image|field|table|sub, should be exactly 1")
	}
	return nil
} //Item.Validate()

type Header struct {
	Title       string `json:"title,omitempty" doc:"Title is printed bigger than description"`
	Description string `json:"description,omitempty" doc:"Use markdown to style"`
}

func (h Header) Validate() error {
	if h.Title == "" {
		return errors.Errorf("missing title")
	}
	return nil
} //Header.Validate()

type Image struct {
	Header
	ID string `json:"id" doc:"Reference to uploaded image to display"`
	//todo: basic display options
}

func (i Image) Validate() error {
	if err := i.Header.Validate(); err != nil {
		return errors.Wrapf(err, "invalid header")
	}
	if i.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type Table struct {
	Header
	Name   string   `json:"name" doc:"Value is stored as this name which is unique in this form"`
	Min    int      `json:"min" doc:"Minimum nr of values required (0..50)"`
	Max    int      `json:"max" doc:"Maximum nr of values required (1..50)"`
	Uniq   []string `json:"uniq" doc:"Names of fields that must be unique"`
	Fields []Field  `json:"fields" doc:"One of more fields"`
}

func (t Table) Validate() error {
	if err := t.Header.Validate(); err != nil {
		return errors.Wrapf(err, "invalid header")
	}
	if t.Name == "" {
		return errors.Errorf("missing name")
	}
	if t.Min < 0 || t.Min > 50 {
		return errors.Errorf("min:%d must be 0..50", t.Min)
	}
	if t.Max < 1 || t.Max > 50 {
		return errors.Errorf("max:%d must be 1..50", t.Max)
	}
	if t.Min > t.Max {
		return errors.Errorf("min:%d > max:%d", t.Min, t.Max)
	}
	if len(t.Fields) < 1 {
		return errors.Errorf("missing fields")
	}
	for i, f := range t.Fields {
		if err := f.Validate(); err != nil {
			return errors.Wrapf(err, "invalid field[%d]", i)
		}
	}
	if len(t.Uniq) < 1 {
		return errors.Errorf("missing uniq")
	}
	for i, u := range t.Uniq {
		found := false
		for _, f := range t.Fields {
			if f.Name == u {
				found = true
				break
			}
		}
		if !found {
			return errors.Errorf("uniq[%d]=\"%s\" is not a field name", i, u)
		}
		for ii, uu := range t.Uniq {
			if i != i && u == uu {
				return errors.Errorf("uniq[%d]=\"%s\" duplicates uniq[%d]", i, u, ii)
			}
		}
	}
	return nil
}

type Sub struct {
	Header
	Name    string   `json:"name" doc:"Value is stored as this name which is unique in this form"`
	Min     int      `json:"min" doc:"Minimum nr of values required"`
	Max     int      `json:"max" doc:"Maximum nr of values required"`
	Section *Section `json:"section,omitempty"`
}

func (s *Sub) Validate() error {
	if err := s.Header.Validate(); err != nil {
		return errors.Wrapf(err, "invalid header")
	}
	if s.Name == "" {
		return errors.Errorf("missing name")
	}
	if s.Min < 0 || s.Min > 50 {
		return errors.Errorf("min:%d must be 0..50", s.Min)
	}
	if s.Max < 1 || s.Max > 50 {
		return errors.Errorf("max:%d must be 1..50", s.Max)
	}
	if s.Min > s.Max {
		return errors.Errorf("min:%d > max:%d", s.Min, s.Max)
	}
	if err := s.Section.Validate(); err != nil {
		return errors.Wrapf(err, "invalid section")
	}
	return nil
} //Sub.Validate()

type Field struct {
	Header
	Name      string     `json:"name" doc:"Value is stored as this name which is unique in this form"`
	Short     *Short     `json:"short,omitempty" doc:"Enter a short answer in one line"`
	Integer   *Integer   `json:"integer,omitempty" doc:"Integer value displayed as a slider or a up-down toggle or type it"`
	Number    *Number    `json:"number,omitempty" doc:"Enter a number which could have fractions"`
	Text      *Text      `json:"text,omitempty" doc:"Enter a multi-line response"`
	Date      *Date      `json:"date,omitempty" doc:"Enter/select a date in your local time zone"`
	Time      *Time      `json:"time,omitempty" doc:"Enter/select a time of day"`
	Duration  *Duration  `json:"duration,omitempty" doc:"Enter/select a duration of time"`
	Choice    *Choice    `json:"choice,omitempty" doc:"Select one from a list. Display as radio button or drop down"`
	Selection *Selection `json:"selection,omitempty" doc:"Select multiple options. Displayed as check boxes"`
	// Grid coice (choices repeats for each row)
	// Grid check (check repeats for each row)
	// ...
}

func (f *Field) Validate() error {
	if err := f.Header.Validate(); err != nil {
		return errors.Wrapf(err, "invalid header")
	}
	if f.Name == "" {
		return errors.Errorf("missing name")
	}
	count := 0
	if f.Short != nil {
		count++
		if err := f.Short.Validate(); err != nil {
			return errors.Wrapf(err, "invalid short")
		}
	}
	if f.Integer != nil {
		count++
		if err := f.Integer.Validate(); err != nil {
			return errors.Wrapf(err, "invalid integer")
		}
	}
	if f.Number != nil {
		count++
		if err := f.Number.Validate(); err != nil {
			return errors.Wrapf(err, "invalid number")
		}
	}
	if f.Text != nil {
		count++
		if err := f.Text.Validate(); err != nil {
			return errors.Wrapf(err, "invalid text")
		}
	}
	if f.Date != nil {
		count++
		if err := f.Date.Validate(); err != nil {
			return errors.Wrapf(err, "invalid date")
		}
	}
	if f.Time != nil {
		count++
		if err := f.Time.Validate(); err != nil {
			return errors.Wrapf(err, "invalid time")
		}
	}
	if f.Duration != nil {
		count++
		if err := f.Duration.Validate(); err != nil {
			return errors.Wrapf(err, "invalid duration")
		}
	}
	if f.Choice != nil {
		count++
		if err := f.Choice.Validate(); err != nil {
			return errors.Wrapf(err, "invalid choice")
		}
	}
	if f.Selection != nil {
		count++
		if err := f.Selection.Validate(); err != nil {
			return errors.Wrapf(err, "invalid selection")
		}
	}
	if count != 1 {
		return errors.Errorf("has %d of short|integer|number|text|date|time|duration|choice|selection instead of 1", count)
	}
	return nil
} //Field.Validate()

// todo: add validation and display options to each of these
type Short struct {
	MinLen *int    `json:"min_length,omitempty"`
	MaxLen *int    `json:"max_length,omitempty"`
	Regex  *string `json:"regex,omitempty"`
}

func (s Short) Validate() error {
	if s.MinLen != nil && *s.MinLen < 0 {
		return errors.Errorf("min_length:%d < 0", *s.MinLen)
	}
	if s.MaxLen != nil && *s.MaxLen < 0 {
		return errors.Errorf("max_length:%d < 0", *s.MaxLen)
	}
	if s.MinLen != nil && s.MaxLen != nil && (*s.MinLen > *s.MaxLen) {
		return errors.Errorf("min_length:%d > max_lengh:%d", *s.MinLen, *s.MaxLen)
	}
	if s.Regex != nil {
		if _, err := regexp.Compile(*s.Regex); err != nil {
			return errors.Errorf("invalid regex:\"%s\"", *s.Regex)
		}
	}
	return nil
} //Short.Validate()

type Integer struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

func (i Integer) Validate() error {
	if i.Min != nil && i.Max != nil && (*i.Min > *i.Max) {
		return errors.Errorf("min:%d > max:%d", *i.Min, *i.Max)
	}
	return nil
} //Integer.Validate()

type Number struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

func (i Number) Validate() error {
	if i.Min != nil && i.Max != nil && (*i.Min > *i.Max) {
		return errors.Errorf("min:%d > max:%d", *i.Min, *i.Max)
	}
	return nil
} //Number.Validate()

type Text struct {
	MinLen *int `json:"min_length,omitempty"`
	MaxLen *int `json:"max_length,omitempty"`
	NrRows *int `json:"nr_rows,omitempty" doc:"Nr of rows to display in the form. Scrolling allowed to edit more. Valid values = 2..20."`
	//todo: display options: nr rows / cols /maxlen
}

func (s Text) Validate() error {
	if s.MinLen != nil && *s.MinLen < 0 {
		return errors.Errorf("min_length:%d < 0", *s.MinLen)
	}
	if s.MaxLen != nil && *s.MaxLen < 0 {
		return errors.Errorf("max_length:%d < 0", *s.MaxLen)
	}
	if s.MinLen != nil && s.MaxLen != nil && (*s.MinLen > *s.MaxLen) {
		return errors.Errorf("min_length:%d > max_lengh:%d", *s.MinLen, *s.MaxLen)
	}
	if s.NrRows != nil && *s.NrRows < 2 || *s.NrRows > 20 {
		return errors.Errorf("nr_rows:%d is not 2..20")
	}
	return nil
} //Text.Validate()

type Date struct {
	Min *string `json:"min" doc:"Optional minimum date CCYY-MM-DD"`
	Max *string `json:"max" doc:"Optional maximum date CCYY-MM-DD"`
}

func (d Date) Validate() error {
	var min time.Time
	var max time.Time
	if d.Min != nil {
		var err error
		if min, err = time.Parse("2006-01-02", *d.Min); err != nil {
			return errors.Errorf("min:\"%s\" is not CCYY-MM-DD", *d.Min)
		}
	}
	if d.Max != nil {
		var err error
		if max, err = time.Parse("2006-01-02", *d.Max); err != nil {
			return errors.Errorf("max:\"%s\" is not CCYY-MM-DD", *d.Max)
		}
	}
	if d.Min != nil && d.Max != nil {
		if min.After(max) {
			return errors.Errorf("min:\"%s\" is after max:\"%s\"", *d.Min, *d.Max)
		}
	}
	return nil
} //Date.Validate()

type Time struct {
	Min *string `json:"min" doc:"Optional minimum date HH:MM"`
	Max *string `json:"max" doc:"Optional maximum date HH:MM"`
}

func (d Time) Validate() error {
	var min time.Time
	var max time.Time
	if d.Min != nil {
		var err error
		if min, err = time.Parse("15:04`", *d.Min); err != nil {
			return errors.Errorf("min:\"%s\" is not HH:MM", *d.Min)
		}
	}
	if d.Max != nil {
		var err error
		if max, err = time.Parse("15:04", *d.Max); err != nil {
			return errors.Errorf("max:\"%s\" is not HH:MM", *d.Max)
		}
	}
	if d.Min != nil && d.Max != nil {
		if min.After(max) {
			return errors.Errorf("min:\"%s\" is after max:\"%s\"", *d.Min, *d.Max)
		}
	}
	return nil
} //Time.Validate()

type Duration struct {
	//todo: should have options like 15min or 1yr granularity...
	// Min *time.Duration
	// Max *time.Duration
	//todo: and display should be drop down list of values to select from... but also allow flexible text input like 2h3min
	//for not just text input without validation and stored like that...
}

func (s Duration) Validate() error {
	return nil
} //Duration.Validate()

type Choice struct {
	Options []Option `json:"options"`
}

func (s Choice) Validate() error {
	if len(s.Options) < 1 {
		return errors.Errorf("missing options")
	}
	for i, o := range s.Options {
		if err := o.Validate(); err != nil {
			return errors.Wrapf(err, "invalid option[%d]", i)
		}
	}
	return nil
} //Choide.Validate()

type Selection struct {
	Options []Option `json:"options"`
}

func (s Selection) Validate() error {
	if len(s.Options) < 1 {
		return errors.Errorf("missing options")
	}
	for i, o := range s.Options {
		if err := o.Validate(); err != nil {
			return errors.Wrapf(err, "invalid option[%d]", i)
		}
	}
	return nil
} //Selection.Validate()

type Option struct {
	Header
	Value string `json:"value" doc:"Stored value when selected"`
}

func (o Option) Validate() error {
	if err := o.Header.Validate(); err != nil {
		return err
	}
	if o.Value == "" {
		return errors.Errorf("missing value")
	}
	return nil
} //Option.Validate()
