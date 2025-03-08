// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gmodel

import (
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

type Address struct {
	ID          int64          `json:"id" sql:"primary_key"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Latitude    float64        `json:"latitude"`
	Longitude   float64        `json:"longitude"`
	Distance    *float64       `json:"distance,omitempty" alias:"address.distance"`
	MapsLink    string         `json:"mapsLink"`
	FullAddress string         `json:"fullAddress"`
	CountryCode string         `json:"countryCode"`
	Country     *string        `json:"country,omitempty" alias:"country.name"`
	CreatedByID *int64         `json:"createdById,omitempty"`
	CreatedBy   *CreatedByUser `json:"createdBy,omitempty"`
	UpdatedByID *int64         `json:"updatedById,omitempty"`
	UpdatedBy   *UpdatedByUser `json:"updatedBy,omitempty"`
}

type AdministrativeDivision struct {
	Name   string `json:"name" alias:"administrative_division.administrative_division"`
	Cities string `json:"cities"`
}

type Auth struct {
	Token     string `json:"token"`
	User      *User  `json:"user"`
	IsNewUser *bool  `json:"isNewUser,omitempty"`
}

type Country struct {
	Code                    string                    `json:"code" sql:"primary_key"`
	Name                    string                    `json:"name"`
	AdministrativeDivisions []*AdministrativeDivision `json:"administrativeDivisions"`
	Currency                *Currency                 `json:"currency,omitempty"`
	CallingCode             *string                   `json:"callingCode,omitempty"`
	Language                *string                   `json:"language,omitempty"`
}

type CreateAccountInput struct {
	Email       string  `json:"email" validate:"required,email"`
	PhoneNumber *string `json:"phoneNumber,omitempty" validate:"omitempty,e164"`
	Name        string  `json:"name" validate:"required"`
	Password    string  `json:"password" validate:"required"`
}

type CreateAddress struct {
	Latitude               float64 `json:"latitude" validate:"required,latitude"`
	Longitude              float64 `json:"longitude" validate:"required,longitude"`
	MapsLink               string  `json:"mapsLink" validate:"required,http_url"`
	FullAddress            string  `json:"fullAddress" validate:"required,contains"`
	City                   string  `json:"city" validate:"required"`
	AdministrativeDivision string  `json:"administrativeDivision" validate:"required"`
	CountryCode            string  `json:"countryCode" validate:"iso3166_1_alpha2"`
}

type CreateProduct struct {
	Name                 string   `json:"name"`
	Image                string   `json:"image"`
	Description          string   `json:"description"`
	URL                  *string  `json:"url,omitempty"`
	Brand                string   `json:"brand"`
	Code                 string   `json:"code"`
	Color                *string  `json:"color,omitempty"`
	Model                *string  `json:"model,omitempty"`
	Category             *string  `json:"category,omitempty"`
	Weight               *string  `json:"weight,omitempty"`
	LowestRecordedPrice  *float64 `json:"lowestRecordedPrice,omitempty"`
	HighestRecordedPrice *float64 `json:"highestRecordedPrice,omitempty"`
}

type CreatedByUser struct {
	ID     int64   `json:"id" sql:"primary_key" alias:"created_by_user.id"`
	Name   string  `json:"name" alias:"created_by_user.name"`
	Avatar *string `json:"avatar,omitempty" alias:"created_by_user.avatar"`
	Active *bool   `json:"active,omitempty" alias:"created_by_user.active"`
}

type Currency struct {
	CurrencyCode string `json:"currencyCode" sql:"primary_key"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	SymbolNative string `json:"symbolNative"`
	Decimals     int    `json:"decimals"`
	NumToBasic   *int   `json:"numToBasic,omitempty"`
}

type Mutation struct {
}

type Product struct {
	ID                   int64     `json:"id" sql:"primary_key"`
	Name                 string    `json:"name"`
	Image                string    `json:"image"`
	Description          string    `json:"description"`
	URL                  *string   `json:"url,omitempty"`
	Brand                string    `json:"brand"`
	Code                 string    `json:"code"`
	Color                *string   `json:"color,omitempty"`
	Model                *string   `json:"model,omitempty"`
	Category             *string   `json:"category,omitempty"`
	Weight               *string   `json:"weight,omitempty"`
	LowestRecordedPrice  *float64  `json:"lowestRecordedPrice,omitempty"`
	HighestRecordedPrice *float64  `json:"highestRecordedPrice,omitempty"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

type Query struct {
}

type UpdateUser struct {
	Name       *string         `json:"name,omitempty"`
	Avatar     *string         `json:"avatar,omitempty" validate:"omitempty,uuid"`
	AvatarFile *graphql.Upload `json:"avatarFile,omitempty"`
	BirthDate  *time.Time      `json:"birthDate,omitempty"`
	Bio        *string         `json:"bio,omitempty"`
}

type UpdatedByUser struct {
	ID     int64   `json:"id" sql:"primary_key" alias:"updated_by_user.id"`
	Name   string  `json:"name" alias:"updated_by_user.name"`
	Avatar *string `json:"avatar,omitempty" alias:"updated_by_user.avatar"`
	Active *bool   `json:"active,omitempty" alias:"updated_by_user.active"`
}

type User struct {
	ID           int64             `json:"id" sql:"primary_key"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
	Email        string            `json:"email"`
	PhoneNumber  *string           `json:"phoneNumber,omitempty"`
	Name         string            `json:"name"`
	Avatar       *string           `json:"avatar,omitempty"`
	BirthDate    *time.Time        `json:"birthDate,omitempty"`
	Bio          *string           `json:"bio,omitempty"`
	Active       bool              `json:"active"`
	AuthPlatform *AuthPlatformType `json:"authPlatform,omitempty" alias:"auth_state.platform"`
	AuthDevice   *AuthDeviceType   `json:"authDevice,omitempty" alias:"auth_state.device_type"`
	AuthStateID  *int64            `json:"authStateId,omitempty" alias:"auth_state.id"`
}

type UserShallow struct {
	ID     int64   `json:"id" sql:"primary_key" alias:"user.id"`
	Name   string  `json:"name" alias:"user.name"`
	Avatar *string `json:"avatar,omitempty" alias:"user.avatar"`
	Active *bool   `json:"active,omitempty" alias:"user.active"`
}

type AuthDeviceType string

const (
	AuthDeviceTypeIos     AuthDeviceType = "ios"
	AuthDeviceTypeAndroid AuthDeviceType = "android"
	AuthDeviceTypeWeb     AuthDeviceType = "web"
	AuthDeviceTypeOther   AuthDeviceType = "other"
	AuthDeviceTypeUnknown AuthDeviceType = "unknown"
)

var AllAuthDeviceType = []AuthDeviceType{
	AuthDeviceTypeIos,
	AuthDeviceTypeAndroid,
	AuthDeviceTypeWeb,
	AuthDeviceTypeOther,
	AuthDeviceTypeUnknown,
}

func (e AuthDeviceType) IsValid() bool {
	switch e {
	case AuthDeviceTypeIos, AuthDeviceTypeAndroid, AuthDeviceTypeWeb, AuthDeviceTypeOther, AuthDeviceTypeUnknown:
		return true
	}
	return false
}

func (e AuthDeviceType) String() string {
	return string(e)
}

func (e *AuthDeviceType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AuthDeviceType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AuthDeviceType", str)
	}
	return nil
}

func (e AuthDeviceType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type AuthPlatformType string

const (
	AuthPlatformTypeInternal AuthPlatformType = "INTERNAL"
	AuthPlatformTypeApple    AuthPlatformType = "APPLE"
	AuthPlatformTypeGoogle   AuthPlatformType = "GOOGLE"
)

var AllAuthPlatformType = []AuthPlatformType{
	AuthPlatformTypeInternal,
	AuthPlatformTypeApple,
	AuthPlatformTypeGoogle,
}

func (e AuthPlatformType) IsValid() bool {
	switch e {
	case AuthPlatformTypeInternal, AuthPlatformTypeApple, AuthPlatformTypeGoogle:
		return true
	}
	return false
}

func (e AuthPlatformType) String() string {
	return string(e)
}

func (e *AuthPlatformType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AuthPlatformType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AuthPlatformType", str)
	}
	return nil
}

func (e AuthPlatformType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
