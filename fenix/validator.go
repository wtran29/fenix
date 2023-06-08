package fenix

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type Validation struct {
	Data   url.Values
	Errors map[string]string
}

func (f *Fenix) Validator(data url.Values) *Validation {
	return &Validation{
		Errors: make(map[string]string),
		Data:   data,
	}
}

func (v *Validation) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validation) AddError(key, msg string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = msg
	}
}

func (v *Validation) Has(field string, r *http.Request) bool {
	x := r.Form.Get(field)
	return x != ""
}

func (v *Validation) Required(r *http.Request, fields ...string) {
	for _, field := range fields {
		value := r.Form.Get(field)
		if strings.TrimSpace(value) == "" {
			v.AddError(field, "This field cannot be blank")
		}
	}
}

func (v *Validation) Check(ok bool, key, msg string) {
	if !ok {
		v.AddError(key, msg)
	}
}

func (v *Validation) IsEmail(field, val string) {
	if !govalidator.IsEmail(val) {
		v.AddError(field, "Invalid email address")
	}
}

func (v *Validation) IsInt(field, val string) {
	_, err := strconv.Atoi(val)
	if err != nil {
		v.AddError(field, "This field must be an integer")
	}
}

func (v *Validation) IsFloat(field, val string) {
	_, err := strconv.ParseFloat(val, 64)
	if err != nil {
		v.AddError(field, "This field must be a float number")
	}
}

func (v *Validation) IsDateISO(field, val string) {
	_, err := time.Parse("2006-01-02", val)
	if err != nil {
		v.AddError(field, "This field must be a date in the form of YYYY-MM-DD")
	}
}

func (v *Validation) IsDateMMDDYYYY(field, val string) {
	_, err := time.Parse("01-02-2006", val)
	if err != nil {
		v.AddError(field, "This field must be a date in the form of MM-DD-YYYY")
	}
}

func (v *Validation) NoSpaces(field, val string) {
	if govalidator.HasWhitespace(val) {
		v.AddError(field, "White spaces are not permitted")
	}
}
