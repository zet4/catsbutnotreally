package main

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file3 := &embedded.EmbeddedFile{
		Filename:    `css/style.css`,
		FileModTime: time.Unix(1481917340, 0),
		Content:     string("html {\n    height: 100%;\n}\n\nbody {\n\tcolor: white;\n\twidth: 95%;\n\theight: 95%;\n\tbackground-color: rgb(24, 20, 21);\n\tbackground-size:     contain;\n    background-repeat:   no-repeat;\n    background-position: center center; \n}\n\nh1 {\n\tfont-size: 6em;\n\tfont-family: monospace;\n\ttext-align: center;\n    left: 0;\n    line-height: 200px;\n    margin: auto;\n    margin-top: -100px;\n    position: absolute;\n    top: 50%;\n    width: 100%;\n\ttext-shadow: 2px 0 0 #000, -2px 0 0 #000, 0 2px 0 #000, 0 -2px 0 #000, 1px 1px #000, -1px -1px 0 #000, 1px -1px 0 #000, -1px 1px 0 #000;\n}"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    `index.html`,
		FileModTime: time.Unix(1481917457, 0),
		Content:     string("<!DOCTYPE html>\n<html>\n\n<head>\n    <meta charset=\"UTF-8\">\n    <meta content=\"width=device-width, initial-scale=1.0, maximum-scale=1, user-scalable=no\" name=\"viewport\">\n    <title>Cats but not really</title>\n    <link rel=\"stylesheet\" href=\"css/style.css\">\n</head>\n\n<body>\n    <script src=\"js/index.js\"></script>\n</body>\n\n</html>"),
	}
	file6 := &embedded.EmbeddedFile{
		Filename:    `js/index.js`,
		FileModTime: time.Unix(1481918129, 0),
		Content:     string("document.body.onload = addElement;\n\nfunction addElement() {\n\tvar evtSource = new EventSource(\"/events\");\n\tvar totalCount = document.createElement(\"h1\")\n\tdocument.body.appendChild(totalCount)\n\tevtSource.addEventListener(\"stats\", function(e) {\n\t\tvar obj = JSON.parse(e.data);\n\t\ttotalCount.innerHTML = obj.total;\n\t\tdocument.body.style.backgroundImage = \"url('\" +  obj.last_image + \"')\";\n\t}, false);\n}"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   ``,
		DirModTime: time.Unix(1481917491, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file4, // index.html

		},
	}
	dir2 := &embedded.EmbeddedDir{
		Filename:   `css`,
		DirModTime: time.Unix(1481917340, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file3, // css/style.css

		},
	}
	dir5 := &embedded.EmbeddedDir{
		Filename:   `js`,
		DirModTime: time.Unix(1481917339, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file6, // js/index.js

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{
		dir2, // css
		dir5, // js

	}
	dir2.ChildDirs = []*embedded.EmbeddedDir{}
	dir5.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`web`, &embedded.EmbeddedBox{
		Name: `web`,
		Time: time.Unix(1481917491, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"":    dir1,
			"css": dir2,
			"js":  dir5,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"css/style.css": file3,
			"index.html":    file4,
			"js/index.js":   file6,
		},
	})
}
