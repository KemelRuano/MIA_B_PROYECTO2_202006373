package process

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Admin_UG struct{}
type User struct {
	User     string
	Password string
	Id       string
	Uid      int
}

var RUTAIMAGEN string = ""
var RUTASB string = ""
var RUTATREE string = ""
var RUTAFILES string = ""
var Logeado User
var estado bool = false

var admindisk AdminDisk

type ExtraerPath struct {
	Namedisk string
	Listrep  []Report
}
type Report struct {
	Disk string
	Tree string
	Sb   string
	File string
}

func (m *ExtraerPath) AddId(disk string, tree string, sb string, file string) {
	m.Listrep = append(m.Listrep, Report{Disk: disk, Tree: tree, Sb: sb, File: file})
}

var List_reporte []ExtraerPath

func (a Admin_UG) Login(user string, password string, id string, admin AdminDisk) string {
	admindisk = admin
	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()

	if !estado {
		estado = true
	} else {
		return "~~~ ERROR [LOGIN] YA HAY UNA SESION INICIADA"
	}
	var paths string
	particion, err := admindisk.EncontrarParticion(id, &paths)
	if err != nil {
		estado = false
		return "~~~ ERROR [LOGIN] PARA LOGEARSE NECESITA UN DISCO MONTADO"
	}

	readfiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readfiles.Close()
	readfiles.Seek(int64(particion.PART_start), 0)
	binary.Read(readfiles, binary.LittleEndian, &Superblock)
	readfiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readfiles, binary.LittleEndian, &fileblock)

	var archivo string
	archivo += string(fileblock.B_content[:])
	list_users := a.extraer(archivo, 10)
	var encontrado bool = false
	var correct_user bool = false
	var correct_password bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'U' || list_users[i][2] == 'u' {
			Users := a.extraer(list_users[i], 44)
			for j := 0; j < len(Users); j++ {
				if Users[3] == user && Users[4] == password {
					encontrado = true
					Logeado.User = Users[3]
					Logeado.Password = Users[4]
					Logeado.Id = id
					uid, _ := strconv.Atoi(string(Users[0]))
					Logeado.Uid = uid
					break
				} else if Users[3] != user && Users[4] == password {
					correct_user = true
					break
				} else if Users[3] == user && Users[4] != password {
					correct_password = true
					break
				} else if Users[3] != user && Users[4] != password {
					correct_password = true
					correct_user = true
					break
				}

			}
		}
		if encontrado {
			break
		}

	}
	if !encontrado {
		estado = false
		if correct_user && !correct_password {
			return "~~~ ERROR [LOGIN] EL USUARIO NO EXISTE"
		} else if correct_password && !correct_user {
			return "~~~ ERROR [LOGIN] CONTRASEÑA INCORRECTA"
		} else if correct_user && correct_password {
			return "~~~ ERROR [LOGIN] EL USUARIO Y LA CONTRASEÑA SON INCORRECTOS"
		}
	}
	return "██████ [LOGIN] --- SESION INICIADA CON EXITO ██████"
}

func (a Admin_UG) Logout() string {
	if Logeado.User == "" {
		return "~~~ ERROR [LOGOUT] NO HAY UNA SESION INICIADA"
	}
	Logeado = User{}
	estado = false
	return "██████ [LOGOUT] --- SESION CERRADA CON EXITO ██████"

}

