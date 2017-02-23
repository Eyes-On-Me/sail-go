package fs

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// 删除
func Delete(file string) error {
	if IsExist(file) {
		return os.RemoveAll(file)
	}
	return nil
}

// 改名
func Rename(old_name, new_name string) error {
	return os.Rename(old_name, new_name)
}

// 是否为文件
func IsFile(file string) bool {
	f, e := os.Stat(file)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

// 是否文件存在
func IsExist(path string) bool {
	_, e := os.Stat(path)
	return e == nil || os.IsExist(e)
}

// 是否目录
func IsDir(dir string) bool {
	f, e := os.Stat(dir)
	if e != nil {
		return false
	}
	return f.IsDir()
}

// 创建文件
func FileNew(path string, name string) (string, error) {
	src := path + name + "/"
	if IsExist(src) {
		return src, nil
	}
	if e := os.MkdirAll(src, 0755); e != nil {
		if os.IsPermission(e) {
			fmt.Println("你不够权限创建文件")
		}
		return "", e
	}
	return src, nil
}

// 获取文件修改时间
func FileGetModifiedTime(file string) (int64, error) {
	f, e := os.Stat(file)
	if e != nil {
		return 0, e
	}
	return f.ModTime().Unix(), nil
}

// 获取文件大小
func FileGetSize(file string) (int64, error) {
	f, e := os.Stat(file)
	if e != nil {
		return 0, e
	}
	return f.Size(), nil
}

// 追加内容到文件
func FilePutS(file string, content string) (int, error) {
	fs, e := os.Create(file)
	if e != nil {
		return 0, e
	}
	defer fs.Close()
	return fs.WriteString(content)
}

// 写文件
// 0755
func FileWrite(file string, bytes []byte, perm os.FileMode) error {
	return ioutil.WriteFile(file, bytes, perm)
}

// 写文件
func FileWriteS(file string, str string, perm os.FileMode) error {
	return ioutil.WriteFile(file, []byte(str), perm)
}

// 读文件
func FileRead(file string) ([]byte, error) {
	if !IsFile(file) {
		return []byte(""), os.ErrNotExist
	}
	b, e := ioutil.ReadFile(file)
	if e != nil {
		return []byte(""), e
	}
	return b, nil
}

// 读文件
func FileReadS(file string) (string, error) {
	b, e := FileRead(file)
	return string(b), e
}

// 拷贝
func FileCopy(src_file string, dst_file string, perm os.FileMode) error {
	data, e := ioutil.ReadFile(src_file)
	if e != nil {
		return e
	}
	e = ioutil.WriteFile(dst_file, data, perm)
	if e != nil {
		return e
	}
	return nil
}

// 搜索文件
func FileSearch(file string, paths ...string) (full_path []string, err error) {
	file = filepath.Base(file)
	for _, path := range paths {
		has_path := filepath.Join(path, file)
		if IsExist(has_path) {
			full_path = append(full_path, has_path)
		}
	}
	if len(full_path) <= 0 {
		err = errors.New("Search file not found in paths!")
	}
	return
}

// 创建目录
func PathNew(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func DirStat(rootPath string, includeDir ...bool) ([]string, error) {
	if !IsDir(rootPath) {
		return nil, errors.New("not a directory or does not exist: " + rootPath)
	}

	isIncludeDir := false
	if len(includeDir) >= 1 {
		isIncludeDir = includeDir[0]
	}
	return statDir(rootPath, "", isIncludeDir, false)
}

func DirCopy(srcPath, destPath string, filters ...func(filePath string) bool) error {
	if IsExist(destPath) {
		return errors.New("file or directory alreay exists: " + destPath)
	}
	err := os.MkdirAll(destPath, os.ModePerm)
	if err != nil {
		return err
	}
	infos, err := DirStat(srcPath, true)
	if err != nil {
		return err
	}
	var filter func(filePath string) bool
	if len(filters) > 0 {
		filter = filters[0]
	}
	for _, info := range infos {
		if filter != nil && filter(info) {
			continue
		}
		curPath := path.Join(destPath, info)
		if strings.HasSuffix(info, "/") {
			err = os.MkdirAll(curPath, os.ModePerm)
		} else {
			err = DirCopy2(path.Join(srcPath, info), curPath)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func DirCopy2(src, dest string) error {
	si, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if si.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(target, dest)
	}
	sr, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sr.Close()
	dw, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dw.Close()
	if _, err = io.Copy(dw, sr); err != nil {
		return err
	}
	if err = os.Chtimes(dest, si.ModTime(), si.ModTime()); err != nil {
		return err
	}
	return os.Chmod(dest, si.Mode())
}

func DirGetFilesBySuffix(dirPath, suffix string) ([]string, error) {
	if !IsExist(dirPath) {
		return nil, fmt.Errorf("given path does not exist: %s", dirPath)
	} else if IsFile(dirPath) {
		return []string{dirPath}, nil
	}
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	fis, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(fis))
	for _, fi := range fis {
		if strings.HasSuffix(fi.Name(), suffix) {
			files = append(files, path.Join(dirPath, fi.Name()))
		}
	}
	return files, nil
}

// 列出某目录下文件列表
func DirGetList(dir string) []os.FileInfo {
	files, _ := ioutil.ReadDir(dir)
	return files
}

func DirGetSubDirs(rootPath string) ([]string, error) {
	if !IsDir(rootPath) {
		return nil, errors.New("not a directory or does not exist: " + rootPath)
	}
	return statDir(rootPath, "", true, true)
}

// 获取本应用路径
func AppPathGet() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

// 获取本应用所在目录
func AppDirGet() string {
	return filepath.Dir(AppPathGet())
}

// 获取用户主目录
func UserPathGet() (home string, err error) {
	if runtime.GOOS == "windows" {
		home = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
	} else {
		home = os.Getenv("HOME")
	}
	if home== "" {
		return "", errors.New("Cannot specify home directory because it's empty")
	}
	return home, nil
}

func statDir(dirPath, recPath string, includeDir, isDirOnly bool) ([]string, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	fis, err := dir.Readdir(0)
	if err != nil {
		return nil, err
	}
	statList := make([]string, 0)
	for _, fi := range fis {
		if strings.Contains(fi.Name(), ".DS_Store") {
			continue
		}
		relPath := path.Join(recPath, fi.Name())
		curPath := path.Join(dirPath, fi.Name())
		if fi.IsDir() {
			if includeDir {
				statList = append(statList, relPath+"/")
			}
			s, err := statDir(curPath, relPath, includeDir, isDirOnly)
			if err != nil {
				return nil, err
			}
			statList = append(statList, s...)
		} else if !isDirOnly {
			statList = append(statList, relPath)
		}
	}
	return statList, nil
}
