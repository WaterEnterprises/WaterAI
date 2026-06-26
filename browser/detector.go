package browser

// detector.go

type Detector interface {
	DetectFromImage(imageB64 string, scaleFactor float64, detectSheets bool) ([]InteractiveElement, error)
}