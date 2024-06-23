package utilsModel

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"errors"
	"time"

	"github.com/dghubble/oauth1"
)

type CustomTime struct {
	time.Time
}

func (c *CustomTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	parse, err := time.Parse("2006-01-02 15:04:05", v)
	if err != nil {
		return err
	}
	*c = CustomTime{parse}
	return nil
}

// Scan para leer el tiempo desde la base de datos
func (c *CustomTime) Scan(value interface{}) error {
	if value == nil {
		*c = CustomTime{Time: time.Time{}}
		return nil
	}
	v, ok := value.(time.Time)
	if !ok {
		return errors.New("invalid type for CustomTime")
	}
	*c = CustomTime{v}
	return nil
}

// Value para escribir el tiempo en la base de datos
func (c CustomTime) Value() (driver.Value, error) {
	return c.Time, nil
}

// JSONType is a generic slice type for JSON serialization
type JSONType[T any] []T

// Scan implements the sql.Scanner interface for JSONType
func (j *JSONType[T]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for JSONType")
	}
	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface for JSONType
func (j JSONType[T]) Value() (driver.Value, error) {
	return json.Marshal(j)
}

type Token struct {
	oauth1.Token `json:"-"`
}

func (t Token) Value() (driver.Value, error) {
	if t.Token.Token == "" {
		return "", nil
	}
	tokenBytes, err := json.Marshal(t.Token)
	if err != nil {
		return nil, err
	}
	return string(tokenBytes), nil
}

func (t *Token) Scan(value interface{}) error {
	if value == nil {
		*t = Token{}
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return errors.New("type assertion to string failed")
	}

	return json.Unmarshal([]byte(str), &t.Token)
}
