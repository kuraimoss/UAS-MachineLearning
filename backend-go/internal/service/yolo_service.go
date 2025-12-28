package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"time"

	"plat-detection-system/backend-go/internal/config"
	"plat-detection-system/backend-go/internal/errx"
	"plat-detection-system/backend-go/internal/httpapi"
)

type YOLOService struct {
	cfg          config.Config
	pythonBin    string
	pythonScript string
	timeout      time.Duration
}

func NewYOLOService(cfg config.Config) *YOLOService {
	return &YOLOService{
		cfg:          cfg,
		pythonBin:    cfg.PythonBin,
		pythonScript: cfg.PythonScript,
		timeout:      cfg.PythonTimeout,
	}
}

type yoloOut struct {
	PlateRaw     string  `json:"plate_raw"`
	PlateCleaned string  `json:"plate_cleaned"`
	Confidence   float64 `json:"confidence"`
	Error        string  `json:"error"`
}

type PlateResult struct {
	Raw        string
	Cleaned    string
	Confidence float64
}

func (s *YOLOService) Detect(ctx context.Context, imagePath string) (PlateResult, error) {
	if s.pythonScript == "" {
		return PlateResult{}, errx.New("PYTHON_SCRIPT_NOT_SET", "python script belum di-set", http.StatusInternalServerError)
	}

	ctx, cancel := withTimeout(ctx, s.timeout)
	defer cancel()

	pythonExe := s.pythonBin
	if pythonExe == "" {
		pythonExe = "python"
	}

	cmd := exec.CommandContext(ctx, pythonExe, "-u", s.pythonScript, imagePath)
	cmd.Env = append(os.Environ(),
		"PYTHONUNBUFFERED=1",
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	var out yoloOut
	runErr := cmd.Run()

	// Python diminta selalu output JSON (termasuk saat error) via stdout.
	stdoutBytes := bytes.TrimSpace(stdout.Bytes())
	if err := json.Unmarshal(stdoutBytes, &out); err != nil {
		if runErr != nil || errorsIsContextDeadline(ctx) {
			if len(stdoutBytes) == 0 {
				if stderr.Len() > 0 {
					httpapi.Logger(ctx).Error("python failed (stdout empty)", "stderr", stderr.String())
				}
				if errorsIsContextDeadline(ctx) {
					return PlateResult{}, errx.Wrap(runErr, "PYTHON_TIMEOUT", "python timeout", http.StatusGatewayTimeout)
				}
				return PlateResult{}, errx.Wrap(runErr, "PYTHON_FAILED", "python gagal dieksekusi", http.StatusBadGateway)
			}
			preview := stdoutBytes
			if len(preview) > 500 {
				preview = preview[:500]
			}
			if stderr.Len() > 0 {
				httpapi.Logger(ctx).Error("python failed", "stderr", stderr.String())
			}
			httpapi.Logger(ctx).Error("python bad output", "stdout_preview", string(preview))
			return PlateResult{}, errx.Wrap(runErr, "PYTHON_BAD_OUTPUT", "python output tidak valid", http.StatusBadGateway)
		}
		return PlateResult{}, errx.Wrap(err, "PYTHON_BAD_OUTPUT", "python output tidak valid", http.StatusBadGateway)
	}

	if out.Error != "" {
		// Error domain dari python (mis. plat tidak terdeteksi / gambar tidak bisa dibaca).
		return PlateResult{}, errx.New("INFERENCE_FAILED", out.Error, http.StatusUnprocessableEntity)
	}
	if runErr != nil {
		// Jika proses non-zero tapi payload tidak punya error, fallback ke stderr/runErr.
		if stderr.Len() > 0 {
			httpapi.Logger(ctx).Error("python exit non-zero", "stderr", stderr.String())
		}
		if errorsIsContextDeadline(ctx) {
			return PlateResult{}, errx.Wrap(runErr, "PYTHON_TIMEOUT", "python timeout", http.StatusGatewayTimeout)
		}
		return PlateResult{}, errx.Wrap(runErr, "PYTHON_FAILED", "python gagal dieksekusi", http.StatusBadGateway)
	}

	return PlateResult{Raw: out.PlateRaw, Cleaned: out.PlateCleaned, Confidence: out.Confidence}, nil
}

func errorsIsContextDeadline(ctx context.Context) bool {
	return ctx.Err() == context.DeadlineExceeded || ctx.Err() == context.Canceled
}
