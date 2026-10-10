package pathlayout
import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)
type Layout struct {
	CertRoot   string
	CARoot     string
	ServerRoot string
}
var dirNoRe = regexp.MustCompile(`^[0-9]{4}$`)
func New(certRoot, caRoot, serverRoot string) *Layout {
	return &Layout{
		CertRoot:   filepath.Clean(certRoot),
		CARoot:     filepath.Clean(caRoot),
		ServerRoot: filepath.Clean(serverRoot),
	}
}
func (l *Layout) AllocServerDir() (string, error) {
	if err := os.MkdirAll(l.ServerRoot, 0750); err != nil {
		return "", err
	}
	entries, err := os.ReadDir(l.ServerRoot)
	if err != nil {
		return "", err
	}
	max := -1
	for _, e := range entries {
		if e.IsDir() && dirNoRe.MatchString(e.Name()) {
			n, _ := strconv.Atoi(e.Name())
			if n > max {
				max = n
			}
		}
	}
	dirNo := fmt.Sprintf("%04d", max+1)
	dir := filepath.Join(l.ServerRoot, dirNo)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", err
	}
	return dirNo, nil
}
func (l *Layout) ServerDir(dirNo string) (string, error) {
	if !dirNoRe.MatchString(dirNo) {
		return "", fmt.Errorf("invalid dir_no")
	}
	return filepath.Join(l.ServerRoot, dirNo), nil
}
