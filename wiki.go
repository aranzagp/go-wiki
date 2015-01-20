package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
)

var (
	addr = flag.Bool("addr", false, "find open address and print to final-port.txt")
)

type Page1 struct {
	Title string
	Body  []byte
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html")) //PArse and compile the regular expression
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (pg *Page1) save() error { //Metodo llamado save que recibe pg, el cual es un puntero a Page1.
	//No toma parametros y regresa un valor tipo error. Este metodo guardara el Body
	filename := pg.Title + ".txt"                    //de la pag en un archivo de texto, Title.txt
	return ioutil.WriteFile(filename, pg.Body, 0600) //octal integer literal, indica que el archivo se cea con permisos read-write
}

//El metodo save regresa un valor de error por que es el tipo de return de WriteFile(library that writes a
//byte slice to a file). Regresa un valor de error... Si todo sale bien Page.save() regresara nil. Valor cero para
// pointers, interfaces y otros tipos.

func loadPage1(title string) (*Page1, error) { //Construye el  nombre del archivo del parametro titulo
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename) //lee el contenido del file en una nueva variable body
	//body, _ := ioutil.ReadFile(filename) "_" ignora el valor de error que regresa ReadFile, no lo asigna
	if err != nil {
		return nil, err
	}
	return &Page1{Title: title, Body: body}, nil //regresa un puntero hacia la literal Page1
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) { //this will allow users to view a wiki page. It will handle
	// URLs  prefixed with "/view/"
	// Extrae el titulo de la pag de r.URL.Path, el componente del path de la URL pedida, the path is re-sliced with [len("/view/"):] to
	// drop the leading "/view/" component of the request path, esto es por que el path empezara con "/view/" el cual no es parte
	// del titulo de la pagina
	pg, err := loadPage1(title) // carga los datos de la pagina, formatea la pag con un string de HTML y se la escribe a w
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", pg)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	pg, err := loadPage1(title)
	if err != nil {
		pg = &Page1{Title: title}
	}
	renderTemplate(w, "edit", pg)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) { //handle the submission of forms located on the edit pages
	body := r.FormValue("body")
	pg := &Page1{Title: title, Body: []byte(body)}
	err := pg.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, tmpl string, pg *Page1) {
	err := templates.ExecuteTemplate(w, tmpl+".html", pg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//Extract the page title from the Request and call the provided handler 'fn'
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	flag.Parse()
	http.HandleFunc("/view/", makeHandler(viewHandler)) //handle any requests under the path /view/
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	if *addr {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile("final-port.txt", []byte(l.Addr().String()), 0644)
		if err != nil {
			log.Fatal(err)
		}
		s := &http.Server{}
		s.Serve(l)
		return
	}
	http.ListenAndServe(":8080", nil)
}

//go build wiki.go
//./wiki
//http://localhost:8080/view/ANewPage
