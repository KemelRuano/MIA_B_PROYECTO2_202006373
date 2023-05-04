package analizadores

import (
	"fmt"
	"regexp"
	"strings"
)

type Token struct {
	Token  string
	Lexema string
}

func Lexico(comando string) []Token {
	var index int = 0
	var estado int = 0
	var lexema string = ""
	PROCESOS := []string{"mkdisk", "rmdisk", "fdisk", "mount", "mkfs", "login", "logout", "mkgrp", "rmgrp", "mkusr", "rmusr", "rep", "mkdir", "pause"}
	ATRIBUTOS := []string{">size=", ">path=", ">fit=", ">unit=", ">name=", ">type=", ">id=", ">user=", ">pwd=", ">grp=", ">r", ">cont=", ">ruta="}
	COMANDO_LIST := []Token{}
	for index < len(comando) {
		var caracter byte = comando[index]
		if estado == 0 {
			if caracter >= 97 && caracter <= 122 || caracter >= 65 && caracter <= 90 {
				estado = 1
			} else if caracter >= 48 && caracter <= 57 {
				estado = 2
			} else if caracter == '"' {
				estado = 3
				index++
			} else if caracter == '>' {
				estado = 4
				index++
				lexema += string(caracter)
			} else if caracter == '/' {
				estado = 5
				index++
				lexema += string(caracter)
			} else if caracter == 32 {
				index++
			} else if caracter == 9 {
				index++
			} else if caracter == 10 {
				index++
			} else if caracter == 35 {
				estado = 6
			} else {
				index++
				estado = 0
				lexema = ""
				break
			}
		} else if estado == 1 {
			if caracter >= 97 && caracter <= 122 || caracter >= 48 && caracter <= 57 || caracter >= 65 && caracter <= 90 {
				index++
				lexema += string(caracter)
			} else {
				ya_valido := false
				for _, valor := range PROCESOS {
					if valor == strings.ToLower(lexema) {
						lexema = strings.ToLower(lexema)
						COMANDO_LIST = append(COMANDO_LIST, Token{Token: "PROCESO", Lexema: lexema})
						ya_valido = true
					}
				}

				if strings.ToLower(lexema) == "bf" || strings.ToLower(lexema) == "wf" || strings.ToLower(lexema) == "ff" {
					lexema = strings.ToLower(lexema)
					COMANDO_LIST = append(COMANDO_LIST, Token{Token: "TIPO_AJUSTE", Lexema: lexema})

				} else if strings.ToLower(lexema) == "m" || strings.ToLower(lexema) == "k" || strings.ToLower(lexema) == "b" {
					lexema = strings.ToLower(lexema)
					COMANDO_LIST = append(COMANDO_LIST, Token{Token: "UNIDAD", Lexema: lexema})

				} else if strings.ToLower(lexema) == "p" || strings.ToLower(lexema) == "e" || strings.ToLower(lexema) == "l" {
					lexema = strings.ToLower(lexema)
					COMANDO_LIST = append(COMANDO_LIST, Token{Token: "TIPO_PARTICION", Lexema: lexema})
				} else {
					if !ya_valido {
						lexema = strings.ToLower(lexema)
						COMANDO_LIST = append(COMANDO_LIST, Token{Token: "ID", Lexema: lexema})
					}

				}
				lexema = ""
				estado = 0
				ya_valido = false
			}

		} else if estado == 2 {

			if caracter >= 48 && caracter <= 57 || caracter >= 97 && caracter <= 122 {
				index++
				lexema += string(caracter)
			} else {
				re := regexp.MustCompile("[a-z]")
				contieneLetra := re.MatchString(lexema)
				if contieneLetra {
					COMANDO_LIST = append(COMANDO_LIST, Token{Token: "MOUNT", Lexema: lexema})
					lexema = ""
					estado = 0
				} else {
					COMANDO_LIST = append(COMANDO_LIST, Token{Token: "NUMERO", Lexema: lexema})
					lexema = ""
					estado = 0
				}

			}
		} else if estado == 3 {
			if caracter != '"' {
				index++
				lexema += string(caracter)
			}
			if caracter == '"' {
				index++
				COMANDO_LIST = append(COMANDO_LIST, Token{Token: "PATH", Lexema: lexema})
				lexema = ""
				estado = 0
			}
		} else if estado == 4 {
			if caracter >= 65 && caracter <= 90 || caracter >= 97 && caracter <= 122 || caracter == 61 {
				index++
				lexema += string(caracter)
			} else {

				if strings.Contains(strings.Join(ATRIBUTOS, ","), strings.ToLower(lexema)) {
					lexema = strings.ToLower(lexema)
					COMANDO_LIST = append(COMANDO_LIST, Token{Token: "ATRIBUTO", Lexema: lexema})
				}
				lexema = ""
				estado = 0

			}

		} else if estado == 5 {
			if caracter >= 46 && caracter <= 57 || caracter >= 97 && caracter <= 122 || caracter == 95 || caracter >= 65 && caracter <= 90 {
				index++
				lexema += string(caracter)
			} else {
				COMANDO_LIST = append(COMANDO_LIST, Token{Token: "PATHSINE", Lexema: lexema})
				lexema = ""
				estado = 0
			}
		} else if estado == 6 {
			if caracter != '\n' {
				index += 1
				lexema += string(caracter)
				fmt.Println(string(caracter))
			}
			if caracter == '\n' {
				fmt.Println("nuevo")
				index += 1
				estado = 0
				lexema = ""
			}

		}

	}

	return COMANDO_LIST

}
