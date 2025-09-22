package models

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SpecificationType represents the type of product specification
type SpecificationType string

const (
	// SpecificationTypeTire represents tire specifications
	SpecificationTypeTire SpecificationType = "Tire"
	// SpecificationTypeWheel represents wheel specifications
	SpecificationTypeWheel SpecificationType = "Wheel"
)

// TireSpecification represents tire-specific specifications
type TireSpecification struct {
	Width       string `json:"width"`       // e.g., "225"
	AspectRatio string `json:"aspectRatio"` // e.g., "45"
	Diameter    string `json:"diameter"`    // e.g., "17"
	LoadIndex   string `json:"loadIndex"`   // e.g., "91"
	SpeedRating string `json:"speedRating"` // e.g., "W"
	DOTYear     string `json:"dotYear"`     // e.g., "2023"
	Season      string `json:"season"`      // e.g., "All-Season", "Summer", "Winter"
	RunFlat     bool   `json:"runFlat"`     // Run-flat technology
}

// WheelSpecification represents wheel-specific specifications
type WheelSpecification struct {
	Diameter    string `json:"diameter"`    // e.g., "17"
	Width       string `json:"width"`       // e.g., "8.5"
	Offset      string `json:"offset"`      // e.g., "35"
	BoltPattern string `json:"boltPattern"` // e.g., "5x114.3"
	CenterBore  string `json:"centerBore"`  // e.g., "67.1"
	Color       string `json:"color"`       // e.g., "Black", "Silver"
	Finish      string `json:"finish"`      // e.g., "Matte", "Glossy"
	Weight      string `json:"weight"`      // e.g., "22.5" (in kg)
}

