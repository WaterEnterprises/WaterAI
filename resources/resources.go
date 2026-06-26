package resources

import (
	"embed"
	"io"

	"fyne.io/fyne/v2"
)

//go:embed logo.png logo-only.png vscode.png
var assetsFS embed.FS

// Resource names
const (
	ResourceLogo      = "logo.png"
	ResourceLogoOnly  = "logo-only.png"
	ResourceVSCode    = "vscode.png"
)

// GetResource loads a resource from the embedded assets
func GetResource(name string) fyne.Resource {
	data, err := assetsFS.ReadFile(name)
	if err != nil {
		return nil
	}

	return fyne.NewStaticResource(name, data)
}

// GetLogo returns the main logo resource
func GetLogo() fyne.Resource {
	return GetResource(ResourceLogo)
}

// GetLogoOnly returns the logo-only resource (without text)
func GetLogoOnly() fyne.Resource {
	return GetResource(ResourceLogoOnly)
}

// GetVSCodeIcon returns the VS Code icon resource
func GetVSCodeIcon() fyne.Resource {
	return GetResource(ResourceVSCode)
}

// ReadFile reads a file from the embedded assets
func ReadFile(name string) ([]byte, error) {
	return assetsFS.ReadFile(name)
}

// OpenFile opens a file from the embedded assets
func OpenFile(name string) (io.ReadCloser, error) {
	return assetsFS.Open(name)
}
