package browser

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/fogleman/gg" // Graphics library equivalent to PIL ImageDraw
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// PutHighlightElementsOnScreenshot draws bounding boxes and labels on the screenshot.
func PutHighlightElementsOnScreenshot(elements map[int]InteractiveElement, screenshotB64 string) string {
	decodedData, err := base64.StdEncoding.DecodeString(screenshotB64)
	if err != nil {
		log.Printf("Failed to decode screenshot base64: %v", err)
		return screenshotB64
	}

	img, _, err := image.Decode(bytes.NewReader(decodedData))
	if err != nil {
		log.Printf("Failed to decode screenshot image: %v", err)
		return screenshotB64
	}

	// Use gg context for drawing
	dc := gg.NewContextForImage(img)

	// Load font
	var face font.Face
	if len(OpenSansFont) > 0 {
		f, err := opentype.Parse(OpenSansFont)
		if err == nil {
			face, _ = opentype.NewFace(f, &opentype.FaceOptions{
				Size:    11,
				DPI:     72,
				Hinting: font.HintingFull,
			})
			dc.SetFontFace(face)
		} else {
			log.Printf("Failed to parse font: %v", err)
		}
	}

	baseColors := [][]int{
		{204, 0, 0}, {0, 136, 0}, {0, 0, 204}, {204, 112, 0},
		{102, 0, 102}, {0, 102, 102}, {204, 51, 153}, {44, 0, 102},
		{204, 35, 0}, {28, 102, 66}, {170, 0, 0}, {36, 82, 123},
	}

	type LabelRect struct {
		Left, Top, Right, Bottom float64
	}
	var placedLabels []LabelRect

	// Map iteration is random in Go, but we want stability if we were looping strictly.
	// However, the input is a map, so we iterate as is. The logic relies on ID.
	
	// Create a sorted list of keys to ensure deterministic drawing order
	var keys []int
	for k := range elements {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, idx := range keys {
		element := elements[idx]

		// Skip sheets elements
		if strings.HasPrefix(element.BrowserAgentID, "row_") || strings.HasPrefix(element.BrowserAgentID, "column_") {
			continue
		}

		// Color generation
		baseColor := baseColors[idx%len(baseColors)]
		r, g, b := generateUniqueColor(baseColor, idx)

		rect := element.Rect

		// Draw rectangle
		dc.SetRGB255(r, g, b)
		dc.SetLineWidth(2)
		dc.DrawRectangle(rect.Left, rect.Top, rect.Width, rect.Height)
		dc.Stroke()

		// Prepare label
		labelText := fmt.Sprintf("%d", idx)
		textWidth, textHeight := dc.MeasureString(labelText)
		
		// Adjust dimensions for aesthetics
		labelWidth := textWidth + 4
		labelHeight := textHeight + 4

		labelX := rect.Left + rect.Width - labelWidth
		labelY := rect.Top

		if labelWidth > rect.Width || labelHeight > rect.Height {
			labelX = rect.Left + rect.Width
			labelY = rect.Top
		}

		// Check overlap
		currLabel := LabelRect{labelX, labelY, labelX + labelWidth, labelY + labelHeight}
		
		for _, existing := range placedLabels {
			if !(currLabel.Right < existing.Left || currLabel.Left > existing.Right || currLabel.Bottom < existing.Top || currLabel.Top > existing.Bottom) {
				// Overlap detected, push down
				labelY = existing.Bottom + 2
				currLabel.Top = labelY
				currLabel.Bottom = labelY + labelHeight
				// Simple break, might need restart in complex cases but matches Python
				break 
			}
		}
		
		// Boundaries check
		imgWidth := float64(dc.Width())
		imgHeight := float64(dc.Height())

		if currLabel.Left < 0 {
			currLabel.Left = 0
			currLabel.Right = labelWidth
		} else if currLabel.Right >= imgWidth {
			currLabel.Left = imgWidth - labelWidth - 1
			currLabel.Right = imgWidth - 1
		}

		if currLabel.Top < 0 {
			currLabel.Top = 0
			currLabel.Bottom = labelHeight
		} else if currLabel.Bottom >= imgHeight {
			currLabel.Top = imgHeight - labelHeight - 1
			currLabel.Bottom = imgHeight - 1
		}

		// Draw Label Background
		dc.SetRGB255(r, g, b)
		dc.DrawRectangle(currLabel.Left, currLabel.Top, labelWidth, labelHeight)
		dc.Fill()

		// Draw Text
		dc.SetRGB255(255, 255, 255)
		// gg draws text anchored at bottom-left by default roughly, but MeasureString helps.
		// However, gg's DrawString anchors at baseline.
		// We use magic offsets from Python: x+3, y-1 (but Python DrawText is top-left anchor).
		// In gg, we need to center it or approximate.
		dc.DrawString(labelText, currLabel.Left+2, currLabel.Top+textHeight) // Approximation
		
		placedLabels = append(placedLabels, currLabel)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		log.Printf("Failed to encode highlighted image: %v", err)
		return screenshotB64
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func generateUniqueColor(baseColor []int, idx int) (int, int, int) {
	r, g, b := baseColor[0], baseColor[1], baseColor[2]

	offsetR := (idx * 17) % 31 - 15
	offsetG := (idx * 23) % 29 - 14
	offsetB := (idx * 13) % 27 - 13

	r = clamp(r + offsetR)
	g = clamp(g + offsetG)
	b = clamp(b + offsetB)

	return r, g, b
}

func clamp(val int) int {
	if val < 0 {
		return 0
	}
	if val > 255 {
		return 255
	}
	return val
}

func ScaleB64Image(imageB64 string, scaleFactor float64) string {
	if scaleFactor == 1.0 {
		return imageB64
	}
	
	decodedData, err := base64.StdEncoding.DecodeString(imageB64)
	if err != nil {
		return imageB64
	}

	img, _, err := image.Decode(bytes.NewReader(decodedData))
	if err != nil {
		return imageB64
	}

	// Use gg for resizing (it wraps standard image libs nicely)
	origW := img.Bounds().Dx()
	origH := img.Bounds().Dy()
	newW := int(float64(origW) * scaleFactor)
	newH := int(float64(origH) * scaleFactor)

	dc := gg.NewContext(newW, newH)
	dc.DrawImage(img, 0, 0) // This doesn't resize, wait. gg doesn't have built-in resize.
	// We need standard library resize or a helper.
	// For standard lib simplicity, let's use a basic nearest neighbor or rely on gg context scaling.
	
	dc.Scale(scaleFactor, scaleFactor)
	dc.DrawImage(img, 0, 0)
	
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return imageB64
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

// FilterElements implementation matching Python
func FilterElements(elements []InteractiveElement, iouThreshold float64) []InteractiveElement {
	filtered := filterOverlappingElements(elements, iouThreshold)
	return sortElementsByPosition(filtered)
}

func filterOverlappingElements(elements []InteractiveElement, iouThreshold float64) []InteractiveElement {
	if len(elements) == 0 {
		return nil
	}

	// Sort by Area descending, then Weight descending
	sort.Slice(elements, func(i, j int) bool {
		areaI := elements[i].Rect.Width * elements[i].Rect.Height
		areaJ := elements[j].Rect.Width * elements[j].Rect.Height
		if areaI != areaJ {
			return areaI > areaJ
		}
		return elements[i].Weight > elements[j].Weight
	})

	var filtered []InteractiveElement

	for _, current := range elements {
		shouldAdd := true

		for k := 0; k < len(filtered); k++ {
			existing := filtered[k]
			
			iou := calculateIOU(current.Rect, existing.Rect)
			if iou > iouThreshold {
				shouldAdd = false
				break
			}

			if isFullyContained(current.Rect, existing.Rect) {
				if existing.Weight >= current.Weight && existing.ZIndex == current.ZIndex {
					shouldAdd = false
					break
				} else {
					// Check if current is bigger than 50% of existing
					currArea := current.Rect.Width * current.Rect.Height
					existArea := existing.Rect.Width * existing.Rect.Height
					if currArea >= existArea*0.5 {
						// Remove existing
						filtered = append(filtered[:k], filtered[k+1:]...)
						k-- // Adjust index
						break
					}
				}
			}
		}

		if shouldAdd {
			filtered = append(filtered, current)
		}
	}
	return filtered
}

func calculateIOU(r1, r2 Rect) float64 {
	intersectLeft := math.Max(r1.Left, r2.Left)
	intersectTop := math.Max(r1.Top, r2.Top)
	intersectRight := math.Min(r1.Right, r2.Right)
	intersectBottom := math.Min(r1.Bottom, r2.Bottom)

	if intersectRight < intersectLeft || intersectBottom < intersectTop {
		return 0.0
	}

	area1 := (r1.Right - r1.Left) * (r1.Bottom - r1.Top)
	area2 := (r2.Right - r2.Left) * (r2.Bottom - r2.Top)
	intersectArea := (intersectRight - intersectLeft) * (intersectBottom - intersectTop)

	unionArea := area1 + area2 - intersectArea
	if unionArea > 0 {
		return intersectArea / unionArea
	}
	return 0.0
}

func isFullyContained(r1, r2 Rect) bool {
	return r1.Left >= r2.Left && r1.Right <= r2.Right &&
		r1.Top >= r2.Top && r1.Bottom <= r2.Bottom
}

func sortElementsByPosition(elements []InteractiveElement) []InteractiveElement {
	if len(elements) == 0 {
		return nil
	}
	rowThreshold := 20.0

	// Sort by Y first
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Rect.Top < elements[j].Rect.Top
	})

	var rows [][]InteractiveElement
	var currentRow []InteractiveElement

	for _, el := range elements {
		if len(currentRow) == 0 {
			currentRow = append(currentRow, el)
		} else {
			lastEl := currentRow[len(currentRow)-1]
			if math.Abs(el.Rect.Top-lastEl.Rect.Top) <= rowThreshold {
				currentRow = append(currentRow, el)
			} else {
				rows = append(rows, currentRow)
				currentRow = []InteractiveElement{el}
			}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	// Sort each row by X and flatten
	var sortedList []InteractiveElement
	idx := 0
	for _, row := range rows {
		sort.Slice(row, func(i, j int) bool {
			return row[i].Rect.Left < row[j].Rect.Left
		})
		for _, el := range row {
			el.Index = idx
			sortedList = append(sortedList, el)
			idx++
		}
	}

	return sortedList
}

func IsPDFURL(targetURL string) bool {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false
	}
	if strings.HasSuffix(strings.ToLower(u.Path), ".pdf") {
		return true
	}

	client := &http.Client{Timeout: 5 * time.Second}
	
	// HEAD request
	resp, err := client.Head(targetURL)
	if err == nil {
		defer resp.Body.Close()
		ct := strings.ToLower(resp.Header.Get("Content-Type"))
		if strings.Contains(ct, "application/pdf") {
			return true
		}
	}

	// Fallback GET
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return false
	}
	// Use Range header to just get bytes (simulating minimal get) or simply abort
	resp, err = client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		ct := strings.ToLower(resp.Header.Get("Content-Type"))
		return strings.Contains(ct, "application/pdf")
	}

	return false
}