package fileutil

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	// 默认的文件权限
	DefaultFileMode os.FileMode = os.ModePerm

	// linux下需加上O_WRONLY或是O_RDWR
	DefaultFileFlag int = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

//构建一个每日写日志文件的写入器
func OpenFile(pathfile string, fileflag int, filemode os.FileMode) (wc *os.File, err error) {
	pathfile = strings.Replace(filepath.Clean(pathfile), "\\", "/", -1)
	dir := path.Dir(pathfile)
	if _, err = os.Stat(dir); err != nil && !os.IsExist(err) {
		if !os.IsNotExist(err) {
			return nil, err
		}
		if err = os.MkdirAll(dir, filemode); err != nil {
			return nil, err
		}
		if _, err = os.Stat(dir); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(pathfile, fileflag, filemode)
}

func DirExists(dir string) bool {
	d, e := os.Stat(dir)
	switch {
	case e != nil:
		return false
	case !d.IsDir():
		return false
	}
	return true
}

func FileExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

//eg:filename=`../a/test/aa.txt` newExt=`.csv` -->return=`../a/test/aa.csv`
func ChangeFileExt(filename, newExt string) string {
	filename = strings.Replace(filepath.Clean(filename), "\\", "/", -1)
	file := path.Base(filename)
	file = strings.TrimSuffix(file, path.Ext(file)) + newExt
	dir := path.Dir(filename)
	return path.Join(dir, file)
}
func PathJoin(dir, filename string) string {
	return strings.Replace(path.Join(filepath.Clean(dir), filename), "\\", "/", -1)
}

//将路径转成统一使用的路径格式
func TransPath(path string) string {
	return strings.Replace(filepath.Clean(path), "\\", "/", -1)
}