func (a Admin_UG) MKGRP(name string) string {
	if len(name) > 10 {
		return "~~~ ERROR [MKGRP] EL NOMBRE DEL GRUPO NO PUEDE TENER MAS DE 10 CARACTERES"
	}
	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(Logeado.User == "root" && Logeado.Password == "123") {
		return "~~~ ERROR [MKGRP] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := admindisk.EncontrarParticion(Logeado.Id, &paths)
	if err != nil {
		return "~~~ ERROR [MKGRP] PARA CREAR UN GRUPO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)

	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")
	list_users := a.extraer(archivo, 10)
	var cont_grp int = 1
	var newcont_grp int = 0
	var encontrado bool = false
	var ya_esta bool = false
	var newarchivo string = ""
	var newecontrado bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'G' || list_users[i][2] == 'g' {
			Users := a.extraer(list_users[i], 44)
			cont_grp++
			for j := 0; j < len(Users); j++ {
				if Users[0] != "0" && Users[2] == name {
					encontrado = true
					break
				} else if Users[0] == "0" && Users[2] == name {
					ya_esta = true
					newecontrado = true
					newcont_grp, _ = strconv.Atoi(string(list_users[i-1][0]))
					newcont_grp++
					newarchivo += strconv.Itoa(newcont_grp) + ",G," + name + "\n"
					break
				}
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if encontrado {
		fmt.Println("ARCHIVO: ", archivo)
		return "~~~ ERROR [MKGRP] EL GRUPO YA EXISTE"
	}
	if newecontrado {
		var bytes [64]byte
		copy(bytes[:], []byte(newarchivo))
		fileblock.B_content = bytes
		fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
		readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
		binary.Write(readFiles, binary.LittleEndian, &fileblock)
		return "██████ [MKGRP] --- GRUPO CREADO CON EXITO ██████"
	}

	archivo += strconv.Itoa(cont_grp) + ",G," + name + "\n"
	var bytes [64]byte
	copy(bytes[:], []byte(archivo))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)

	return "██████ [MKGRP] --- GRUPO CREADO CON EXITO ██████"
}

func (a Admin_UG) extraer(txt string, tab byte) []string {
	var enviar []string = strings.Split(txt, string(tab))
	for _, v := range enviar {
		if v == "" {
			enviar = enviar[:len(enviar)-1]
		}
	}
	return enviar
}

func (a Admin_UG) RMGRP(name string) string {

	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(Logeado.User == "root" && Logeado.Password == "123") {
		return "~~~ ERROR [RMGRP] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := admindisk.EncontrarParticion(Logeado.Id, &paths)
	if err != nil {
		return "~~~ ERROR [RMGRP] PARA ELIMINAR UN GRUPO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)

	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")
	fmt.Println("ARCHIVO: ", archivo)
	list_users := a.extraer(archivo, 10)
	var newarchivo string = ""
	var encontrado bool = false
	var ya_esta bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'G' || list_users[i][2] == 'g' {
			Users := a.extraer(list_users[i], 44)
			for j := 0; j < len(Users); j++ {
				if Users[0] != "0" && Users[2] == name {
					encontrado = true
					ya_esta = true
					newarchivo += strconv.Itoa(0) + ",G," + name + "\n"
					break
				} else if Users[0] == "0" && Users[2] == name {
					fmt.Println("ya esta eliminado el grupo")
					return "~~~ ERROR [RMGRP] EL GRUPO YA ESTA ELIMINADO"
				}
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if !encontrado {
		fmt.Println("El grupo no existe")
		return "~~~ ERROR [RMGRP] EL GRUPO NO EXISTE"
	}

	var bytes [64]byte
	copy(bytes[:], []byte(newarchivo))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)
	return "██████ [RMGRP] --- GRUPO ELIMINADO CON EXITO ██████"
}

func (a Admin_UG) MKUSR(user string, pwd string, grp string) string {
	if len(user) > 10 {
		return "~~~ ERROR [MKUSR] EL NOMBRE DE USUARIO NO PUEDE TENER MAS DE 10 CARACTERES"
	}
	if len(pwd) > 10 {
		return "~~~ ERROR [MKUSR] LA CONTRASEÑA NO PUEDE TENER MAS DE 10 CARACTERES"
	}
	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(Logeado.User == "root" && Logeado.Password == "123") {

		return "~~~ ERROR [MKUSR] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := admindisk.EncontrarParticion(Logeado.Id, &paths)
	if err != nil {
		return "~~~ ERROR [MKUSR] PARA CREAR UN USUARIO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)
	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")

	list_users := a.extraer(archivo, 10)
	var cont_user int = 0
	var ya_esta bool = false
	var validacion bool = false
	var newarchivo string = ""
	var newecontrado bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'G' {
			Users := a.extraer(list_users[i], 44)
			if Users[0] != "0" && Users[2] == grp {
				validacion = true
				cont_user, _ = strconv.Atoi(Users[0])
			} else if Users[0] == "0" && Users[2] == grp {
				return "~~~ ERROR [MKUSR] EL GRUPO ESTA ELIMINADO"
			}
		} else if list_users[i][2] == 'U' {
			Users := a.extraer(list_users[i], 44)
			if Users[0] != "0" && Users[3] == user {
				return "~~~ ERROR [MKUSR] EL USUARIO YA EXISTE"
			} else if Users[0] == "0" && Users[3] == user {
				ya_esta = true
				newecontrado = true
				newarchivo += strconv.Itoa(cont_user) + ",U," + grp + "," + user + "," + pwd + "\n"
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if !validacion {

		return "~~~ ERROR [MKUSR] EL GRUPO NO EXISTE"
	}

	var Contenido string = ""
	if newecontrado {
		Contenido = newarchivo
	} else {
		archivo += strconv.Itoa(cont_user) + ",U," + grp + "," + user + "," + pwd + "\n"
		Contenido = archivo
	}
	fmt.Println("byte es :", len(Contenido))
	if len(Contenido) > 64 {

		for i := 0; i < len(Contenido); i++ {

		}
	}
	var bytes [64]byte
	copy(bytes[:], []byte(Contenido))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)
	return "██████ [MKUSR] --- USUARIO CREADO CON EXITO ██████"
}

func (a Admin_UG) RMUSER(usuario string) string {
	Superblock := NewSuperblock()
	var fileblock Fileblock
	particion := NewPartition()
	if !(Logeado.User == "root" && Logeado.Password == "123") {

		return "~~~ ERROR [RMUSER] NO TIENE PERMISOS PARA EJECUTAR ESTE COMANDO"
	}
	var paths string
	particion, err := admindisk.EncontrarParticion(Logeado.Id, &paths)
	if err != nil {
		return "~~~ ERROR [RMUSER] PARA ELIMINAR UN USUARIO NECESITA UN DISCO MONTADO"
	}
	readFiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer readFiles.Close()
	readFiles.Seek(int64(particion.PART_start), 0)
	binary.Read(readFiles, binary.LittleEndian, &Superblock)
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(readFiles, binary.LittleEndian, &fileblock)

	archivo := strings.TrimRight(string(fileblock.B_content[:]), "\x00")

	list_users := a.extraer(archivo, 10)
	var newarchivo string = ""
	var encontrado bool = false
	var ya_esta bool = false
	for i := 0; i < len(list_users); i++ {
		if list_users[i][2] == 'U' {
			Users := a.extraer(list_users[i], 44)
			for j := 0; j < len(Users); j++ {
				if Users[0] != "0" && Users[3] == usuario {
					encontrado = true
					ya_esta = true
					newarchivo += strconv.Itoa(0) + ",U," + Users[2] + "," + usuario + "," + Users[4] + "\n"
					break
				} else if Users[0] == "0" && Users[3] == usuario {
					return "~~~ ERROR [RMUSR] EL USUARIO YA ESTA ELIMINADO"
				}
			}
		}
		if !ya_esta {
			newarchivo += list_users[i] + "\n"
		} else {
			ya_esta = false
			continue
		}
	}
	if !encontrado {
		return "~~~ ERROR [RMUSR] EL USUARIO NO EXISTE"
	}

	var bytes [64]byte
	copy(bytes[:], []byte(newarchivo))
	fileblock.B_content = bytes
	fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
	readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Write(readFiles, binary.LittleEndian, &fileblock)

	return "██████ [RMUSR] --- USUARIO ELIMINADO CON EXITO ██████"
}

func (a Admin_UG) REP(name string, paths string, id string, rute string) string {
	var particion Partition
	pathdisco := ""
	particion, err := admindisk.EncontrarParticion(id, &pathdisco)
	fmt.Println(particion.PART_fit, " ", particion.PART_name, " ", particion.PART_size, " ", particion.PART_start, " ", particion.PART_status, " ", particion.PART_type)
	if err != nil {
		return "~~~ ERROR [REP] PARA EL REPORTE NECESITA UN DISCO MONTADO"
	}

	if name == "disk" {
		DISK(paths, pathdisco)
	} else if name == "sb" {
		Superbloque(paths, pathdisco, particion)
	} else if name == "tree" {
		Tree(paths, pathdisco, particion)
	} else if name == "file" {
		FILE(paths, pathdisco, particion, rute)
	} else {
		return "~~~ ERROR [REP] NO EXISTE ESE TIPO DE REPORTE"
	}

	return "██████ [REP] --- REPORTE GENERADO CON EXITO ██████"
}
func DISK(ruta string, paths string) {
	var mbr Mbr
	imprimir, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer imprimir.Close()
	imprimir.Seek(0, 0)
	binary.Read(imprimir, binary.LittleEndian, &mbr)

	carpetas := strings.Replace(ruta, path.Base(ruta), "", -1)
	os.MkdirAll(carpetas, 0755)
	indicePunto := strings.LastIndex(ruta, ".")
	rutaSinExtension := ruta[:indicePunto]
	rutaSinExtension += ".dot"
	rutaImagen := ruta[:indicePunto]
	rutaImagen += ".pdf"
	archivo, _ := os.Create(rutaSinExtension)
	defer archivo.Close()
	fmt.Fprintf(archivo, "digraph { \n")
	fmt.Fprintf(archivo, "node [shape=plaintext];\n")
	fmt.Fprintf(archivo, "A [label=<<TABLE BORDER=\"6\" CELLBORDER=\"2\" CELLSPACING=\"1\" WIDTH=\"300\" HEIGHT=\"200\">\n")
	fmt.Fprintf(archivo, "<TR>\n")
	fmt.Fprintf(archivo, "<TD ROWSPAN=\"3\" WIDTH=\"300\" HEIGHT=\"200\"> MBR </TD>\n")
	contBlockLogic := 0
	es_acoupado := 0
	// es_acoupadoLogic := 0
	//tot := 0
	// totLogic := 0

	var List_part []Partition = admindisk.List_Partition(mbr)
	for i := 0; i < len(List_part); i++ {
		if List_part[i].PART_status == '1' {
			str := strings.TrimRight(string(List_part[i].PART_name[:]), "\x00")
			if List_part[i].PART_type == 'p' {
				var en_num float64 = float64(List_part[i].PART_size)
				var es_entero float64 = (en_num / float64(mbr.MBR_size)) * 100
				var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
				str1 := strconv.FormatFloat(es_porcentaje, 'f', 3, 64)
				var porcentaje string = str1 + "% del disco"
				fmt.Fprintf(archivo, "<TD ROWSPAN=\"3\" WIDTH=\"300\" HEIGHT=\"200\"> PARTICION PRIMARIA <BR/> %s <BR/> %s</TD>\n", str, porcentaje)
			} else if List_part[i].PART_type == 'e' {

				var numero float64 = float64(List_part[i].PART_size)
				var es_entero float64 = (numero / float64(mbr.MBR_size)) * 100
				var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
				str1 := strconv.FormatFloat(es_porcentaje, 'f', 3, 64)
				var porcentaje string = str1 + "% del disco"
				fmt.Fprintf(archivo, "<TD>\n")
				fmt.Fprintf(archivo, "    <TABLE BORDER=\"2\"  CELLBORDER=\"5\" CELLSPACING=\"3\"  WIDTH=\"300\" HEIGHT=\"200\">\n")
				var list_ext []EBR = admindisk.getlogics(List_part[i], paths)
				var total_br int = 0
				for j := 0; j < len(list_ext); j++ {
					contBlockLogic += 2
					total_br += int(list_ext[j].EBR_size)
				}
				var espacio_ebr = int(List_part[i].PART_size) - total_br
				if espacio_ebr > 0 {
					contBlockLogic += 2
				}
				fmt.Fprintf(archivo, "<TR>\n")
				fmt.Fprintf(archivo, "           <TD COLSPAN=\"%s\" WIDTH=\"300\" HEIGHT=\"200\"> PARTICION EXTENDIDA <BR/> %s <BR/> %s </TD> \n", strconv.Itoa(contBlockLogic), str, porcentaje)
				fmt.Fprintf(archivo, "</TR>\n")
				fmt.Fprintf(archivo, "<TR>\n")
				for j := 0; j < len(list_ext); j++ {
					if list_ext[j].EBR_status == '1' {
						fmt.Fprintf(archivo, "              <TD WIDTH=\"300\" HEIGHT=\"200\"> EBR </TD>\n")
						str1 = strings.TrimRight(string(list_ext[j].EBR_name[:]), "\x00")
						var numero float64 = float64(list_ext[j].EBR_size)
						var es_entero float64 = (numero / float64(List_part[i].PART_size)) * 100
						var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
						ss := strconv.FormatFloat(es_porcentaje, 'f', 2, 64)
						var porcentaje string = ss + "% de la particion extendida"
						fmt.Fprintf(archivo, "                <TD WIDTH=\"300\" HEIGHT=\"200\"> PARTICION LOGICA <BR/> %s <BR/> %s </TD>\n", str1, porcentaje)
					}
				}
				var numero2 float64 = float64(espacio_ebr)
				var es_entero2 float64 = (numero2 / float64(List_part[i].PART_size)) * 100
				var es_porcentaje2 float64 = math.Round(es_entero2*100.0) / 100.0
				ss := strconv.FormatFloat(es_porcentaje2, 'f', 4, 64)
				var porcentaje2 string = ss + "% de la particion extendida"
				fmt.Fprintf(archivo, "                   <TD WIDTH=\"300\" HEIGHT=\"200\">  LIBRE <BR/> %s </TD>\n", porcentaje2)

				fmt.Fprintf(archivo, "		</TR>\n")
				fmt.Fprintf(archivo, "	</TABLE>\n")
				fmt.Fprintf(archivo, "</TD>\n")
			}

			es_acoupado += int(List_part[i].PART_size)
		}
	}
	var es_vacio int = 0
	es_vacio = 136 + es_acoupado
	if es_vacio < int(mbr.MBR_size) {
		var numero float64 = float64(mbr.MBR_size) - float64(es_vacio)
		var es_entero float64 = (numero / float64(mbr.MBR_size)) * 100
		var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
		str1 := strconv.FormatFloat(es_porcentaje, 'f', 3, 64)
		var porcentaje string = str1 + "% del disco"
		fmt.Fprintf(archivo, "<TD ROWSPAN=\"3\" WIDTH=\"300\" HEIGHT=\"200\"> LIBRE <BR/> %s </TD>\n", porcentaje)

	}
	fmt.Fprintf(archivo, "</TR>\n")
	fmt.Fprintf(archivo, "</TABLE>>];\n")
	fmt.Fprintf(archivo, "label = \"Reporte DISK \n By: Kemel Ruano\" \n")
	fmt.Fprintf(archivo, "} \n")
	RUTAIMAGEN = rutaImagen
	exec.Command("dot", "-Tpdf", "-o", rutaImagen, rutaSinExtension).Run()

}
func Superbloque(ruta string, pathdisc string, encontrado Partition) {
	var sup Superblock
	imprimir, _ := os.OpenFile(pathdisc, os.O_RDWR, 0666)
	defer imprimir.Close()
	imprimir.Seek(int64(encontrado.PART_start), 0)
	binary.Read(imprimir, binary.LittleEndian, &sup)
	indicePunto := strings.LastIndex(ruta, ".")
	rutaDot := ruta[:indicePunto]
	rutaDot += ".dot"
	rutaImagen := ruta[:indicePunto]
	rutaImagen += ".pdf"
	carpetas := strings.Replace(ruta, path.Base(ruta), "", -1)
	os.MkdirAll(carpetas, 0755)

	archivo, _ := os.Create(rutaDot)
	defer archivo.Close()
	fmt.Fprintf(archivo, "digraph G {\n")
	fmt.Fprintf(archivo, "node [shape=plaintext]\n")
	fmt.Fprintf(archivo, "   graph [rankdir = LR bgcolor = white style=filled fontname = \"Courier New\"]; \n")
	fmt.Fprintf(archivo, "   Tabla[fontname = \"Courier New\" label=<<table border=\"2\" cellspacing=\"1\" cellborder = \"2\" width = \"300\" bgcolor = \"black\"> \n")
	fmt.Fprintf(archivo, "       <tr>  <td bgcolor=\"orange\" COLSPAN =\"2\" width = \"300\" height = \"50\"><b> SUPER_BLOQUE</b> </td>  </tr> \n")
	fmt.Fprintf(archivo, "       <tr>  <td bgcolor=\"skyblue\"> filesystem_type  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_filesystem_type)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Total inodos  </td><td bgcolor=\"white\"> %s  </td></tr> \n", strconv.Itoa(int(sup.S_inodes_count)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Total bloques  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_blocks_count)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Total bloques Libres  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_free_blocks_count)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Total inodos Libres </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_free_inodes_count)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Ultima Fecha Montado </td><td bgcolor=\"white\">  %s </td></tr> \n", time.Unix(sup.S_mtime, 0).Format("Jan 02, 2006 15:04:05"))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Ultima Fecha Desmontado </td><td bgcolor=\"white\"> %s  </td></tr> \n", time.Unix(sup.S_umtime, 0).Format("Jan 02, 2006 15:04:05"))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Desmontado  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_mnt_count)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Magic  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_magic)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Tamano Inodo  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_inode_size)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Tamano Bloque  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_block_size)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Primer Inodo Libre  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_first_ino)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Primer Bloque Libre  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_first_blo)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Inicio BM Inodo  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_bm_inode_start)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Inicio BM Bloque  </td><td bgcolor=\"white\">  %s </td></tr> \n", strconv.Itoa(int(sup.S_bm_block_start)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Inicio Tabla Inodo  </td><td bgcolor=\"white\"> %s </td></tr> \n", strconv.Itoa(int(sup.S_inode_start)))
	fmt.Fprintf(archivo, "       <tr><td bgcolor=\"skyblue\"> Inicio Tabla Bloques  </td><td bgcolor=\"white\"> % s </td></tr> \n", strconv.Itoa(int(sup.S_block_start)))
	fmt.Fprintf(archivo, "   </table>>]; \n")
	fmt.Fprintf(archivo, "   label = \"Reporte SB By: Kemel Ruano\"; \n")
	fmt.Fprintf(archivo, "}")
	RUTASB = rutaImagen
	exec.Command("dot", "-Tpdf", "-o", rutaImagen, rutaDot).Run()
}

