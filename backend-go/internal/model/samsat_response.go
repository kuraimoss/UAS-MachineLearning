package model

type SamsatResponse struct {
	Plate         string `json:"plate"`
	OwnerName     string `json:"owner_name,omitempty"`
	VehicleBrand  string `json:"vehicle_brand,omitempty"`
	VehicleType   string `json:"vehicle_type,omitempty"`
	ChassisNumber string `json:"chassis_number,omitempty"`
	EngineNumber  string `json:"engine_number,omitempty"`
	ValidUntil    string `json:"valid_until,omitempty"`
}

