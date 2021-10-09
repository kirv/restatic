package server

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/relogHQ/restatic/config"
)

type fsHandler struct{}
type dirlist struct {
	Files   []*fInfo
	DirInfo *dInfo
}
type fInfo struct {
	Name    string
	Mode    string
	ModTime string
	Size    string
	Path    string
	IsDir   bool
}
type dInfo struct {
	Name string
	Path string
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func toFInfo(entry os.DirEntry, pwd string) *fInfo {
	info, err := entry.Info()
	if err != nil {
		return nil
	}

	path, err := filepath.Rel(config.Directory, pwd)
	if err != nil {
		return nil
	}

	if entry.Type().IsDir() {
        fmt.Println(info)
		// info.Name = info.Name + "/"
    }

	return &fInfo{
		Name:    entry.Name(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime().Format(time.RFC1123),
		Size:    ByteCountIEC(info.Size()),
		Path:    path,
		IsDir:   entry.Type().IsDir(),
	}
}

func toFInfos(infos []os.DirEntry, pwd string) []*fInfo {
	fInfos := make([]*fInfo, len(infos))
	for i, info := range infos {
		fInfos[i] = toFInfo(info, pwd)
	}
	return fInfos
}

func toDInfo(info os.FileInfo, pwd string) *dInfo {
	rPath, err := filepath.Rel(config.Directory, pwd)
	if err != nil {
		return nil
	}

	pPath := filepath.Dir(rPath)

	if rPath == "." {
		rPath = filepath.Base(config.Directory)
	} else {
		rPath = path.Join(filepath.Base(config.Directory), rPath)
	}

	return &dInfo{
		Name: rPath,
		Path: pPath,
	}
}

func write500(w http.ResponseWriter) {
	http.Error(w, http.StatusText(500), 500)
}

func writeDirectory(w http.ResponseWriter, path string, dirInfo os.FileInfo) {
	tmpl := template.Must(template.ParseFiles("templates/dir.html", "templates/layout.html"))

	files, err := os.ReadDir(path)
	if err != nil {
		write500(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, dirlist{
		Files:   toFInfos(files, path),
		DirInfo: toDInfo(dirInfo, path),
	})
}

func writeFile(w http.ResponseWriter, path string, info os.FileInfo) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		write500(w)
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		write500(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func (f fsHandler) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	baseDir := config.Directory
	if _, err := os.Stat(baseDir); err != nil {
		log.Fatal(err)
	}

	cPath := path.Clean(path.Join(baseDir, request.URL.Path))

	linfo, err := os.Lstat(cPath)
	if os.IsNotExist(err) {
		http.NotFound(w, request)
		return
	}

	info, err := os.Stat(cPath)
	if os.IsNotExist(err) {
		http.NotFound(w, request)
		return
	}

	if err != nil {
		write500(w)
		return
	}

	if linfo.IsDir() {
		iPath := path.Clean(path.Join(cPath, "index.html"))
		iInfo, err := os.Stat(iPath)
		if err == nil && !iInfo.IsDir() {
			writeFile(w, iPath, iInfo)
			return
		} else {
			writeDirectory(w, cPath, info)
		}
		return
	}

	writeFile(w, cPath, info)
}

func NewFSHandler() *fsHandler {
	return &fsHandler{}
}
