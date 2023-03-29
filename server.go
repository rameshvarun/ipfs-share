package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

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

// Add the given buffer as a file to IPFS, returning it's hash.
func AddToIPFS(buf *bytes.Buffer) (string, error) {
	log.Println("Adding file to IPFS...")

	// Pipe buffer as STDIN to "ipfs add"
	cmd := exec.Command("ipfs", "add")
	cmd.Stdin = buf

	// Read and parse output from IPFS command.
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(output[:]), "\n")
	words := strings.Split(lines[len(lines)-2], " ")

	// Return file hash.
	hash := words[1]
	return hash, nil
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
		// Load paste content into a buffer.
		buf := new(bytes.Buffer)
		buf.WriteString(paste.Content)

		// Add buffer to IPFS
		hash, err := AddToIPFS(buf)
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

		// Try to determine appropriate file extension.
		exts, err := mime.ExtensionsByType(dataURL.ContentType())
		fmt.Println(exts)
		ext := ""
		if err == nil && exts != nil && len(exts) > 0 {
			ext = exts[0]
		}

		// Load paste content into a buffer.
		buf := new(bytes.Buffer)
		buf.Write(dataURL.Data)

		// Write content to a file
		hash, err := AddToIPFS(buf)

		// Create a share URL and return
		url := *gatewayURL + "/ipfs/" + hash
		if ext != "" {
			url += "?filename=paste" + ext
		}
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
		hash, err := AddToIPFS(buf)
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