func Tree(ruta string, paths string, partition Partition) {
	var sup Superblock
	var inode Inodes = NewInodes()
	imprimir, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer imprimir.Close()
	imprimir.Seek(int64(partition.PART_start), 0)
	binary.Read(imprimir, binary.LittleEndian, &sup)

	imprimir.Seek(int64(sup.S_bm_inode_start), 0)
	bmInodo := make([]byte, sup.S_inodes_count)
	binary.Read(imprimir, binary.LittleEndian, &bmInodo)

	imprimir.Seek(int64(sup.S_bm_block_start), 0)
	bmBloque := make([]byte, sup.S_blocks_count)
	binary.Read(imprimir, binary.LittleEndian, &bmBloque)

	imprimir.Seek(int64(sup.S_inode_start), 0)
	binary.Read(imprimir, binary.LittleEndian, &inode)
	var freeI int = Inodosiguiente(sup, paths)
	fmt.Println("freeI: ", freeI)
	indicePunto := strings.LastIndex(ruta, ".")
	rutaDot := ruta[:indicePunto]
	rutaDot += ".dot"
	rutaImagen := ruta[:indicePunto]
	rutaImagen += ".pdf"
	carpetas := strings.Replace(ruta, path.Base(ruta), "", -1)
	os.MkdirAll(carpetas, 0755)

	archivo, _ := os.Create(rutaDot)
	defer archivo.Close()
	fmt.Fprintf(archivo, "digraph G {\n")
	fmt.Fprintf(archivo, "node [shape = plaintext];\n")
	fmt.Fprintf(archivo, "   graph [rankdir = LR bgcolor = white style=filled]; \n")
	for i := 0; i < freeI; i++ {
		fmt.Fprintf(archivo, "inode%s [label=<<table border=\"2\" cellspacing=\"2\">\n", strconv.Itoa(i))
		fmt.Fprintf(archivo, "       <tr> \n ")
		fmt.Fprintf(archivo, " <td bgcolor=\"red\" COLSPAN =\"2\"><b> Inodo %s </b> </td> \n", strconv.Itoa(i))
		fmt.Fprintf(archivo, "</tr>\n")
		fmt.Fprintf(archivo, "<tr>\n")
		fmt.Fprint(archivo, "<td bgcolor=\"skyblue\"> UID:     </td>\n")
		fmt.Fprintf(archivo, "<td> %s </td>\n", strconv.Itoa(int(inode.I_uid)))
		fmt.Fprintf(archivo, "</tr>\n")
		fmt.Fprintf(archivo, "<tr>\n")
		fmt.Fprint(archivo, "<td bgcolor=\"skyblue\"> GID:     </td>\n")
		fmt.Fprintf(archivo, "<td> %s </td>\n", strconv.Itoa(int(inode.I_gid)))
		fmt.Fprintf(archivo, "</tr>\n")
		fmt.Fprintf(archivo, "<tr>\n")
		fmt.Fprint(archivo, "<td bgcolor=\"skyblue\"> TAMANO:     </td>\n")
		fmt.Fprintf(archivo, "<td> %s </td>\n", strconv.Itoa(int(inode.I_size)))
		fmt.Fprintf(archivo, "</tr>\n")
		fmt.Fprintf(archivo, "<tr>\n")
		fmt.Fprint(archivo, "<td bgcolor=\"skyblue\"> TIPO:     </td>\n")
		fmt.Fprintf(archivo, "<td> %s </td>\n", string(inode.I_type))
		fmt.Fprintf(archivo, "</tr>\n")
		fmt.Fprintf(archivo, "<tr>\n")
		fmt.Fprint(archivo, "<td bgcolor=\"skyblue\"> TIME_A:     </td>\n")
		fmt.Fprintf(archivo, "<td> %s </td>\n", time.Unix(inode.I_atime, 0).Format("Jan 02, 2006 15:04:05"))
		fmt.Fprintf(archivo, "</tr>\n")
		fmt.Fprintf(archivo, "<tr>\n")
		fmt.Fprint(archivo, "<td bgcolor=\"skyblue\"> TIME_C:     </td>\n")
		fmt.Fprintf(archivo, "<td> %s </td>\n", time.Unix(inode.I_ctime, 0).Format("Jan 02, 2006 15:04:05"))
		fmt.Fprintf(archivo, "</tr>\n")
		fmt.Fprintf(archivo, "<tr>\n")
		fmt.Fprint(archivo, "<td bgcolor=\"skyblue\"> TIME_M:     </td>\n")
		fmt.Fprintf(archivo, "<td> %s </td>\n", time.Unix(inode.I_mtime, 0).Format("Jan 02, 2006 15:04:05"))
		fmt.Fprintf(archivo, "</tr>\n")
		for j := 0; j < 15; j++ {
			fmt.Fprintf(archivo, "<tr>\n")
			fmt.Fprintf(archivo, "<td bgcolor=\"skyblue\"> i_block_%s </td> \n", strconv.Itoa(j))
			fmt.Fprintf(archivo, "<td > %s  </td> \n", strconv.Itoa(int(inode.I_block[j])))
			fmt.Fprintf(archivo, "</tr>\n")
		}

		fmt.Fprintf(archivo, "<tr><td bgcolor=\"skyblue\"> PERMISO:    </td><td> %s </td></tr> \n", strconv.Itoa(int(inode.I_perm)))
		fmt.Fprintf(archivo, "</table>>];\n")

		if inode.I_type == 48 {
			for j := 0; j < 15; j++ {
				if inode.I_block[j] != -1 {
					fmt.Fprintf(archivo, "inode%s -> BLOCK%s; \n", strconv.Itoa(i), strconv.Itoa(int(inode.I_block[j])))
					var foldertemp Folderblock
					imprimir.Seek(int64(sup.S_block_start+(int32(unsafe.Sizeof(Folderblock{}))*inode.I_block[j])), 0)
					binary.Read(imprimir, binary.LittleEndian, &foldertemp)
					fmt.Fprintf(archivo, "BLOCK%s [label=<<table border=\"2\" cellspacing=\"2\"> \n", strconv.Itoa(int(inode.I_block[j])))
					fmt.Fprintf(archivo, "<tr><td bgcolor=\"yellow\" COLSPAN =\"2\"><b> Bloque %s </b> </td></tr> \n", strconv.Itoa(int(inode.I_block[j])))
					for k := 0; k < 4; k++ {
						var ctmp string
						ctmp += strings.TrimRight(string(foldertemp.B_content[k].B_name[:]), "\x00")
						fmt.Fprintf(archivo, "<tr>\n")
						fmt.Fprintf(archivo, "<td bgcolor=\"skyblue\">  %s </td>\n", ctmp)
						fmt.Fprintf(archivo, "<td> %s </td>\n", strconv.Itoa(int(foldertemp.B_content[k].B_inodo)))
						fmt.Fprintf(archivo, "</tr>\n")
					}
					fmt.Fprintf(archivo, "</table>>];\n")

					for b := 0; b < 4; b++ {
						if foldertemp.B_content[b].B_inodo != -1 {
							es := strings.TrimRight(string(foldertemp.B_content[b].B_name[:]), "\x00")
							if !(es == "." || es == "..") {
								fmt.Fprintf(archivo, "BLOCK%s -> inode%s; \n", strconv.Itoa(int(inode.I_block[j])), strconv.Itoa(int(foldertemp.B_content[b].B_inodo)))
							}
						}
					}

				}
			}

		} else {
			for j := 0; j < 15; j++ {
				if inode.I_block[j] != -1 {
					if i < 12 {
						fmt.Fprintf(archivo, "inode%s -> BLOCK%s; \n", strconv.Itoa(i), strconv.Itoa(int(inode.I_block[j])))
						var filetemp Fileblock
						imprimir.Seek(int64(sup.S_block_start+(int32(unsafe.Sizeof(filetemp))*inode.I_block[j])), 0)
						binary.Read(imprimir, binary.LittleEndian, &filetemp)
						fmt.Fprintf(archivo, "BLOCK%s [label = <<table border=\"2\" cellspacing=\"2\"> \n", strconv.Itoa(int(inode.I_block[j])))
						fmt.Fprintf(archivo, "<tr><td bgcolor=\"yellow\" COLSPAN =\"2\"><b> Bloque %s </b> </td></tr>\n", strconv.Itoa(int(inode.I_block[j])))
						fmt.Fprintf(archivo, "<tr><td bgcolor=\"skyblue\"> %s </td></tr>\n", strings.TrimRight(string(filetemp.B_content[:]), "\x00"))
						fmt.Fprintf(archivo, "</table>>];\n")
					}
				}
			}
		}
		inode = NewInodes()
		imprimir.Seek(int64(sup.S_inode_start+(int32(unsafe.Sizeof(inode))*int32(i+1))), 0)
		binary.Read(imprimir, binary.LittleEndian, &inode)
	}

	fmt.Fprintf(archivo, "Inode2 [ label = <<table border=\"5\"> \n ")
	fmt.Fprintf(archivo, " <tr><td bgcolor=\"yellow\" COLSPAN=\"20\">BITMAP BLOQUE</td></tr>")
	var contes2 int = 0
	var esI2 bool = false
	for i := 0; i < int(sup.S_blocks_count); i++ {
		if contes2 == 0 {
			fmt.Fprintf(archivo, "<tr> \n")
		}
		fmt.Fprintf(archivo, " 	<td bgcolor = \"white\"> %s </td> \n", string(bmBloque[i]))
		if contes2 == 14 {
			fmt.Fprintf(archivo, "</tr> \n")
			contes2 = 0
			esI2 = true
		}
		if esI2 {
			esI2 = false
		} else {
			contes2++
		}
	}
	if contes2 != 0 {
		fmt.Fprintf(archivo, "</tr> \n")
	}
	fmt.Fprintf(archivo, "   </table>>]; \n")

	fmt.Fprintf(archivo, "Inode [ label = <<table border=\"5\"> \n ")
	fmt.Fprintf(archivo, " <tr><td colspan=\"5\" bgcolor=\"red\" COLSPAN=\"20\">BITMAP INODO</td></tr>")
	var contes int = 0
	var esI bool = false
	for i := 0; i < int(sup.S_inodes_count); i++ {
		if contes == 0 {
			fmt.Fprintf(archivo, "<tr> \n")
		}
		fmt.Fprintf(archivo, " 	<td bgcolor = \"white\"> %s </td> \n", string(bmInodo[i]))
		if contes == 14 {
			fmt.Fprintf(archivo, "</tr> \n")
			contes = 0
			esI = true
		}
		if esI {
			esI = false
		} else {
			contes++
		}
	}
	if contes != 0 {
		fmt.Fprintf(archivo, "</tr> \n")
	}
	fmt.Fprintf(archivo, "   </table>>]; \n")

	fmt.Fprintf(archivo, "size = \"10,8\"; \n")
	fmt.Fprintf(archivo, "ranksep=\"6\"; \n")
	fmt.Fprintf(archivo, "   label = \"Reporte TREE By: Kemel Ruano\"; \n")
	fmt.Fprintf(archivo, "}\n")

	RUTATREE = rutaImagen
	exec.Command("dot", "-Tpdf", "-o", rutaImagen, rutaDot).Run()

}

