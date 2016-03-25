# Hot [![Build Status](https://travis-ci.org/gernest/hot.svg)](https://travis-ci.org/gernest/hot)

Hot is a library for rendering hot golang templates. This means with `hot` you won't need to reload your application everytime you edit your templates.Hot watches for file changes in your templates directory and reloads everytime you make changes to your files.

hot renders go templates using `html/template` package.

# Installation

	go get github.com/gernest/hot

# Usage

Just pass the configuration object to `hot.New`

```go
package main

import (
    "os"

    "github.com/gernest/hot"
)

func main() {
    config := &hot.Config{
        Watch:          true,
        BaseName:       "hot",
        Dir:            "fixtures",
        FilesExtension: []string{".tpl", ".html", ".tmpl"},
    }

    tpl, err := hot.New(config)
    if err != nil {
        panic(err)
    }

    // execute the template named "hello.tpl
    tpl.Execute(os.Stdout, "hello.tpl", nil)
}

```

Note that the fixtures directory should exist and there is a template file `hello.tpl` in it , `hot` will be watching any changes to the files inside this directory.

# configuration

The `hot.Config` object is used to configure the hot template

property| details
--------|---------
Watch| If set to true, the hot reload is enabled
BaseName| Is the root template name, e.g "base", "hot" etc
Dir| The directory in which you keep your templates
FilesExtension| Supported file extensions. These are the ones parsed in the template.
Funcs| (optional) A map of names to functions that can be used inside your templates. ([more information](https://golang.org/pkg/text/template/#FuncMap))
LeftDelim| left template delimiter e.g {{
RightDelim| rignt template delimiter e.g }}



# Contributing

Start with clicking the star button to make the author and his neighbors happy. Then fork the repository and submit a pull request for whatever change you want to be added to this project.

If you have any questions, just open an issue.

# Author
Geofrey Ernest

Twitter  : [@gernesti](https://twitter.com/gernesti)


# Licence

This project is released under the MIT licence. See [LICENCE](LICENCE) for more details.
