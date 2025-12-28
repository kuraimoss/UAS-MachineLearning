package scraper

import (
	"context"
	"errors"
	"fmt"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"plat-detection-system/backend-go/internal/config"
	"plat-detection-system/backend-go/internal/httpapi"
)

type SamsatScraper struct {
	cfg config.Config
}

func NewSamsatScraper(cfg config.Config) *SamsatScraper { return &SamsatScraper{cfg: cfg} }

type SamsatResult struct {
	Daerah        string `json:"daerah"`
	Provinsi      string `json:"provinsi"`
	WilayahSamsat string `json:"wilayah_samsat"`
	AlamatSamsat  string `json:"alamat_samsat"`
	Source        string `json:"source"`
}

func (s *SamsatScraper) Fetch(ctx context.Context, plate string) (*SamsatResult, error) {
	kodeDepan, angka, hurufBelakang, err := splitPlate(plate)
	if err != nil {
		return nil, err
	}
	return scrapeSamsatWithContext(ctx, s.cfg, kodeDepan, angka, hurufBelakang)
}

func ScrapeSamsat(kodeDepan string, angka string, hurufBelakang string) (*SamsatResult, error) {
	return scrapeSamsatWithContext(context.Background(), config.Load(), kodeDepan, angka, hurufBelakang)
}

