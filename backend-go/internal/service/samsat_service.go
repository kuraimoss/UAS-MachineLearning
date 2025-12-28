package service

import (
	"context"

	"plat-detection-system/backend-go/internal/scraper"
)

type SamsatService struct {
	scraper *scraper.SamsatScraper
}

func NewSamsatService(s *scraper.SamsatScraper) *SamsatService {
	return &SamsatService{scraper: s}
}

func (s *SamsatService) LookupByPlate(ctx context.Context, plate string) (*scraper.SamsatResult, error) {
	if s.scraper == nil {
		return nil, nil
	}
	return s.scraper.Fetch(ctx, plate)
}
