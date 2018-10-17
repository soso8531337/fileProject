package main

import (
	"encoding/json"
	"flag"
	"path"
	"strings"

	//	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
//	"strconv"
)

type fileInfo struct{
	FileName string 	`json:"name"`
	IsDir	 bool		`json:"isdir"`
	FileSize int64		`json:"size"`
	Thunbnail string	`json:"thumbnail"`
}
type fileList struct{
	Errno int		`json:"error"`
	Count int		`json:"count"`
	Data []fileInfo	`json:"data"`
}

func getThunmbnailPath(oriFile string) string{
	bDir := path.Dir(oriFile)
	thubFile := path.Clean(ThumbDir+"/"+path.Base(oriFile)+ThumSuffix)
	thumInfo, err:= os.Stat(bDir+"/"+thubFile)
	if err != nil || thumInfo.Size() == 0{
		return ""
	}
	return thubFile
}
func genFilePath(basePath, filter, filePath string) string{
	if len(filePath) < len(filter){
		return basePath
	}
	return basePath+filePath[len(filter):]
}
func fileListFunc(w http.ResponseWriter, r *http.Request){
	//join the file path
	decURI, err := url.QueryUnescape(r.RequestURI)
	fileFullName:= genFilePath(FileListDirectory, FilterPath, decURI)
	fileFullName = filepath.Clean(fileFullName)
	if len(fileFullName) == 0 {
		log.Printf("%s Not Exist\n", decURI)
		http.Error(w, "Path is Empty", 400)
		return
	}
	nodeInfo, err:= os.Stat(fileFullName)
	if err != nil{
		log.Println(err)
		http.Error(w, "File Not Exist", 400)
		return
	}

	if nodeInfo.IsDir(){
		files, err := ioutil.ReadDir(fileFullName)
		if err != nil{
			log.Println(err)
			http.Error(w, "Read Directory Failed", 400)
			return
		}
		var resData fileList

		resData.Count = len(files)
		for _, pFIle:= range files{
			var pFileInfo fileInfo
			pFileInfo.FileName = filepath.Clean(r.RequestURI+"/"+pFIle.Name())
			if strings.HasPrefix(path.Base(pFileInfo.FileName), "."){
				continue
			}
			pFileInfo.IsDir = pFIle.IsDir()
			if pFileInfo.IsDir{
				pFileInfo.Thunbnail = ""
			}else{
				t := getThunmbnailPath(fileFullName+"/"+pFIle.Name())
				if t == ""{
					pFileInfo.Thunbnail = ""
				}else{
					pFileInfo.Thunbnail = path.Clean(r.RequestURI+"/"+ t)
				}
			}
			pFileInfo.FileSize = pFIle.Size()
			resData.Data = append(resData.Data, pFileInfo)
		}
		resData.Errno = 0
		d, err := json.Marshal(resData)
		if err != nil{
			log.Println(err)
			http.Error(w, "Json Format Error", 400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(d)
		return
	}else{
		//download file
	/*
		Openfile, err := os.Open(fileFullName)
		defer Openfile.Close()
		if err != nil{
			log.Println(err)
			http.Error(w, "Open file Failed", 400)
			return
		}
		FileHeader := make([]byte, 512)
		Openfile.Read(FileHeader)
		FileContentType := http.DetectContentType(FileHeader)
		FileSize := strconv.FormatInt(nodeInfo.Size(), 10)
		w.Header().Set("Content-Type", FileContentType)
		w.Header().Set("Content-Length", FileSize)
		Openfile.Seek(0, 0)
		io.Copy(w, Openfile)
	*/
		http.ServeFile(w, r, fileFullName)
		return
	}
}
const FilterPath string = "/download"
var FileListDirectory string
func main(){
	baseDir := flag.String("d", "/home/zhangwei/filebrowser/webdav/chenjintao/video/", "File list Base Directory")
	listenPort := flag.String("p", "10090", "Listen Port")
	thumbSh := flag.String("m", "ffmpegthumbnailer", "Generate thumbnail script")

	flag.Parse()
	log.SetFlags(log.Flags() | log.Lshortfile)
	FileListDirectory = *baseDir
	ThumbScript = *thumbSh
	log.Printf("Dir:%s Port:%s\n", *baseDir, *listenPort)
	go wathDirectory(FileListDirectory)
	http.HandleFunc("/", fileListFunc)
	http.ListenAndServe(":"+*listenPort, nil)
}