func Inodosiguiente(superbloque Superblock, paths string) int {

	Leer_modificar, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer Leer_modificar.Close()
	bitMapInodo := make([]byte, superbloque.S_inodes_count)
	Leer_modificar.Seek(int64(superbloque.S_bm_inode_start), 0)
	binary.Read(Leer_modificar, binary.LittleEndian, &bitMapInodo)
	for i := 0; i < int(superbloque.S_inodes_count); i++ {
		if bitMapInodo[i] == 48 {
			return i
		}
	}
	return -1
}

var nombreDisko string = ""

func (a Admin_UG) ViewsReporte(user string, pwd string, ids string) string {

	if user == "" || pwd == "" || ids == "" {
		return "FALTAN DATOS"
	}
	if Logeado.User != user && Logeado.Password == pwd && Logeado.Id == ids {
		return "NO EXISTE EL USUARIO"
	} else if Logeado.User == user && Logeado.Password != pwd && Logeado.Id == ids {
		return "CONTRASEÑA INCORRECTA"
	} else if Logeado.User == user && Logeado.Password == pwd && Logeado.Id != ids {
		return "ID INCORRECTO"
	}
	var paths string
	admindisk.EncontrarParticion(ids, &paths)
	fileName := path.Base(paths)
	nombreDisko = fileName
	estadoac := false
	for i := 0; i < len(List_reporte); i++ {
		if List_reporte[i].Namedisk == fileName {
			estadoac = true
			break
		} else {
			continue
		}
	}
	if !estadoac {
		List_reporte = append(List_reporte, ExtraerPath{})
		List_reporte[len(List_reporte)-1].Namedisk = fileName
		List_reporte[len(List_reporte)-1].AddId(RUTAIMAGEN, RUTATREE, RUTASB, RUTAFILES)
	}

	return "BIENVENIDO USUARIO"
}

