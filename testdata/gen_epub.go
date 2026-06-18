//go:build ignore

package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	out := filepath.Join(".", "testdata", "sample.epub")
	os.MkdirAll(filepath.Dir(out), 0755)

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	writestr(w, "META-INF/container.xml", `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`)

	writestr(w, "OEBPS/content.opf", `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="bid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="bid">urn:uuid:00000000-0000-0000-0000-000000000001</dc:identifier>
    <dc:title>测试书</dc:title>
    <dc:creator>测试作者</dc:creator>
    <dc:language>zh</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="ch1.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch2" href="ch2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
    <itemref idref="ch2"/>
  </spine>
</package>`)

	writestr(w, "OEBPS/ch1.xhtml", `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>第一章</title></head>
<body><h1>第一章 开始</h1><p>第一段内容。</p><p>第二段内容。</p></body></html>`)

	writestr(w, "OEBPS/ch2.xhtml", `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>第二章</title></head>
<body><h1>第二章 继续</h1><p>这是第二章的内容。</p></body></html>`)

	w.Close()

	if err := os.WriteFile(out, buf.Bytes(), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Created", out)
}

func writestr(w *zip.Writer, name, content string) {
	f, err := w.Create(name)
	if err != nil {
		panic(err)
	}
	f.Write([]byte(content))
}
