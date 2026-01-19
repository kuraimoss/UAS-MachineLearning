package service

import (
	"context"
	"net/http"
	"strings"

	"plat-detection-system/backend-go/internal/config"
	"plat-detection-system/backend-go/internal/errx"
	"plat-detection-system/backend-go/internal/httpapi"
	"plat-detection-system/backend-go/internal/model"
	"plat-detection-system/backend-go/internal/util"
)

type DetectService struct {
	cfg    config.Config
	yolo   *YOLOService
	samsat *SamsatService
}

func NewDetectService(cfg config.Config, yolo *YOLOService, samsat *SamsatService) *DetectService {
	return &DetectService{cfg: cfg, yolo: yolo, samsat: samsat}
}

func (s *DetectService) Detect(ctx context.Context, imagePath string) (model.DetectData, model.DetectMeta, error) {
	if strings.TrimSpace(imagePath) == "" {
		return model.DetectData{}, model.DetectMeta{}, errx.New("INVALID_IMAGE_PATH", "image path kosong", http.StatusBadRequest)
	}
	if s.yolo == nil {
		return model.DetectData{}, model.DetectMeta{}, errx.New("YOLO_NOT_CONFIGURED", "YOLO service tidak tersedia", http.StatusInternalServerError)
	}

	plate, err := s.yolo.Detect(ctx, imagePath)
	if err != nil {
		return model.DetectData{}, model.DetectMeta{}, err
	}

	plateRaw := plate.Raw
	if strings.TrimSpace(plateRaw) == "" {
		plateRaw = plate.Cleaned
	}
	plateForFormat := plate.Cleaned
	if strings.TrimSpace(plateForFormat) == "" {
		plateForFormat = plateRaw
	}
	plateRawNoSpace := util.NormalizePlateRaw(plateForFormat)
	if plateRawNoSpace == "" {
		return model.DetectData{}, model.DetectMeta{}, errx.New("OCR_EMPTY", "hasil OCR kosong", http.StatusUnprocessableEntity)
	}

	if s.cfg.MinPlateConfidence > 0 && plate.Confidence < s.cfg.MinPlateConfidence {
		return model.DetectData{}, model.DetectMeta{}, errx.New("PLATE_CONFIDENCE_LOW", "confidence terlalu rendah", http.StatusUnprocessableEntity)
	}

	components, formatted := util.SplitAndFormatPlate(plateRawNoSpace)
	data := model.DetectData{
		Plate: model.PlateInfo{
			Raw:        plateRaw,
			Formatted:  formatted,
			Confidence: plate.Confidence,
			Components: components,
		},
	}

	meta := model.DetectMeta{
		Model: model.ModelMeta{
			YOLO:     envOr("YOLO_META_NAME", "YOLOv8 custom"),
			OCR:      envOr("OCR_META_NAME", "PaddleOCR"),
			Language: envOr("OCR_LANG", "id"),
		},
	}

	if s.samsat == nil {
		meta.Warnings = append(meta.Warnings, "vehicle_region service not configured")
		return data, meta, nil
	}

	scrapeCtx, cancel := context.WithTimeout(ctx, s.cfg.ScrapeTimeout)
	defer cancel()

	vr, err := s.samsat.LookupByPlate(scrapeCtx, plateRawNoSpace)
	if err != nil {
		httpapi.Logger(ctx).Warn("vehicle_region lookup failed", "error", err.Error())
		meta.Warnings = append(meta.Warnings, "vehicle_region lookup failed")
		return data, meta, nil
	}
	if vr == nil {
		meta.Warnings = append(meta.Warnings, "vehicle_region not found")
		return data, meta, nil
	}

	data.VehicleRegion = &model.VehicleRegion{
		Daerah:        vr.Daerah,
		Provinsi:      vr.Provinsi,
		WilayahSamsat: vr.WilayahSamsat,
		AlamatSamsat:  vr.AlamatSamsat,
		Source:        "https://samsat.info",
	}
	return data, meta, nil
}
