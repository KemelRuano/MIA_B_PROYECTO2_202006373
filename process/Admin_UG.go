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
var Logeado User
var estado bool = false

var admindisk AdminDisk

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
	if newecontrado {
		var bytes [64]byte
		copy(bytes[:], []byte(newarchivo))
		fileblock.B_content = bytes
		fmt.Println("NEW ARCHIVO: ", string(fileblock.B_content[:]))
		readFiles.Seek(int64(Superblock.S_block_start)+int64(unsafe.Sizeof(Folderblock{})), 0)
		binary.Write(readFiles, binary.LittleEndian, &fileblock)
		return "██████ [MKUSR] --- USUARIO CREADO CON EXITO ██████"
	}

	archivo += strconv.Itoa(cont_user) + ",U," + grp + "," + user + "," + pwd + "\n"
	var bytes [64]byte
	copy(bytes[:], []byte(archivo))
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

	pathdisco := ""
	_, err := admindisk.EncontrarParticion(id, &pathdisco)
	if err != nil {
		return "~~~ ERROR [REP] PARA EL REPORTE NECESITA UN DISCO MONTADO"
	}

	if name == "disk" {
		DISK(paths, pathdisco)
	} else if name == "sb" {
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

	carpetas := strings.Replace(paths, path.Base(ruta), "", -1)
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
	es_acoupadoLogic := 0
	tot := 0
	totLogic := 0

	List_part := admindisk.List_Partition(mbr)
	for i := 0; i < len(List_part); i++ {
		if List_part[i].PART_status == '1' {
			str := strings.TrimRight(string(List_part[i].PART_name[:]), "\x00")
			if List_part[i].PART_type == 'p' {
				var en_num float64 = float64(List_part[i].PART_size)
				var es_entero float64 = (en_num / float64(mbr.MBR_size)) * 100
				var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
				str1 := strconv.FormatFloat(es_porcentaje, 'f', 4, 64)
				var porcentaje string = str1 + "% del disco"
				fmt.Fprintf(archivo, "<TD ROWSPAN=\"3\" WIDTH=\"300\" HEIGHT=\"200\"> PARTICION PRIMARIA <BR/> %s <BR/> %s</TD>\n", str, porcentaje)
			} else if List_part[i].PART_type == 'e' {

				var numero float64 = float64(List_part[i].PART_size)
				var es_entero float64 = (numero / float64(mbr.MBR_size)) * 100
				var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
				str1 := strconv.FormatFloat(es_porcentaje, 'f', 4, 64)
				var porcentaje string = str1 + "% del disco"
				fmt.Fprintf(archivo, "<TD>\n")
				fmt.Fprintf(archivo, "    <TABLE BORDER=\"2\"  CELLBORDER=\"5\" CELLSPACING=\"3\"  WIDTH=\"300\" HEIGHT=\"200\">\n")
				var list_ext []EBR = admindisk.getlogics(List_part[i], paths)
				for j := 0; j < len(list_ext); j++ {
					contBlockLogic += 2
					es_acoupadoLogic = int(list_ext[j].EBR_start) + int(list_ext[j].EBR_size) + int(unsafe.Sizeof(EBR{}))
					if !(es_acoupadoLogic == int(list_ext[j+1].EBR_start)) {
						contBlockLogic += 1
					}

				}
				fmt.Fprintf(archivo, "<TR>\n")
				fmt.Fprintf(archivo, "           <TD COLSPAN=\"%s\" WIDTH=\"300\" HEIGHT=\"200\"> PARTICION EXTENDIDA <BR/> %s <BR/> %s </TD> \n", strconv.Itoa(contBlockLogic), str, porcentaje)
				fmt.Fprintf(archivo, "</TR>\n")
				fmt.Fprintf(archivo, "<TR>\n")
				es_acoupado = 0
				for j := 0; j < len(list_ext); j++ {
					if list_ext[j].EBR_status == '1' {
						fmt.Fprintf(archivo, "              <TD WIDTH=\"300\" HEIGHT=\"200\"> EBR </TD>\n")
						str1 = string(list_ext[j].EBR_name[:])
						var numero float64 = float64(list_ext[j].EBR_size)
						var es_entero float64 = (numero / float64(List_part[j].PART_size)) * 100
						var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
						ss := strconv.FormatFloat(es_porcentaje, 'f', 4, 64)
						var porcentaje string = ss + "% de la particion extendida"
						fmt.Fprintf(archivo, "                <TD WIDTH=\"300\" HEIGHT=\"200\"> PARTICION LOGICA <BR/> %s <BR/> %s </TD>\n", str1, porcentaje)
					}
					es_acoupadoLogic = int(list_ext[j].EBR_start) + int(list_ext[j].EBR_size) + int(unsafe.Sizeof(EBR{}))
					totLogic += int(list_ext[j].EBR_size)
					if !(es_acoupadoLogic == int(list_ext[j+1].EBR_start)) {
						var numero float64 = float64(List_part[i].PART_size) - float64(totLogic)
						var es_entero float64 = (numero / float64(List_part[i].PART_size)) * 100
						var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
						ss := strconv.FormatFloat(es_porcentaje, 'f', 4, 64)
						var porcentaje string = ss + "% de la particion extendida"
						fmt.Fprintf(archivo, "                   <TD WIDTH=\"300\" HEIGHT=\"200\">  LIBRE <BR/> %s </TD>\n", porcentaje)
					}
				}
				fmt.Fprintf(archivo, "		</TR>\n")
				fmt.Fprintf(archivo, "	</TABLE>\n")
				fmt.Fprintf(archivo, "</TD>\n")
			}
			es_acoupado = int(List_part[i].PART_start) + int(List_part[i].PART_size)
			tot += int(List_part[i].PART_size)
			if !(es_acoupado == int(List_part[i+1].PART_start)) {

				var numero float64 = float64(mbr.MBR_size) - float64(tot)
				var es_entero float64 = (numero / float64(mbr.MBR_size)) * 100
				var es_porcentaje float64 = math.Round(es_entero*100.0) / 100.0
				str1 := strconv.FormatFloat(es_porcentaje, 'f', 4, 64)
				var porcentaje string = str1 + "% del disco"
				fmt.Fprintf(archivo, "<TD ROWSPAN=\"3\" WIDTH=\"300\" HEIGHT=\"200\"> LIBRE <BR/> %s </TD>\n", porcentaje)
			}

		}
	}
	fmt.Fprintf(archivo, "</TR>\n")
	fmt.Fprintf(archivo, "</TABLE>>];\n")
	fmt.Fprintf(archivo, "label = \"Reporte DISK \n By: Kemel Ruano\" \n")
	fmt.Fprintf(archivo, "} \n")
	RUTAIMAGEN = rutaImagen
	exec.Command("dot", "-Tpdf", "-o", rutaImagen, rutaSinExtension).Run()

}

func (a Admin_UG) ViewsReporte(user string, pwd string, ids string) string {

	if Logeado.User != user && Logeado.Password == pwd && Logeado.Id == ids {
		return "NO EXISTE EL USUARIO"
	} else if Logeado.User == user && Logeado.Password != pwd && Logeado.Id == ids {
		return "CONTRASEÑA INCORRECTA"
	} else if Logeado.User == user && Logeado.Password == pwd && Logeado.Id != ids {
		return "ID INCORRECTO"
	}

	return "BIENVENIDO USUARIO"
}

func (a Admin_UG) DRUTE() string {
	return RUTAIMAGEN
}
