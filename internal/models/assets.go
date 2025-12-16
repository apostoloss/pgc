package models

import (
	"errors"
)

// AssetBase holds common fields for all asset types (Chart, Insight, Audience).
type AssetBase struct {
	ID          string // Unique identifier for the asset
	Name        string // User-friendly name
	Description string // Short description or summary
}

// Validate checks common fields.
func (a AssetBase) Validate() error {
	if a.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// Chart represents a favorite chart asset.
type Chart struct {
	AssetBase         // Embeds the common asset fields
	ChartType  string // Type of chart, e.g. bar, line, pie
	DataSource string // Data source or reference
}

// Validate checks Chart fields.
func (c Chart) Validate() error {
	if err := c.AssetBase.Validate(); err != nil {
		return err
	}
	if c.ChartType == "" {
		return errors.New("chart type is required")
	}
	return nil
}

// Insight represents a favorite insight asset.
type Insight struct {
	AssetBase        // Embeds the common asset fields
	Metric    string // Key metric or subject
	Value     string // Insight value or summary
}

// Validate checks Insight fields.
func (i Insight) Validate() error {
	if err := i.AssetBase.Validate(); err != nil {
		return err
	}
	if i.Metric == "" {
		return errors.New("metric is required")
	}
	if i.Value == "" {
		return errors.New("value is required")
	}
	return nil
}

// Audience represents a favorite audience asset.
type Audience struct {
	AssetBase        // Embeds the common asset fields
	Segment   string // Segment or group name
	Size      int    // Estimated audience size
}

// Validate checks Audience fields.
func (a Audience) Validate() error {
	if err := a.AssetBase.Validate(); err != nil {
		return err
	}
	if a.Segment == "" {
		return errors.New("segment is required")
	}
	if a.Size < 0 {
		return errors.New("size cannot be negative")
	}
	return nil
}

// Asset is an interface for polymorphism, allowing operations on any asset type.
type Asset interface {
	GetID() string
	GetName() string
	GetDescription() string
	Validate() error
}

// Implement Asset interface for Chart
func (c Chart) GetID() string          { return c.ID }
func (c Chart) GetName() string        { return c.Name }
func (c Chart) GetDescription() string { return c.Description }

// Implement Asset interface for Insight
func (i Insight) GetID() string          { return i.ID }
func (i Insight) GetName() string        { return i.Name }
func (i Insight) GetDescription() string { return i.Description }

// Implement Asset interface for Audience
func (a Audience) GetID() string          { return a.ID }
func (a Audience) GetName() string        { return a.Name }
func (a Audience) GetDescription() string { return a.Description }
