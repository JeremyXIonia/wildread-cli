package parser

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/JeremyXIonia/wildread-cli/models"
)

// ParseEPUB parses an EPUB file.
func ParseEPUB(path string) (*models.Book, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("open epub: %w", err)
	}
	defer zr.Close()

	// 1. Find OPF via container.xml
	var opfPath string
	for _, f := range zr.File {
		if f.Name == "META-INF/container.xml" {
			data, err := readZipFile(f)
			if err != nil {
				return nil, err
			}
			var cont struct {
				Rootfiles struct {
					Rootfile struct {
						FullPath string `xml:"full-path,attr"`
					} `xml:"rootfile"`
				} `xml:"rootfiles"`
			}
			if err := xml.Unmarshal(data, &cont); err != nil {
				return nil, fmt.Errorf("parse container: %w", err)
			}
			opfPath = cont.Rootfiles.Rootfile.FullPath
			break
		}
	}
	if opfPath == "" {
		return nil, fmt.Errorf("opf not found in container")
	}

	// 2. Read OPF
	var opfData []byte
	for _, f := range zr.File {
		if f.Name == opfPath {
			data, err := readZipFile(f)
			if err != nil {
				return nil, err
			}
			opfData = data
			break
		}
	}
	if opfData == nil {
		return nil, fmt.Errorf("opf not found: %s", opfPath)
	}

	type opfPackage struct {
		Metadata struct {
			Title   string `xml:"title"`
			Creator string `xml:"creator"`
		} `xml:"metadata"`
		Manifest struct {
			Items []struct {
				ID        string `xml:"id,attr"`
				Href      string `xml:"href,attr"`
				MediaType string `xml:"media-type,attr"`
			} `xml:"item"`
		} `xml:"manifest"`
		Spine struct {
			Items []struct {
				IDRef string `xml:"idref,attr"`
			} `xml:"itemref"`
		} `xml:"spine"`
	}
	var pkg opfPackage
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		return nil, fmt.Errorf("parse opf: %w", err)
	}

	opfDir := filepath.Dir(opfPath)

	idToHref := map[string]string{}
	for _, it := range pkg.Manifest.Items {
		idToHref[it.ID] = it.Href
	}

	// 3. Extract chapters in spine order
	var chapters []models.Chapter
	for _, ref := range pkg.Spine.Items {
		href, ok := idToHref[ref.IDRef]
		if !ok {
			continue
		}
		lower := strings.ToLower(href)
		if !strings.Contains(lower, ".xhtml") && !strings.Contains(lower, ".html") {
			continue
		}
		fullPath := filepath.ToSlash(filepath.Join(opfDir, href))
		var fileData []byte
		for _, f := range zr.File {
			if f.Name == fullPath {
				d, err := readZipFile(f)
				if err != nil {
					return nil, err
				}
				fileData = d
				break
			}
		}
		if fileData == nil {
			continue
		}
		title, paras := HTMLToText(string(fileData))
		content := strings.Join(paras, "\n\n")
		chapters = append(chapters, models.Chapter{Title: title, Content: content})
	}

	return &models.Book{
		Title:    pkg.Metadata.Title,
		Author:   pkg.Metadata.Creator,
		Format:   "epub",
		Chapters: chapters,
	}, nil
}

// readZipFile reads the full contents of a zip file entry.
func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, rc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
