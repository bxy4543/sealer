package types

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

// GraphDriverData Information about a container's graph driver.
// swagger:model GraphDriverData
type GraphDriverData struct {

	// data
	// Required: true
	Data map[string]string `json:"Data"`

	// name
	// Required: true
	Name string `json:"Name"`
}
