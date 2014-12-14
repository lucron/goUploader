package main

import (
	"crypto/rand"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var StdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

func main() {
	log.Println("Start Uploader.")

	//TODO startup routine to create folders and stuff
	// html, html/img, html/index.html, html/img/index.html

	cwd, err := os.Getwd()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Println("Root: " + cwd)
	fs := http.FileServer(http.Dir("html"))
	http.Handle("/", fs)
	http.HandleFunc("/upload", uploadHandler)
	log.Fatal(http.ListenAndServe("localhost:4242", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Println(err.Error())
		return
	}
	// check if filetype is allowed
	err = CheckMIME(file)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//hackish workaround, CheckMIME already reads some byte,
	//so the file is not complete when written.
	file.Close()
	file, header, _ = r.FormFile("file")
	defer file.Close()
	//TODO: dont parse fileextension without verifying
	filename, err := SaveFile(file, filepath.Ext(header.Filename))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, "/img/"+filename, http.StatusFound)
	//fmt.Fprintf(w, "<a href='/img/"+filename+"'>"+filename+"</a>")
}

func CheckMIME(file io.Reader) error {
	b := make([]byte, 512)
	if _, err := file.Read(b); err != nil {
		return err
	}
	mime := http.DetectContentType(b)
	if !strings.HasPrefix(mime, "video") && !strings.HasPrefix(mime, "image") {
		return (errors.New("filetype " + mime + " not allowed"))
	}
	return nil
}

func SaveFile(src io.Reader, ext string) (string, error) {
	fn := newLenChars(10, StdChars)
	dest, err := os.Create("html/img/" + fn + ext)
	defer dest.Close()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(dest, src)
	if err != nil {
		return "", nil
	}
	return fn + ext, nil
}

func newLenChars(length int, chars []byte) string {
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, r); err != nil {
			panic("error reading from random source: " + err.Error())
		}
		for _, c := range r {
			if c >= maxrb {
				continue
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
	panic("unreachable")
}