func scrapeSamsatWithContext(ctx context.Context, cfg config.Config, kodeDepan string, angka string, hurufBelakang string) (*SamsatResult, error) {
	kodeDepan = strings.ToUpper(strings.TrimSpace(kodeDepan))
	angka = strings.TrimSpace(angka)
	hurufBelakang = strings.ToUpper(strings.TrimSpace(hurufBelakang))

	if kodeDepan == "" || angka == "" || hurufBelakang == "" {
		return nil, errors.New("kodeDepan, angka, dan hurufBelakang wajib diisi")
	}
	if !regexp.MustCompile(`^\d{1,4}$`).MatchString(angka) {
		return nil, errors.New("angka harus berupa 1-4 digit")
	}
	if !regexp.MustCompile(`^[A-Z]{1,3}$`).MatchString(hurufBelakang) {
		return nil, errors.New("hurufBelakang harus 1-3 huruf A-Z")
	}

	// Catatan:
	// Halaman samsat.info ini render hasil via JS dan mengambil data dari Firebase Firestore.
	// Jadi HTTP POST ke page HTML biasanya TIDAK mengembalikan div hasil.
	// Untuk mendapatkan data yang sama seperti di UI, kita panggil Firestore document endpoint langsung.
	if res, err := scrapeSamsatViaFirestore(ctx, cfg, kodeDepan, hurufBelakang); err == nil && res != nil {
		res.Source = "samsat.info"
		return res, nil
	}

	// Fallback: coba HTTP POST + parse HTML (kalau sewaktu-waktu mereka ubah jadi SSR).
	endpoint := cfg.SamsatPageURL

	form := url.Values{}
	// Karena atribut name di HTML tidak diketahui, kita kirim beberapa kandidat key
	// untuk memaksimalkan kompatibilitas server-side (tanpa browser automation).
	form.Set("kode_depan", kodeDepan)
	form.Set("angka", angka)
	form.Set("huruf_belakang", hurufBelakang)
	form.Set("kodeDepan", kodeDepan)
	form.Set("hurufBelakang", hurufBelakang)
	form.Set("depan", kodeDepan)
	form.Set("nomor", angka)
	form.Set("belakang", hurufBelakang)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; plat-detection-system/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{Timeout: cfg.ScrapeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("samsat.info status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}

	result := parseSamsatDocument(doc)
	if result == nil {
		return nil, errors.New("hasil tidak ditemukan pada response HTML")
	}
	result.Source = "samsat.info"
	return result, nil
}

type firestoreDoc struct {
	Fields map[string]struct {
		StringValue string `json:"stringValue"`
	} `json:"fields"`
}

func scrapeSamsatViaFirestore(ctx context.Context, cfg config.Config, kodeDepan string, hurufBelakang string) (*SamsatResult, error) {
	// logic key sama seperti UI:
	// untuk beberapa kode depan, mereka pakai huruf terakhir; sisanya huruf pertama.
	key := firestoreHurufKey(kodeDepan, hurufBelakang)
	if key == "" {
		return nil, errors.New("hurufBelakang tidak valid untuk lookup")
	}

	apiKey := cfg.SamsatFirestoreAPIKey
	u, _ := url.Parse(cfg.SamsatFirestoreBaseURL + "/nopol/" + url.PathEscape(kodeDepan) + "/belakang/" + url.PathEscape(key))
	q := u.Query()
	q.Set("key", apiKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; plat-detection-system/1.0)")
	req.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{Timeout: cfg.ScrapeTimeout}
	httpapi.Logger(ctx).Info("samsat lookup", "source", "firestore", "kode_depan", kodeDepan, "key", key)
	resp, err := client.Do(req)
	if err != nil {
		httpapi.Logger(ctx).Warn("samsat lookup failed", "error", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("firestore status %d", resp.StatusCode)
	}

	var doc firestoreDoc
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&doc); err != nil {
		return nil, err
	}

	get := func(k string) string {
		if v, ok := doc.Fields[k]; ok {
			return strings.TrimSpace(v.StringValue)
		}
		return ""
	}

	out := &SamsatResult{
		Daerah:        get("Daerah"),
		Provinsi:      get("Provinsi"),
		WilayahSamsat: get("Samsat"),
		AlamatSamsat:  get("Alamat"),
	}
	if out.Daerah == "" && out.Provinsi == "" && out.WilayahSamsat == "" && out.AlamatSamsat == "" {
		return nil, nil
	}
	return out, nil
}

func firestoreHurufKey(kodeDepan string, hurufBelakang string) string {
	hurufBelakang = strings.ToUpper(strings.TrimSpace(hurufBelakang))
	if hurufBelakang == "" {
		return ""
	}
	specialLast := map[string]bool{
		"G": true, "H": true, "K": true, "R": true,
		"AA": true, "AD": true, "DT": true,
	}
	if specialLast[kodeDepan] {
		return string(hurufBelakang[len(hurufBelakang)-1])
	}
	return string(hurufBelakang[0])
}

func parseSamsatDocument(doc *goquery.Document) *SamsatResult {
	if doc == nil {
		return nil
	}

	// Struktur yang diminta:
	// <div class="mt-4 bg-gray-200 p-4 rounded"> ... <p>Label</p><p>Value</p> ... </div>
	// Urutan class bisa berubah, jadi kita gunakan kombinasi selector + fallback.
	containers := doc.Find("div.mt-4.bg-gray-200.p-4.rounded")
	if containers.Length() == 0 {
		containers = doc.Find("div.mt-4.bg-gray-200")
	}

	var out SamsatResult
	foundAny := false

	containers.EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		texts := make([]string, 0, 16)
		sel.Find("p").Each(func(_ int, p *goquery.Selection) {
			t := strings.TrimSpace(p.Text())
			if t != "" {
				texts = append(texts, t)
			}
		})

		// label-value pairs: ["Daerah Kendaraan:", "Kota Medan", "Provinsi:", "Provinsi Sumatera Utara", ...]
		for i := 0; i+1 < len(texts); i++ {
			label := normalizeLabel(texts[i])
			value := strings.TrimSpace(texts[i+1])
			if value == "" {
				continue
			}

			switch label {
			case "daerah kendaraan":
				if out.Daerah == "" {
					out.Daerah = value
					foundAny = true
				}
			case "provinsi":
				if out.Provinsi == "" {
					out.Provinsi = value
					foundAny = true
				}
			case "wilayah samsat":
				if out.WilayahSamsat == "" {
					out.WilayahSamsat = value
					foundAny = true
				}
			case "alamat samsat":
				if out.AlamatSamsat == "" {
					out.AlamatSamsat = value
					foundAny = true
				}
			}
		}

		return false // stop after first container
	})

	if !foundAny {
		return nil
	}
	return &out
}

func normalizeLabel(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimSuffix(s, ":")
	return s
}

func splitPlate(plate string) (kodeDepan string, angka string, hurufBelakang string, err error) {
	compact := strings.ToUpper(strings.TrimSpace(plate))
	compact = strings.ReplaceAll(compact, " ", "")
	compact = strings.ReplaceAll(compact, "-", "")

	// Umum di Indonesia: 1-2 huruf depan, 1-4 digit, 1-3 huruf belakang.
	re := regexp.MustCompile(`^([A-Z]{1,2})(\d{1,4})([A-Z]{1,3})$`)
	m := re.FindStringSubmatch(compact)
	if len(m) != 4 {
		return "", "", "", fmt.Errorf("format plat tidak valid: %q", plate)
	}
	return m[1], m[2], m[3], nil
}
