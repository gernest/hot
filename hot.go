package hot

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

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
}

type Template struct {
	tpl *template.Template
	cfg *Config
	Out io.Writer
}

func New(cfg *Config) (*Template, error) {
	leftDelim := "{{"
	rightDelim := "}}"
	if cfg.LeftDelim != "" {
		leftDelim = cfg.LeftDelim
	}
	if cfg.RightDelim != "" {
		rightDelim = cfg.RightDelim
	}
	tmpl := template.New(cfg.BaseName).Delims(leftDelim, rightDelim)
	if cfg.Funcs != nil {
		tmpl = template.New(cfg.BaseName).Funcs(cfg.Funcs).Delims(leftDelim, rightDelim)
	}
	tpl := &Template{
		tpl: tmpl,
		cfg: cfg,
		Out: os.Stdout,
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
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				fmt.Fprintln(t.Out, err)
				return
			}
			err = filepath.Walk(t.cfg.Dir, func(fPath string, info os.FileInfo, ferr error) error {
				if ferr != nil {
					return ferr
				}
				if info.IsDir() {
					fmt.Fprintln(t.Out, "start watching ", fPath)
					return watcher.Add(fPath)
				}
				return nil
			})
			if err != nil {
				fmt.Fprintln(t.Out, err)
				return
			}

			for {
				select {
				case <-c:
					watcher.Close()
					fmt.Println("shutting down hot templates... done")
					os.Exit(0)
				case evt := <-watcher.Events:
					fmt.Fprintf(t.Out, "%s:  reloading... \n", evt.String())
					t.Reload()
				}
			}
		}()
	}
}

func (t *Template) Reload() {
	tpl := *t.tpl
	t.tpl = template.New(t.cfg.BaseName)
	if t.cfg.Funcs != nil {
		t.tpl = template.New(t.cfg.BaseName).Funcs(t.cfg.Funcs)
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
