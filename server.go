package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/codegangsta/martini-contrib/render"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/vincent-petithory/dataurl"
)

type Paste struct {
	Content string `form:"content" binding:"required"`
}

type Image struct {
	DataURL string `form:"dataurl" binding:"required"`
}

type UploadFile struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func main() {
	// Create content storage directories
	if err := os.MkdirAll(path.Join("files", "text"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(path.Join("files", "images"), 0755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(path.Join("files", "other"), 0755); err != nil {
		log.Fatal(err)
	}

	// Parse command-line flags.
	gatewayURL := flag.String("gateway", "https://ipfs.io", "The HTTP gateway used in shared file links.")
	port := flag.Int("port", 3000, "The port number to run the server on.")
	hostname := flag.String("hostname", "0.0.0.0", "The hostname to run under.")
	flag.Parse()

	m := martini.Classic()
	m.Use(render.Renderer())

	m.Get("/", func(r render.Render) {
		r.HTML(200, "home", struct {
			Gateway string
		}{
			*gatewayURL,
		})
	})

	m.Post("/paste", binding.Form(Paste{}), func(paste Paste, ferr binding.Errors, r render.Render) {
		// Write content to a file
		filename := time.Now().Format(time.UnixDate) + ".txt"
		filepath := path.Join("files", "text", filename)

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
		response := map[string]interface{}{"url": url}
		r.JSON(200, response)
	})

	m.Post("/image", binding.Form(Image{}), func(image Image, ferr binding.Errors, r render.Render) {
		// Parse Data URL
		dataURL, err := dataurl.DecodeString(image.DataURL)
		if err != nil {
			log.Fatal(err)
		}

		// Write content to a file
		filename := time.Now().Format(time.UnixDate)
		filepath := path.Join("files", "images", filename)

		f, err := os.Create(filepath)
		if err != nil {
			log.Fatal(err)
		}
		f.Write(dataURL.Data)
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
		log.Println("Adding file to IPFS...")
		cmd := exec.Command("ipfs", "add")
		cmd.Stdin = buf

		// Read and parse output from IPFS command.
		output, err := cmd.Output()
		log.Println(string(output))
		if err != nil {
			log.Fatal(err)
		}

		// Create a File URL and return
		log.Println("Returning result...")
		lines := strings.Split(string(output[:]), "\n")
		words := strings.Split(lines[len(lines)-2], " ")
		hash := words[1]
		url := *gatewayURL + "/ipfs/" + hash + "/?filename=" + header.Filename
		response := map[string]interface{}{"url": url}
		r.JSON(200, response)
	})

	m.RunOnAddr(fmt.Sprintf("%s:%d", *hostname, *port))
}
