package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"proyecto2/Backend/analizadores"
	"strconv"

	"github.com/gorilla/mux"
)

type Comand struct {
	Parametro string `json:"comando"`
}
type Login struct {
	Usuario  string `json:"usuario"`
	Password string `json:"password"`
	Id       string `json:"id"`
}

var Lista_Token analizadores.Token

func main() {
	router := mux.NewRouter()
	enableCORS(router)

	router.HandleFunc("/Comands", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var comando Comand
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&comando)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if analizadores.Estado {
			valor := analizadores.Delete(comando.Parametro)
			respuesta := Comand{
				Parametro: valor,
			}
			jsonBytes, _ := json.Marshal(respuesta)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonBytes)
		} else {
			terminal := comando.Parametro
			terminal += " "
			Lista_Token := analizadores.Lexico(terminal)
			analizadores.Sintactico(Lista_Token, terminal, w, r)

		}

	}).Methods("POST")

	router.HandleFunc("/Login", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var logeado Login
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&logeado)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		valor := analizadores.Acceder(logeado.Usuario, logeado.Password, logeado.Id)
		respuesta := Comand{
			Parametro: valor,
		}
		jsonBytes, _ := json.Marshal(respuesta)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)

	}).Methods("POST")

	router.HandleFunc("/disk", func(w http.ResponseWriter, r *http.Request) {
		pdfFile, err := os.Open(analizadores.LD())
		if err != nil {
			http.Error(w, "Archivo no encontrado", 404)
			return
		}
		defer pdfFile.Close()
		stat, err := pdfFile.Stat()
		if err != nil {
			http.Error(w, "Error al obtener información del archivo", 500)
			return
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=archivo.pdf")
		w.Header().Set("Content-Length", fileSize)

		io.Copy(w, pdfFile)

	}).Methods("POST")

	router.HandleFunc("/superbloque", func(w http.ResponseWriter, r *http.Request) {
		pdfFile, err := os.Open(analizadores.LSB())
		if err != nil {
			http.Error(w, "Archivo no encontrado", 404)
			return
		}
		defer pdfFile.Close()
		stat, err := pdfFile.Stat()
		if err != nil {
			http.Error(w, "Error al obtener información del archivo", 500)
			return
		}
		fileSize := strconv.FormatInt(stat.Size(), 10)

		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "inline; filename=archivo.pdf")
		w.Header().Set("Content-Length", fileSize)

		io.Copy(w, pdfFile)

	}).Methods("POST")

	http.ListenAndServe(":8080", router)
}

func enableCORS(router *mux.Router) {
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}).Methods(http.MethodOptions)
	router.Use(middlewareCors)
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization,Access-Control-Allow-Origin")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
			next.ServeHTTP(w, req)
		})
}