func (a Admin_UG) DRUTE() string {
	for i := 0; i < len(List_reporte); i++ {
		if List_reporte[i].Namedisk == nombreDisko {
			for j := 0; j < len(List_reporte[i].Listrep); j++ {
				RUTAIMAGEN = List_reporte[i].Listrep[j].Disk
			}
		}
	}
	return RUTAIMAGEN
}

func (a Admin_UG) SBRUTE() string {
	for i := 0; i < len(List_reporte); i++ {
		if List_reporte[i].Namedisk == nombreDisko {
			for j := 0; j < len(List_reporte[i].Listrep); j++ {
				RUTASB = List_reporte[i].Listrep[j].Sb
			}
		}
	}
	return RUTASB
}
func (a Admin_UG) FILES2() string {
	for i := 0; i < len(List_reporte); i++ {
		if List_reporte[i].Namedisk == nombreDisko {
			for j := 0; j < len(List_reporte[i].Listrep); j++ {
				RUTAFILES = List_reporte[i].Listrep[j].File
			}
		}
	}
	return RUTAFILES
}

func FILE(paths string, pathdisco string, particion Partition, rute string) {
	var super Superblock
	var filetemp Fileblock
	archivo, _ := os.OpenFile(pathdisco, os.O_RDWR, 0666)
	defer archivo.Close()
	archivo.Seek(int64(particion.PART_start), 0)
	binary.Read(archivo, binary.LittleEndian, &super)

	archivo.Seek(int64(super.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
	binary.Read(archivo, binary.LittleEndian, &filetemp)

	carpetas := strings.Replace(paths, path.Base(paths), "", -1)
	os.MkdirAll(carpetas, 0755)

	archivotmp, _ := os.Create(paths)
	defer archivotmp.Close()
	Imprimir := strings.TrimRight(string(filetemp.B_content[:]), "\x00")
	archivotmp.WriteString(Imprimir)
	RUTAFILES = paths

}

func (a Admin_UG) TREERUTE() string {
	for i := 0; i < len(List_reporte); i++ {
		if List_reporte[i].Namedisk == nombreDisko {
			for j := 0; j < len(List_reporte[i].Listrep); j++ {
				RUTATREE = List_reporte[i].Listrep[j].Tree
			}
		}
	}
	return RUTATREE
}

func (d Admin_UG) Mkdir(paths string, Es_padre bool) {
	var sup Superblock
	var inode Inodes
	var particion Partition
	if Logeado.User == "" && Logeado.Password == "" {
		fmt.Println("NO HAY NINGUN USUARIO LOGEADO")
		return
	}
	path_particion := ""
	particion, _ = admindisk.EncontrarParticion(Logeado.Id, &path_particion)
	fmt.Println("PARTICION: ", particion.PART_name)
	fmt.Println("PARTICION: ", particion.PART_start)

	read, _ := os.OpenFile(path_particion, os.O_RDWR, 0666)
	defer read.Close()
	read.Seek(int64(particion.PART_start), 0)
	binary.Read(read, binary.LittleEndian, &sup)

	read.Seek(int64(sup.S_inode_start), 0)
	binary.Read(read, binary.LittleEndian, &inode)

}