// ProductSpecification represents product specifications for tires and wheels
type ProductSpecification struct {
	ID        uint              `json:"id" gorm:"primaryKey;autoIncrement"`
	ProductID uint              `json:"productId" gorm:"not null;index"`
	SpecType  SpecificationType `json:"specType" gorm:"not null;type:varchar(20)"`
	SpecData  json.RawMessage   `json:"specData" gorm:"type:jsonb"`
	CreatedAt time.Time         `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time         `json:"updatedAt" gorm:"autoUpdateTime"`

	// Relationships
	Product Product `json:"-" gorm:"foreignKey:ProductID;references:ID;constraint:OnDelete:CASCADE"`
}

// ProductSpecificationCreateRequest represents the request payload for creating a product specification
type ProductSpecificationCreateRequest struct {
	SpecType SpecificationType `json:"specType" binding:"required"`
	SpecData interface{}       `json:"specData" binding:"required"`
}

// ProductSpecificationUpdateRequest represents the request payload for updating a product specification
type ProductSpecificationUpdateRequest struct {
	SpecType *SpecificationType `json:"specType,omitempty"`
	SpecData *interface{}       `json:"specData,omitempty"`
}

// ProductSpecificationResponse represents the response payload for product specification data
type ProductSpecificationResponse struct {
	ID        uint              `json:"id"`
	ProductID uint              `json:"productId"`
	SpecType  SpecificationType `json:"specType"`
	SpecData  interface{}       `json:"specData"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// BeforeCreate is a GORM hook that runs before creating a product specification
func (ps *ProductSpecification) BeforeCreate(_ *gorm.DB) error {
	// Validate specification data
	if err := ps.Validate(); err != nil {
		return err
	}

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a product specification
func (ps *ProductSpecification) BeforeUpdate(_ *gorm.DB) error {
	// Validate specification data
	if err := ps.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the product specification data
func (ps *ProductSpecification) Validate() error {
	var validationErrors []string

	// Validate spec type
	if err := ps.ValidateSpecType(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	// Validate spec data
	if err := ps.ValidateSpecData(); err != nil {
		validationErrors = append(validationErrors, err.Error())
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateSpecType validates the specification type
func (ps *ProductSpecification) ValidateSpecType() error {
	validTypes := []SpecificationType{SpecificationTypeTire, SpecificationTypeWheel}

	for _, specType := range validTypes {
		if ps.SpecType == specType {
			return nil
		}
	}

	return errors.New("specType must be one of: Tire, Wheel")
}

// ValidateSpecData validates the specification data based on the spec type
func (ps *ProductSpecification) ValidateSpecData() error {
	if len(ps.SpecData) == 0 {
		return errors.New("specData is required")
	}

	switch ps.SpecType {
	case SpecificationTypeTire:
		return ps.ValidateTireSpecData()
	case SpecificationTypeWheel:
		return ps.ValidateWheelSpecData()
	default:
		return errors.New("invalid specType for validation")
	}
}

// ValidateTireSpecData validates tire specification data
func (ps *ProductSpecification) ValidateTireSpecData() error {
	var tireSpec TireSpecification
	if err := json.Unmarshal(ps.SpecData, &tireSpec); err != nil {
		return errors.New("invalid tire specification data format")
	}

	var validationErrors []string

	// Validate required fields
	if tireSpec.Width == "" {
		validationErrors = append(validationErrors, "width is required for tire specification")
	}
	if tireSpec.AspectRatio == "" {
		validationErrors = append(validationErrors, "aspectRatio is required for tire specification")
	}
	if tireSpec.Diameter == "" {
		validationErrors = append(validationErrors, "diameter is required for tire specification")
	}
	if tireSpec.LoadIndex == "" {
		validationErrors = append(validationErrors, "loadIndex is required for tire specification")
	}
	if tireSpec.SpeedRating == "" {
		validationErrors = append(validationErrors, "speedRating is required for tire specification")
	}

	// Validate season
	validSeasons := []string{"All-Season", "Summer", "Winter", "Performance"}
	if tireSpec.Season != "" {
		valid := false
		for _, season := range validSeasons {
			if tireSpec.Season == season {
				valid = true
				break
			}
		}
		if !valid {
			validationErrors = append(validationErrors, "season must be one of: All-Season, Summer, Winter, Performance")
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// ValidateWheelSpecData validates wheel specification data
func (ps *ProductSpecification) ValidateWheelSpecData() error {
	var wheelSpec WheelSpecification
	if err := json.Unmarshal(ps.SpecData, &wheelSpec); err != nil {
		return errors.New("invalid wheel specification data format")
	}

	var validationErrors []string

	// Validate required fields
	if wheelSpec.Diameter == "" {
		validationErrors = append(validationErrors, "diameter is required for wheel specification")
	}
	if wheelSpec.Width == "" {
		validationErrors = append(validationErrors, "width is required for wheel specification")
	}
	if wheelSpec.Offset == "" {
		validationErrors = append(validationErrors, "offset is required for wheel specification")
	}
	if wheelSpec.BoltPattern == "" {
		validationErrors = append(validationErrors, "boltPattern is required for wheel specification")
	}
	if wheelSpec.CenterBore == "" {
		validationErrors = append(validationErrors, "centerBore is required for wheel specification")
	}

	// Validate color
	validColors := []string{"Black", "Silver", "White", "Gray", "Gold", "Bronze", "Chrome"}
	if wheelSpec.Color != "" {
		valid := false
		for _, color := range validColors {
			if wheelSpec.Color == color {
				valid = true
				break
			}
		}
		if !valid {
			validationErrors = append(validationErrors, "color must be one of: Black, Silver, White, Gray, Gold, Bronze, Chrome")
		}
	}

	// Validate finish
	validFinishes := []string{"Matte", "Glossy", "Satin", "Brushed"}
	if wheelSpec.Finish != "" {
		valid := false
		for _, finish := range validFinishes {
			if wheelSpec.Finish == finish {
				valid = true
				break
			}
		}
		if !valid {
			validationErrors = append(validationErrors, "finish must be one of: Matte, Glossy, Satin, Brushed")
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(strings.Join(validationErrors, "; "))
	}

	return nil
}

// SetTireSpecData sets tire specification data
func (ps *ProductSpecification) SetTireSpecData(tireSpec *TireSpecification) error {
	ps.SpecType = SpecificationTypeTire
	specData, err := json.Marshal(tireSpec)
	if err != nil {
		return err
	}
	ps.SpecData = specData
	return ps.Validate()
}

// SetWheelSpecData sets wheel specification data
func (ps *ProductSpecification) SetWheelSpecData(wheelSpec *WheelSpecification) error {
	ps.SpecType = SpecificationTypeWheel
	specData, err := json.Marshal(wheelSpec)
	if err != nil {
		return err
	}
	ps.SpecData = specData
	return ps.Validate()
}

// GetTireSpecData returns tire specification data
func (ps *ProductSpecification) GetTireSpecData() (*TireSpecification, error) {
	if ps.SpecType != SpecificationTypeTire {
		return nil, errors.New("specification is not a tire specification")
	}

	var tireSpec TireSpecification
	if err := json.Unmarshal(ps.SpecData, &tireSpec); err != nil {
		return nil, err
	}

	return &tireSpec, nil
}

// GetWheelSpecData returns wheel specification data
func (ps *ProductSpecification) GetWheelSpecData() (*WheelSpecification, error) {
	if ps.SpecType != SpecificationTypeWheel {
		return nil, errors.New("specification is not a wheel specification")
	}

	var wheelSpec WheelSpecification
	if err := json.Unmarshal(ps.SpecData, &wheelSpec); err != nil {
		return nil, err
	}

	return &wheelSpec, nil
}

// GetSpecDataAsInterface returns specification data as interface{}
func (ps *ProductSpecification) GetSpecDataAsInterface() (interface{}, error) {
	switch ps.SpecType {
	case SpecificationTypeTire:
		return ps.GetTireSpecData()
	case SpecificationTypeWheel:
		return ps.GetWheelSpecData()
	default:
		return nil, errors.New("unknown specification type")
	}
}

// ToResponse converts a ProductSpecification to ProductSpecificationResponse
func (ps *ProductSpecification) ToResponse() ProductSpecificationResponse {
	specData, _ := ps.GetSpecDataAsInterface()

	return ProductSpecificationResponse{
		ID:        ps.ID,
		ProductID: ps.ProductID,
		SpecType:  ps.SpecType,
		SpecData:  specData,
		CreatedAt: ps.CreatedAt,
		UpdatedAt: ps.UpdatedAt,
	}
}

// TableName returns the table name for the ProductSpecification model
func (ProductSpecification) TableName() string {
	return "product_specifications"
}
