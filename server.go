package main

import (
	"os"
	"flag"
	"log"
	"time"
	"path"
	"os/exec"
	"strings"
	"fmt"
	"mime/multipart"
	"io"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/codegangsta/martini-contrib/render"
	"github.com/boltdb/bolt"
	"github.com/martini-contrib/binding"
)

type Paste struct {
    Content            string  `form:"content" binding:"required"`
}

type UploadFile struct {
    File *multipart.FileHeader `form:"file" binding:"required"`
}

func main() {
	// Create content storage directories
	if err := os.MkdirAll(path.Join("files","text"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(path.Join("files","images"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(path.Join("files","other"), 0755); err != nil {
		log.Fatal(err)
	}

	gatewayURL := flag.String("gateway", "http://gateway.ipfs.io", "The HTTP gateway used in shared file links.")
	port := flag.Int("port", 3000, "The port number to run the server on.")
	hostname := flag.String("hostname", "0.0.0.0", "The hostname to run under.")
	flag.Parse()

	db, err := bolt.Open("blog.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	m := martini.Classic()
	m.Use(render.Renderer())
	m.Post("/paste", binding.Form(Paste{}), func(paste Paste, ferr binding.Errors, r render.Render) {
		// Write content to a file
		filename := time.Now().Format(time.UnixDate) + ".txt"
		filepath := path.Join("files","text", filename)

		f, err := os.Create(filepath)
		if err != nil {
			log.Fatal(err)
		}
		f.WriteString(paste.Content)
		f.Close()

		// Add this file to IPFS
		cmd := exec.Command("ipfs", "add", filepath)
		output, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		// Create a share URL and return
		words := strings.Split(string(output[:]), " ")
		hash := words[1]
		url := *gatewayURL + "/ipfs/" + hash
		response := map[string]interface{}{ "url": url }
		r.JSON(200, response)
	})

	m.Post("/upload", func(req *http.Request, r render.Render) {
		infile, header, err := req.FormFile("file")
		if err != nil {
			log.Fatal(err)
		}
		defer infile.Close()

		outfilename := header.Filename
		outfilepath := path.Join("files","other", outfilename)

		outfile, err := os.Create(outfilepath)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Piping upload into disk file.")
		io.Copy(outfile, infile)
		outfile.Close()

		// Add to IPFS
		cmd := exec.Command("ipfs", "add", outfilepath)
		output, err := cmd.Output()
		if err != nil {
			log.Fatal(err)
		}

		// Create a File URL and return
		words := strings.Split(string(output[:]), " ")
		hash := words[1]
		url := *gatewayURL + "/ipfs/" + hash
		response := map[string]interface{}{ "url": url }
		r.JSON(200, response)
	})

	m.RunOnAddr(fmt.Sprintf("%s:%d",*hostname, *port))
}