package hot

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/fsnotify.v1"
)

type Config struct {
	Watch          bool
	BaseName       string
	Dir            string
	Funcs          template.FuncMap
	LeftDelim      string
	RightDelim     string
	FilesExtension []string
	Log            io.Writer
}

type Template struct {
	tpl          *template.Template
	cfg          *Config
	watcher      *fsnotify.Watcher
	closeChannel chan bool
	Out          io.Writer
}

func New(cfg *Config) (*Template, error) {
	leftDelim, rightDelim := getDelims(cfg)
	tmpl := template.New(cfg.BaseName).Delims(leftDelim, rightDelim)
	if cfg.Funcs != nil {
		tmpl = template.New(cfg.BaseName).Funcs(cfg.Funcs).Delims(leftDelim, rightDelim)
	}
	tpl := &Template{
		tpl: tmpl,
		cfg: cfg,
		Out: os.Stdout,
	}
	if cfg.Log != nil {
		tpl.Out = cfg.Log
	}
	tpl.Init()
	err := tpl.Load(cfg.Dir)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

func (t *Template) Init() {
	if t.cfg.Watch {
		watcher, err := fsnotify.NewWatcher()
		t.watcher = watcher
		if err != nil {
			fmt.Fprintln(t.Out, err)
			return
		}
		t.closeChannel = make(chan bool, 1)

		fmt.Fprintln(t.Out, "start watching ", t.cfg.Dir)
		err = watcher.Add(t.cfg.Dir)
		if err != nil {
			watcher.Close()
			fmt.Fprintln(t.Out, err)
			return
		}
		go func() {
			for {
				select {
				case <-t.closeChannel:
					return
				case evt := <-watcher.Events:
					fmt.Fprintf(t.Out, "%s:  reloading... \n", evt.String())
					t.Reload()
				}

			}
		}()
	}
}

func (t *Template) Close() {
	t.closeChannel <- true
	t.watcher.Close()
}

func (t *Template) Reload() {
	tpl := *t.tpl
	leftDelim, rightDelim := getDelims(t.cfg)
	t.tpl = template.New(t.cfg.BaseName).Delims(leftDelim, rightDelim)
	if t.cfg.Funcs != nil {
		t.tpl = template.New(t.cfg.BaseName).Funcs(t.cfg.Funcs).Delims(leftDelim, rightDelim)
	}
	err := t.Load(t.cfg.Dir)
	if err != nil {
		fmt.Fprintln(t.Out, err.Error())
		t.tpl = &tpl
	}
}

func (t *Template) Load(dir string) error {
	fmt.Fprintln(t.Out, "loading...", dir)
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		extension := filepath.Ext(path)
		found := false
		for _, ext := range t.cfg.FilesExtension {
			if ext == extension {
				found = true
				break
			}
		}
		if !found {
			return nil
		}

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		// We remove the directory name from the path
		// this means if we have directory foo, with file bar.tpl
		// full path for bar file foo/bar.tpl
		// we trim the foo part and remain with /bar.tpl
		name := path[len(dir):]

		name = filepath.ToSlash(name)

		name = strings.TrimPrefix(name, "/") // case  we missed the opening slash

		tpl := t.tpl.New(name)
		_, err = tpl.Parse(string(data))
		if err != nil {
			return err
		}
		return nil
	})
}

func (t *Template) Execute(w io.Writer, name string, ctx interface{}) error {
	return t.tpl.ExecuteTemplate(w, name, ctx)
}

func getDelims(cfg *Config) (leftDelim string, rightDelim string) {
	if cfg.LeftDelim != "" {
		leftDelim = cfg.LeftDelim
	} else {
		leftDelim = "{{"
	}
	if cfg.RightDelim != "" {
		rightDelim = cfg.RightDelim
	} else {
		rightDelim = "}}"
	}
	return
}
