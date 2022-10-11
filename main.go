package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"go-sample-web/model"
	"go-sample-web/storage"
	"go-sample-web/storage/sqlite"
)

const (
	WRITE  = "templates/write.html"
	HEADER = "templates/header.html"
	FOOTER = "templates/footer.html"
	INDEX  = "templates/index.html"
)

type eventHandler func(_ storage.Storage, _ http.ResponseWriter, _ *http.Request) error

func main() {
	var port int
	var host, storageDb string
	flag.IntVar(&port, "port", 8080, "Port to start app")
	flag.StringVar(&host, "host", "localhost", "Host to start app")
	flag.StringVar(&storageDb, "storage-db", "storage.db", "Host to start app")
	flag.Parse()
	if !flag.Parsed() {
		panic("cannot initialize parameters")
	}

	store, err := sqlite.New(storageDb)
	if err != nil {
		log.Fatalf("cannot initialize database: %s", storageDb)
	}
	err = store.Init(getContext())
	if err != nil {
		log.Fatal(err)
	}
	var resource = fmt.Sprintf("%s:%d", host, port)

	log.Printf("Listening on host: %s, port :%d", host, port)
	log.Printf("Using storage in : %s", storageDb)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))

	// closure to pass storage to all event handlers
	hand := func(pattern string, h eventHandler) {
		addHandler(pattern, store, h)
	}
	hand("/", indexHandler)
	hand("/write", writeHandler)
	hand("/edit", editHandler)
	hand("/delete", deleteHandler)
	hand("/save", saveHandler)

	if err := http.ListenAndServe(resource, nil); err != nil {
		panic(err)
	}
}

func addHandler(pattern string, store storage.Storage, f eventHandler) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Processing %s", r.URL.Path)
		if err := f(store, w, r); err != nil {
			log.Printf("error during processing: %v", err.Error())
		}
	}
	http.HandleFunc(pattern, handler)
}

// Handler for index page
func indexHandler(storage storage.Storage, w http.ResponseWriter, _ *http.Request) error {
	if t, err := template.ParseFiles(INDEX, HEADER, FOOTER); err != nil {
		log.Fatalf("cannot parse templates %v", err.Error())
		return nil
	} else {
		var posts []*model.Post
		if posts, err = storage.GetAll(getContext()); err != nil {
			return err
		}
		if err = t.ExecuteTemplate(w, "index", posts); err != nil {
			return fmt.Errorf("error while redirecting to index %v", err.Error())
		}
		return nil
	}
}

func getContext() context.Context {
	return context.Background()
}

// writeHandler - handler for write page
func writeHandler(_ storage.Storage, w http.ResponseWriter, _ *http.Request) error {
	if t, err := template.ParseFiles(WRITE, HEADER, FOOTER); err != nil {
		log.Fatalf("cannot parse templates %v", err.Error())
		return nil
	} else {
		if err = t.ExecuteTemplate(w, "write", nil); err != nil {
			return err
		}
		return nil
	}
}

// editHandler - handler for edit page
func editHandler(store storage.Storage, w http.ResponseWriter, r *http.Request) error {
	if t, err := template.ParseFiles(WRITE, HEADER, FOOTER); err != nil {
		log.Fatalf("cannot parse templates %v", err.Error())
		return nil
	} else {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			return err
		}
		ctx := getContext()
		found, err := store.Exists(ctx, id)
		if !found {
			http.NotFound(w, r)
			return errors.New("id not found in request")
		}
		post, err := store.Find(ctx, id)
		if err != nil {
			return err
		}
		if err = t.ExecuteTemplate(w, "write", post); err != nil {
			return err
		}
		return nil
	}
}

// saveHandler - handler for save
func saveHandler(store storage.Storage, w http.ResponseWriter, r *http.Request) error {
	id := r.FormValue("id")
	title := r.FormValue("title")
	text := r.FormValue("text")
	log.Printf("Saving %s %s %s", id, title, text)
	ctx := getContext()
	if id != "" {
		var id2, err = strconv.Atoi(id)
		if err != nil {
			return err
		}
		post, err := store.Find(ctx, id2)
		post.Title = title
		post.Text = text
		post.Time = time.Now()
		if err = store.Save(ctx, post); err != nil {
			return err
		}
	} else {
		post := model.LocalPost(title, text)
		if err := store.Add(ctx, post); err != nil {
			return err
		}
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}

// deleteHandler - handler for delete
func deleteHandler(store storage.Storage, w http.ResponseWriter, r *http.Request) error {
	id := r.FormValue("id")
	if id == "" {
		http.NotFound(w, r)
		return errors.New("id not found in request")
	}
	id2, err := strconv.Atoi(id)
	if err != nil {
		return err
	}
	if err = store.Delete(getContext(), id2); err != nil {
		return err
	}
	http.Redirect(w, r, "/", http.StatusFound)
	return nil
}
