package dots

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"

	"github.com/nfnt/resize"

	"appengine"
	"appengine/datastore"
)

type Image struct {
	Data []byte
}

const maxImageSide = 1200

var templates = template.Must(template.ParseFiles("index.html", "edit.html", "error.html"))

func init() {
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/edit", serveEdit)
	http.HandleFunc("/img", serveImage)
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	//show the upload form
	if r.Method != "POST" {
		if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			serveError(w, r, err)
			return
		}
		return
	}

	//process the image upload
	f, _, err := r.FormFile("image")
	if err != nil {
		serveError(w, r, err)
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	io.Copy(&buf, f)
	i, _, err := image.Decode(&buf)
	if err != nil {
		serveError(w, r, err)
		return
	}

	//resize the image if too large
	if b := i.Bounds(); b.Dx() > maxImageSide || b.Dy() > maxImageSide {
		w, h := maxImageSide, maxImageSide
		if b.Dx() > b.Dy() {
			h = b.Dy() * h / b.Dx()
		} else {
			w = b.Dx() * w / b.Dy()
		}

		i = resize.Resize(uint(w), uint(h), i, resize.Lanczos3)
	}

	buf.Reset()
	if err := png.Encode(&buf, i); err != nil {
		serveError(w, r, err)
		return
	}

	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "Image", generateKey(buf.Bytes()), 0, nil)
	if _, err := datastore.Put(ctx, key, &Image{buf.Bytes()}); err != nil {
		serveError(w, r, err)
		return
	}

	//redirect to the edit page
	http.Redirect(w, r, "/edit?id="+key.StringID(), http.StatusFound)
}

func serveEdit(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "edit.html", r.FormValue("id")); err != nil {
		serveError(w, r, err)
		return
	}
}

func serveImage(w http.ResponseWriter, r *http.Request) {
	//serve up the requested image
	ctx := appengine.NewContext(r)
	key := datastore.NewKey(ctx, "Image", r.FormValue("id"), 0, nil)
	i := &Image{}
	if err := datastore.Get(ctx, key, i); err != nil {
		serveError(w, r, err)
		return
	}

	img, _, err := image.Decode(bytes.NewBuffer(i.Data))
	if err != nil {
		serveError(w, r, err)
		return
	}

	dots := getInt(r.FormValue("dots"), 20)
	img, err = drawDots(img, dots)
	if err != nil {
		serveError(w, r, err)
		return
	}

	b := &bytes.Buffer{}
	if err := png.Encode(b, img); err != nil {
		serveError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	b.WriteTo(w)
}

func serveError(w http.ResponseWriter, r *http.Request, err error) {
	ctx := appengine.NewContext(r)
	ctx.Errorf("error: %v", err)
	w.WriteHeader(http.StatusInternalServerError)
	if err := templates.ExecuteTemplate(w, "error.html", err); err != nil {
		ctx.Errorf("serveError: %v", err)
	}
}

func generateKey(data []byte) string {
	return fmt.Sprintf("%x", md5.Sum(data))
}

func getInt(data string, defaultInt int) int {
	i, err := strconv.Atoi(data)
	if err != nil {
		i = defaultInt
	}
	return i
}
