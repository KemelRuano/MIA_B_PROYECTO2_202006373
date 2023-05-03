package analizadores

import (
	"MIA_B_PROYECTO2_202006373/Backend/process"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type Estatus struct {
	Parametro string `json:"comando"`
}

var admindisk process.AdminDisk
var adminug process.Admin_UG
var Pathtemp string = ""
var Estado bool = false

func Sintactico(List_sintax []Token, comando string, w http.ResponseWriter, r *http.Request, Tipo_arch string, Envio *string) {
	var path string
	var size string
	var fit string
	var unit string
	var types string
	var name string
	var id string
	var user string
	var password string
	var grupo string
	var count string
	var rrr bool
	var ruta string

	for i := 0; i < len(List_sintax); i++ {

		if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">path=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "PATH" || List_sintax[i+1].Token == "PATHSINE") {
			path = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">size=" && i+1 < len(List_sintax) && List_sintax[i+1].Token == "NUMERO" {
			size = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">fit=" && i+1 < len(List_sintax) && List_sintax[i+1].Token == "TIPO_AJUSTE" {
			fit = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">unit=" && i+1 < len(List_sintax) && List_sintax[i+1].Token == "UNIDAD" {
			unit = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">name=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "ID" || List_sintax[i+1].Token == "PATH") {
			name = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">type=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "TIPO_PARTICION" || List_sintax[i+1].Token == "ID" || List_sintax[i+1].Token == "PATH") {
			types = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">id=" && i+1 < len(List_sintax) && List_sintax[i+1].Token == "MOUNT" {
			id = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">user=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "ID" || List_sintax[i+1].Token == "PATH" || List_sintax[i+1].Token == "NUMERO") {
			user = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">pwd=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "ID" || List_sintax[i+1].Token == "PATH" || List_sintax[i+1].Token == "NUMERO") {
			password = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">grp=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "ID" || List_sintax[i+1].Token == "PATH") {
			grupo = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">cont=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "PATH" || List_sintax[i+1].Token == "PATHSINE") {
			count = List_sintax[i+1].Lexema
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">r" {
			rrr = true
		} else if List_sintax[i].Token == "ATRIBUTO" && List_sintax[i].Lexema == ">ruta=" && i+1 < len(List_sintax) && (List_sintax[i+1].Token == "ID" || List_sintax[i+1].Token == "PATH") {
			ruta = List_sintax[i+1].Lexema
		}
	}

	if strings.Contains(comando, "mkdisk") {
		condicion1 := false
		condicion2 := false
		if ValidarAtributos("mkdisk", ">path=", path, true, comando) {
			condicion1 = true
		}
		if ValidarAtributos("mkdisk", ">size=", size, true, comando) {
			condicion2 = true
		}
		if !ValidarAtributos("mkdisk", ">fit=", fit, false, comando) {
			return
		}
		if !ValidarAtributos("mkdisk", ">unit=", unit, false, comando) {
			return
		}

		if condicion1 && condicion2 {

			valor := admindisk.Mkdisk(path, size, fit, unit)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}

		}

	} else if strings.Contains(comando, "rmdisk") {
		condicion1 := false
		if ValidarAtributos("rmdisk", ">path=", path, true, comando) {
			condicion1 = true
		}

		if condicion1 {
			Pathtemp = path
			Estado = true
			valor := admindisk.Rmdisk(path)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "fdisk") {
		condicion1 := false
		condicion2 := false
		condicion3 := false
		if ValidarAtributos("fdisk", ">path=", path, true, comando) {
			condicion1 = true
		}
		if ValidarAtributos("fdisk", ">size=", size, true, comando) {
			condicion2 = true
		}
		if ValidarAtributos("fdisk", ">name=", name, true, comando) {
			condicion3 = true
		}
		if !ValidarAtributos("mkdisk", ">fit=", fit, false, comando) {
			return
		}
		if !ValidarAtributos("mkdisk", ">unit=", unit, false, comando) {
			return
		}

		if !ValidarAtributos("mkdisk", ">type=", types, false, comando) {
			return
		}

		if condicion1 && condicion2 && condicion3 {
			valor := admindisk.Fdisk(size, path, name, types, unit, fit)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "mount") {

		condicion1 := false
		condicion2 := false
		if ValidarAtributos("mount", ">path=", path, true, comando) {
			condicion1 = true
		}
		if ValidarAtributos("mount", ">name=", name, true, comando) {
			condicion2 = true
		}

		if condicion1 && condicion2 {
			valor := admindisk.MOUNT(path, name)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "mkfs") {
		condicion1 := false
		if ValidarAtributos("mkfs", ">id=", id, true, comando) {
			condicion1 = true
		}
		if !ValidarAtributos("mkfs", ">type=", types, false, comando) {
			return
		}

		if condicion1 {
			valor := admindisk.MKFS(types, id)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "login") {
		condicion1 := false
		condicion2 := false
		condicion3 := false
		if ValidarAtributos("login", ">user=", user, true, comando) {
			condicion1 = true
		}
		if ValidarAtributos("login", ">pwd=", password, true, comando) {
			condicion2 = true
		}
		if ValidarAtributos("login", ">id=", id, true, comando) {
			condicion3 = true
		}

		if condicion1 && condicion2 && condicion3 {
			valor := adminug.Login(user, password, id, admindisk)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "logout") {
		valor := adminug.Logout()
		if Tipo_arch == "comando" {
			respuesta := Estatus{
				Parametro: valor,
			}
			jsonBytes, _ := json.Marshal(respuesta)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonBytes)
		} else {
			*Envio += valor + `;`
		}
	} else if strings.Contains(comando, "mkgrp") {
		condicion1 := false

		if ValidarAtributos("mkgrp", ">name=", name, true, comando) {
			condicion1 = true
		}

		if condicion1 {
			valor := adminug.MKGRP(name)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "rmgrp") {
		condicion1 := false

		if ValidarAtributos("rmgrp", ">name=", name, true, comando) {
			condicion1 = true
		}

		if condicion1 {
			valor := adminug.RMGRP(name)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "mkusr") {
		condicion1 := false
		condicion2 := false
		condicion3 := false
		if ValidarAtributos("mkusr", ">user=", user, true, comando) {
			condicion1 = true
		}
		if ValidarAtributos("mkusr", ">pwd=", password, true, comando) {
			condicion2 = true
		}
		if ValidarAtributos("mkusr", ">grp=", grupo, true, comando) {
			condicion3 = true
		}

		if condicion1 && condicion2 && condicion3 {
			valor := adminug.MKUSR(user, password, grupo)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "rmusr") {
		condicion1 := false
		if ValidarAtributos("mkusr", ">user=", user, true, comando) {
			condicion1 = true
		}
		if condicion1 {
			valor := adminug.RMUSER(user)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	} else if strings.Contains(comando, "mkfile") {
		condicion1 := false
		if ValidarAtributos("mkfile", ">path=", path, true, comando) {
			condicion1 = true
		}

		if !ValidarAtributos("mkfile", ">size=", size, false, comando) {
			return
		}
		if !ValidarAtributos("mkfile", ">cont=", count, false, comando) {
			return
		}

		if condicion1 {
			fmt.Println("mkfile", rrr)
		}
	} else if strings.Contains(comando, "mkdir") {
		condicion1 := false
		if ValidarAtributos("mkdir", ">path=", path, true, comando) {
			condicion1 = true
		}

		if condicion1 {
			fmt.Println("mkdir", rrr)
			adminug.Mkdir(path, rrr)
		}

	} else if strings.Contains(comando, "pause") {
		respuesta := Estatus{
			Parametro: "pause",
		}
		jsonBytes, _ := json.Marshal(respuesta)
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)

	} else if strings.Contains(comando, "rep") {
		condicion1 := false
		condicion2 := false
		condicion3 := false
		if ValidarAtributos("rep", ">path=", path, true, comando) {
			condicion1 = true
		}
		if ValidarAtributos("rep", ">name=", name, true, comando) {
			condicion2 = true
		}

		if ValidarAtributos("rep", ">id=", id, true, comando) {
			condicion3 = true
		}

		if !ValidarAtributos("rep", ">ruta=", ruta, false, comando) {
			return
		}

		if condicion1 && condicion2 && condicion3 {
			valor := adminug.REP(name, path, id, ruta)
			if Tipo_arch == "comando" {
				respuesta := Estatus{
					Parametro: valor,
				}
				jsonBytes, _ := json.Marshal(respuesta)
				w.Header().Set("Content-Type", "application/json")
				w.Write(jsonBytes)
			} else {
				*Envio += valor + `;`
			}
		}

	}

	rrr = false

}
func ValidarAtributos(tipo string, atrib string, valor string, obli bool, comando string) bool {

	if obli {
		if strings.Contains(comando, atrib) {
			if valor == "" {
				fmt.Printf("%s --- falta valor a %s \n", tipo, atrib)
				return false
			}

		} else {
			fmt.Printf("%s --- falta parametro obligatorio %s \n", tipo, atrib)
			return false
		}
	} else {
		if strings.Contains(comando, atrib) {
			if valor == "" {
				fmt.Printf("%s --- falta valor a %s \n", tipo, atrib)
				return false
			}
		}

	}

	return true
}

func Delete(respuesta string) string {
	if respuesta == "Y" || respuesta == "y" {
		os.Remove(Pathtemp)
		Pathtemp = ""
		Estado = false
		return "█████████ [RMDISK] ---- DISCO ELIMINADO CON EXITO █████████"
	} else if respuesta == "N" || respuesta == "n" {
		return "█████████ [RMDISK] ---- CANCELADO █████████"
	}
	return "~~~ ERROR [RMDISK] ---- RESPUESTA INVALIDA"
}

func Acceder(user string, pass string, ids string) string {

	return adminug.ViewsReporte(user, pass, ids)
}

func LD() string {
	return adminug.DRUTE()
}
func LSB() string {
	return adminug.SBRUTE()
}

func FILES() string {
	return adminug.FILES2()
}

func LTREE() string {
	return adminug.TREERUTE()
}
