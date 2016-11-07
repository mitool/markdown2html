package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"html/template"

	"github.com/admpub/log"
	md2html "github.com/russross/blackfriday"
	"github.com/webx-top/webx/lib/com"
)

func md2htmlExtension(fileName string, markdownExts ...string) string {
	for _, ext := range markdownExts {
		if strings.HasSuffix(fileName, ext) {
			fileName = strings.TrimSuffix(fileName, ext)
		}
	}
	fileName += `.html`
	return fileName
}

type TMPLData struct {
	MarkdownHTML template.HTML
}

func main() {
	watchDir := flag.String(`watchDir`, `./markdown`, `-watchDir ./markdown`)
	markdownExt := flag.String(`markdownExt`, `.md,.markdown`, `-markdownExt .md,.markdown`)
	htmlTemplate := flag.String(`htmlTemplate`, `./templates/default.html`, `-htmlTemplate ./templates/default.html`)
	outputDir := flag.String(`outputDir`, `./html`, `-outputDir ./html`)
	flag.Parse()

	log.SetFatalAction(log.ActionExit)
	log.Sync(true)

	savePath, err := filepath.Abs(*outputDir)
	if err != nil {
		log.Fatal(err)
		return
	}
	watchPath, err := filepath.Abs(*watchDir)
	if err != nil {
		log.Fatal(err)
		return
	}
	markdownExts := strings.Split(*markdownExt, `,`)
	tmpl := template.New(filepath.Base(*htmlTemplate))
	content, err := ioutil.ReadFile(*htmlTemplate)
	if err != nil {
		log.Fatal(err)
		return
	}
	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		log.Fatal(err)
		return
	}
	actions := com.MonitorEvent{
		Modify: func(name string) {
			filePath := strings.TrimPrefix(name, watchPath)
			filePath = filepath.Join(savePath, filePath)
			filePath = md2htmlExtension(filePath, markdownExts...)

			b, err := ioutil.ReadFile(name)
			if err != nil {
				log.Error(err)
				return
			}
			b = md2html.MarkdownCommon(b)
			err = os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
			if err != nil {
				log.Error(err)
				return
			}
			fp, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			if err != nil {
				log.Error(err)
				return
			}
			defer fp.Close()
			err = tmpl.Execute(fp, &TMPLData{
				MarkdownHTML: template.HTML(string(b)),
			})
			if err != nil {
				log.Error(err)
			} else {
				log.Info(`成功生成html文件：` + filePath)
			}
		},
		Delete: func(name string) {
			filePath := strings.TrimPrefix(name, watchPath)
			filePath = filepath.Join(savePath, filePath)
			filePath = md2htmlExtension(filePath, markdownExts...)
			err := os.Remove(filePath)
			if err != nil {
				log.Error(err)
			} else {
				log.Info(`成功删除html文件：` + name)
			}
		},
	}
	actions.Create = actions.Modify
	actions.Watch(watchPath)

	err = filepath.Walk(watchPath, func(f string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		actions.Modify(f)
		return nil
	})
	if err != nil {
		log.Error(err)
	}
	<-make(chan int)
}
