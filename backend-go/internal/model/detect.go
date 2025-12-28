package model

type DetectData struct {
	Plate         PlateInfo      `json:"plate"`
	VehicleRegion *VehicleRegion `json:"vehicle_region,omitempty"`
}

type PlateInfo struct {
	Raw        string          `json:"raw"`
	Formatted  string          `json:"formatted"`
	Confidence float64         `json:"confidence"`
	Components PlateComponents `json:"components"`
}

type PlateComponents struct {
	Prefix string `json:"prefix"`
	Number string `json:"number"`
	Suffix string `json:"suffix"`
}

type VehicleRegion struct {
	Daerah        string `json:"daerah"`
	Provinsi      string `json:"provinsi"`
	WilayahSamsat string `json:"wilayah_samsat"`
	AlamatSamsat  string `json:"alamat_samsat"`
	Source        string `json:"source"`
}

type DetectMeta struct {
	Model    ModelMeta `json:"model"`
	Warnings []string  `json:"warnings,omitempty"`
}

type ModelMeta struct {
	YOLO     string `json:"yolo"`
	OCR      string `json:"ocr"`
	Language string `json:"language"`
}

