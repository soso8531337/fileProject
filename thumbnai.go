package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var ThumbScript string
const ThumbDir = ".thumbnail"
const ThumSuffix = ".jpg"
var supFileList  = map[string] bool{
	".mp4":true,
	".mkv":true,
	".rmvb":true,
	".mov":true,
}
func makeThumbnail(srcPath string){
	if ThumbScript == ""{
		return
	}
	sInfo, err:= os.Stat(srcPath)
	if err != nil || sInfo.IsDir(){
		return
	}
	//filter file
	fileSuffix := path.Ext(srcPath)
	if fileSuffix == ""{
		return
	}
	if _, ok := supFileList[strings.ToLower(fileSuffix)]; !ok{
		log.Printf("Not Found Suffix:%s\n", fileSuffix)
		return
	}
	//Confirm if exist thumbnail
	thumbPath := path.Clean(path.Dir(srcPath)+"/"+ThumbDir+"/"+path.Base(srcPath)+ThumSuffix)
	tInfo, err := os.Stat(thumbPath)
	if err == nil && tInfo.Size() > 0{
		log.Printf("Exist Thumbnail:%s\n", path.Base(srcPath))
		return
	}
	if _, err := os.Stat(path.Dir(thumbPath)); err != nil {
		os.Mkdir(path.Dir(thumbPath), 0755)
	}
	//Generate thumbnail
	log.Printf("Begin Generate thumbnail %s\n", path.Base(srcPath))
	cmd:= exec.Command(ThumbScript, srcPath, thumbPath)
	if err:= cmd.Run(); err != nil{
		log.Println(err)
		return
	}
	log.Printf("Finish Generate thumbnail %s\n", path.Base(srcPath))
}

func addDirectoryInotify(watcher *fsnotify.Watcher, rootDir string){
	rInfo, err := os.Stat(rootDir)
	if err != nil || !rInfo.IsDir(){
		return
	}
	baseName := path.Base(rootDir)
	if strings.HasPrefix(baseName, "."){
		log.Printf("Ignrore %s\n", rootDir)
		return
	}

	err = watcher.Add(rootDir)
	if err != nil{
		log.Println(err)
		return
	}
	//add basedir sub dirctory
	err = filepath.Walk(rootDir, func(subPath string, info os.FileInfo, err error) error{
		sInfo, err := os.Stat(subPath)
		if err != nil{
			log.Println(err)
			return nil
		}
		if strings.HasPrefix(path.Base(subPath), "."){
			log.Printf("Ignrore %s\n", subPath)
			return nil
		}
		if strings.Contains(subPath, ThumbDir){
			log.Printf("Ignore Thumbnail Directory\n")
			return nil
		}
		if sInfo.IsDir(){
			watcher.Add(subPath)
			log.Printf("Add Directory OK: %s\n", subPath)
		}else{
			//make thumbnail
			go makeThumbnail(subPath)
		}
		return nil
	})
	if err != nil{
		log.Println(err)
		return
	}
	log.Printf("Add Directory Finish\n")
	/*
	baseFList, err:=ioutil.ReadDir(rootDir)
	if err != nil{
		log.Println(err)
		return
	}
	for _, node:= range baseFList{
		if node.IsDir(){
			sDir := filepath.Clean(rootDir+"/"+node.Name())
			log.Printf("Add Directory OK: %s\n", sDir)
			watcher.Add(sDir)
			addDirectoryInotify(watcher, sDir)
		}
	}
	*/
}

func removeDirectoryInotify(watcher *fsnotify.Watcher, rootDir string){
	watcher.Remove(rootDir)
}
func wathDirectory(baseDir string){
	if len(baseDir) == 0{
		log.Printf("Dir is Empty\n")
		return
	}
	fileInfo, err:= os.Stat(baseDir)
	if err != nil || !fileInfo.IsDir(){
		log.Println(err)
		return
	}
	//inotify
	watch, err:= fsnotify.NewWatcher()
	if err != nil{
		log.Println(err)
		return
	}
	defer watch.Close()
	addDirectoryInotify(watch, baseDir)
	renameFlag := false
	for{
		select {
		case ev := <-watch.Events:
		{
			if ev.Op & fsnotify.Create == fsnotify.Create{
				log.Printf("Create File:%s\n", ev.Name)
				addDirectoryInotify(watch, ev.Name)
				if renameFlag{
					rInfo, err:= os.Stat(ev.Name)
					if err == nil && !rInfo.IsDir(){
						go makeThumbnail(ev.Name)
					}
				}
			}
			/*
			if ev.Op & fsnotify.Write == fsnotify.Write{
				log.Printf("Write File %s\n", ev.Name)
			}
			*/
			if ev.Op & fsnotify.Remove == fsnotify.Remove{
				log.Printf("Remove File %s\n", ev.Name)
				removeDirectoryInotify(watch, ev.Name)
			}
			if ev.Op & fsnotify.Rename == fsnotify.Rename{
				log.Printf("Rename File %s\n", ev.Name)
				renameFlag = true
				removeDirectoryInotify(watch, ev.Name)
			}else{
				renameFlag = false
			}
			/*
			if ev.Op & fsnotify.Chmod == fsnotify.Chmod{
				log.Printf("Chmod File %s\n", ev.Name)
			}
			*/
			if ev.Op &fsnotify.CloseWrite == fsnotify.CloseWrite{
				//add file
				go makeThumbnail(ev.Name)
			}
		}
		case err:=<-watch.Errors:
			log.Println(err)
		}
	}
}
