package e2e

import "os"

func readFile(p string) ([]byte, error)  { return os.ReadFile(p) }
func writeFile(p string, d []byte) error { return os.WriteFile(p, d, 0644) }
