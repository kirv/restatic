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
    "strings"

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
	IsDir   bool                // ordinary directory, possibly via symlink
    // thinobject (tob) fields:
    IsFile          bool        // ordinary file, possibly via symlink
    IsList          bool        // ordinary file with @ prefix, treat as list of lines
    IsMap           bool        // ordinary file with % prefix, treat as map or dictionary of key value lines
    IsJSON          bool        // ordinary file with .jsn suffix, assume to be JSON
    IsSymlink       bool        // symlink, could resolve to a file or directory or not
    IsDeclaration   bool        // non-resolving symlink, but not a symvar
    IsType          bool        // non-resolving symlink where name has prefix ^NAME
    IsSymvar        bool        // non-resolving symlink where value has prefix =
    IsParameters    bool        // symvar where the name has suffix :
    Value           string      // symvar value with prefix = removed, or @LIST contents, or %MAP contents
    IsObject        bool        // an Object is a directory which contains symlink ^
    ObjectType      string      // value of object/^ symlink
    IsPrototype     bool
    IsMixin         bool
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
    // each file in the directory is handled here in turn
	lstat, err := entry.Info()
	if err != nil {
		return nil
	}
    var tob fInfo

    // entry includes the base path, the file name, and a link to, I think, the output of os.Stat()
    // Now compare entry.lstat with os.Stat()
    // entry is the output from os.ReadDir() and 'lstat' is really the output from os.Lstat()!

	path, err := filepath.Rel(config.Directory, pwd)
	if err != nil {
		return nil
	}
    name := lstat.Name()
    pathname := path + "/" + name

	stat, err := os.Stat(pathname)
	if err != nil {
        tob.IsSymlink = true
        // The entry is a non-resolving symlink.. if not something more exotic.
        // It could just be a bad symlink, i.e., one where the target file has been
        // renamed or removed, but to thinobject a non-resolving non-symvar symlink
        // is a 'declaration', possibly a 'type' declaration.
        tob.Value, err = os.Readlink(pathname)
    	if err != nil {
    		return nil
    	}
        if strings.HasPrefix(tob.Value, "=") {
            tob.IsSymvar = true
            // a symvar is a non-resolving symlink where value starts with =
            tob.Value = strings.TrimPrefix(tob.Value, "=")
            // fmt.Println("DEBUG SYMVAR:", name, tob.Value)
            if strings.HasSuffix(name, ":") {
                tob.IsParameters = true
                // a symvar with name ending in : is assumed to be a parameters list
                // fmt.Println("DEBUG PARAMETERS:", name, tob.Value)
            }
        } else {
            // non-resolving non-symvar symlinks are 'types' or 'declarations'
            if strings.HasPrefix(name, "^") {
                // non-resolving symlink named ^ or ^foo denotes a thinobject type
                tob.IsType = true
                // fmt.Println("DEBUG TYPE:", name, tob.Value)
            } else {
                tob.IsDeclaration = true
                // fmt.Println("DEBUG DECLARATION:", name, tob.Value)
            }
        }
    } else { // must be either a file or a directory, perhaps through a symlink
        if stat.IsDir() {
            tob.IsDir = true
            // name might be a local directory or might be a symlinked one
            // fmt.Println("DEBUG DIRECTORY:", name)
	        isObject, _ := os.Lstat(pathname + "/^")
            if isObject != nil {
                tob.IsObject = true
                // a directory containing ^ is a Object of this object
                tob.ObjectType, err = os.Readlink(pathname + "/^")
            	if err != nil {
            		return nil
            	}
                // fmt.Println("DEBUG Object:", name, "TYPE:", tob.ObjectType)
            }
        } else {
            tob.IsFile = true
            // name might be a local or symlinked file
            fmt.Println("DEBUG FILE:", name)
            if lstat.Mode()&os.ModeSymlink != 0 {
                fmt.Println("DEBUG SYMLINK-FILE:", name)
                tob.Value, err = os.Readlink(pathname)
            	if err != nil {
                    fmt.Println("DEBUG err:", err)
            		return nil
            	}
            } else {
                tob.Value = ByteCountIEC(lstat.Size())
            } 

            if strings.HasPrefix(name, "@") {
                tob.IsList = true
                // fmt.Println("DEBUG LIST:", name)
            } else if strings.HasPrefix(name, "%") {
                tob.IsMap = true
                // fmt.Println("DEBUG MAP:", name)
            } else if strings.HasSuffix(name, ".json") {
                tob.IsJSON = true
                fmt.Println("DEBUG JSON:", name)
            }
        }
    }

    if tob.IsList || tob.IsMap {
    	b, err := os.ReadFile(pathname)
    	if err != nil {
    		return nil
    	}
        tob.Value = string(b)
    }

	return &fInfo{
		Name:           entry.Name(),
		Mode:           lstat.Mode().String(),
		ModTime:        lstat.ModTime().Format(time.RFC1123),
		Size:           ByteCountIEC(lstat.Size()),
		Path:           path,
		IsDir:          entry.Type().IsDir(),
        IsFile:         tob.IsFile,
        IsList:         tob.IsList,
        IsJSON:         tob.IsJSON,
        IsMap:          tob.IsMap,
        IsSymlink:      tob.IsSymlink,
        Value:          tob.Value,
        IsType:         tob.IsType,
        IsDeclaration:  tob.IsDeclaration,
        IsSymvar:       tob.IsSymvar,
        IsParameters:   tob.IsParameters,
        IsObject:    tob.IsObject,
        ObjectType:  tob.ObjectType,
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
	tmpl := template.Must(template.ParseFiles("templates/thinobject.html", "templates/layout.html"))

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

	info, err := os.Stat(cPath)
	if os.IsNotExist(err) {
		http.NotFound(w, request)
		return
	}

	if err != nil {
		write500(w)
		return
	}

	// if linfo.IsDir() {
	if info.IsDir() {
		iPath := path.Clean(path.Join(cPath, "index.html"))
		iInfo, err := os.Stat(iPath)
		if err == nil && !iInfo.IsDir() {
            // HINT: this point is reached if index.html is found
			writeFile(w, iPath, iInfo)
			return
		} else {
            // HINT: this point is reached when a directory without index.html is entered
			writeDirectory(w, cPath, info)
		}
		return
	}

    // HINT: render contents of this file to the webpage
	writeFile(w, cPath, info)
}

func NewFSHandler() *fsHandler {
	return &fsHandler{}
}
