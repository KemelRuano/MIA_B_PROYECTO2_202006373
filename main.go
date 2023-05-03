package main

import (
	"MIA_B_PROYECTO2_202006373/Backend/analizadores"
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
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
			datos := ""
			terminal := comando.Parametro
			terminal += " "
			Lista_Token := analizadores.Lexico(terminal)
			analizadores.Sintactico(Lista_Token, terminal, w, r, "comando", &datos)

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

	router.HandleFunc("/tree", func(w http.ResponseWriter, r *http.Request) {
		pdfFile, err := os.Open(analizadores.LTREE())
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

	router.HandleFunc("/repfile", func(w http.ResponseWriter, r *http.Request) {
		archivo, err := os.Open(analizadores.FILES())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer archivo.Close()
		scanner := bufio.NewScanner(archivo)
		var contenido string
		for scanner.Scan() {
			contenido += scanner.Text() + "\n"
		}
		response := map[string]string{
			"contenido": contenido,
		}
		json.NewEncoder(w).Encode(response)
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

	router.HandleFunc("/File", func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("mi_archivo")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()
		verificar := header.Filename
		if verificar[len(verificar)-3:] != "eea" {
			respuesta := Comand{
				Parametro: "!!!!! ERROR EL ARCHIVO DEBE SER .eea !!!!!",
			}
			jsonBytes, _ := json.Marshal(respuesta)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonBytes)
			return

		}
		var Texto string
		scanner := bufio.NewScanner(file)
		Extraido := ""
		for scanner.Scan() {
			line := scanner.Text()
			line += " "
			Texto += line + "\n"
			Lista_Token := analizadores.Lexico(line)
			analizadores.Sintactico(Lista_Token, line, w, r, "file", &Extraido)
		}

		respuesta := Comand{
			Parametro: Extraido,
		}
		jsonBytes, _ := json.Marshal(respuesta)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
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
