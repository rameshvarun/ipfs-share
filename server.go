package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/vincent-petithory/dataurl"

	ipfs "github.com/ipfs/go-ipfs-api"
)

type Paste struct {
	Content string `form:"content" binding:"required"`
}

type Image struct {
	DataURL  string `form:"dataurl" binding:"required"`
	FileName string `form:"filename" binding:"required"`
}

type UploadFile struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

var shell *ipfs.Shell

func main() {
	// Parse command-line flags.
	gatewayURL := flag.String("gateway", "https://ipfs.io", "The HTTP gateway used in shared file links.")
	port := flag.Int("port", 3000, "The port number to run the server on.")
	hostname := flag.String("hostname", "0.0.0.0", "The hostname to run under.")
	daemon := flag.String("daemon", "/ip4/127.0.0.1/tcp/5001", "The address of the Daemon API,")
	flag.Parse()

	// Connect to the IPFS daemon.
	shell = ipfs.NewShell(*daemon)

	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".html"},
		Directory:  ".",
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "index", nil)
	})

	m.Post("/paste", binding.Form(Paste{}), func(paste Paste, ferr binding.Errors, r render.Render) {
		// Load paste content into a buffer.
		buf := new(bytes.Buffer)
		buf.WriteString(paste.Content)

		// Add buffer to IPFS
		hash, err := shell.Add(buf)
		if err != nil {
			log.Fatal(err)
		}

		// Create a share URL and return
		url := *gatewayURL + "/ipfs/" + hash + "/?filename=paste.txt"
		response := map[string]interface{}{"url": url}
		r.JSON(200, response)
	})

	m.Post("/image", binding.Form(Image{}), func(image Image, ferr binding.Errors, r render.Render) {
		// Parse Data URL
		dataURL, err := dataurl.DecodeString(image.DataURL)
		if err != nil {
			log.Fatal(err)
		}

		// Load paste content into a buffer.
		buf := new(bytes.Buffer)
		buf.Write(dataURL.Data)

		// Write content to a file
		hash, err := shell.Add(buf)

		// Create a share URL and return
		url := *gatewayURL + "/ipfs/" + hash + "?filename=" + image.FileName
		response := map[string]interface{}{"url": url}
		r.JSON(200, response)
	})

	m.Post("/upload", func(req *http.Request, r render.Render) {
		// Load file information from the form data.
		infile, header, err := req.FormFile("file")
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()

		// Read file content into buffer.
		log.Println("Reading file content into buffer...")
		buf := new(bytes.Buffer)
		buf.ReadFrom(infile)

		// Add to IPFS
		hash, err := shell.Add(buf)
		if err != nil {
			log.Fatal(err)
		}

		// Return the URL of the added file.
		url := *gatewayURL + "/ipfs/" + hash + "/?filename=" + header.Filename
		response := map[string]interface{}{"url": url}
		r.JSON(200, response)
	})

	m.RunOnAddr(fmt.Sprintf("%s:%d", *hostname, *port))
}
