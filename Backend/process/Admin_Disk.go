package process

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type AdminDisk struct{}

var List_mount []Mount
var aumento int = 1

type Transition struct {
	partition int32
	start     int32
	end       int32
	before    int32
	after     int32
}

var startValue int

func (d AdminDisk) Mkdisk(paths string, size string, fit string, unit string) string {
	tamano, _ := strconv.Atoi(size)
	if tamano <= 0 {
		return "~~~ ERROR [MKDISK] ---- EL TAMANO DE DISCO DEBE SER MAYOR A 0"
	}
	if !strings.Contains(path.Base(paths), ".dsk") {
		return "~~~ ERROR [MKDISK] ---- EXTENSION DE DISCO INVALIDA VALIDOS [.dsk]"
	}
	carpetas := strings.Replace(paths, path.Base(paths), "", -1)
	os.MkdirAll(carpetas, 0755)
	if ExisteArchivo(paths) {
		d.LeerMMR(paths)
		return "~~~ ERROR [MKDISK] ---- EL DISCO YA EXISTE"
	}
	if unit == "m" {
		tamano = tamano * 1024 * 1024
	} else if unit == "k" {
		tamano = tamano * 1024
	} else if unit == "" {
		tamano = tamano * 1024 * 1024
	} else {
		return "~~~ ERROR [MKDISK] ---- UNIDAD DE DISCO INVALIDA"
	}
	var ajust byte
	if fit == "ff" {
		ajust = 'f'
	} else if fit == "wf" {
		ajust = 'w'
	} else if fit == "bf" {
		ajust = 'b'
	}

	// crear el archivo
	archivo, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0666)

	defer archivo.Close()

	if _, err := archivo.Write([]byte{0}); err != nil {
		panic(err)
	}
	if _, err := archivo.Seek(int64(tamano-1), 0); err != nil {
		panic(err)
	}
	if _, err := archivo.Write([]byte{0}); err != nil {
		panic(err)
	}

	if _, err := archivo.Seek(0, 0); err != nil {
		panic(err)
	}
	MBR := Mbr{}
	MBR.MBR_size = int32(tamano)
	MBR.MBR_fit = ajust
	MBR.MBR_time = time.Now().Unix()
	MBR.MBR_asigndisk = int32(rand.Intn(501))
	MBR.MBR_Part_1 = NewPartition()
	MBR.MBR_Part_2 = NewPartition()
	MBR.MBR_Part_3 = NewPartition()
	MBR.MBR_Part_4 = NewPartition()
	binary.Write(archivo, binary.LittleEndian, &MBR)
	return "█████████ [MKDISK] ----  DISCO CREADO CON EXITO █████████"

}
func (d AdminDisk) Rmdisk(paths string) string {
	if !ExisteArchivo(paths) {
		return "~~~ ERROR [RMDISK] ---- EL DISCO NO EXISTE"
	}
	return "DESEA ELIMINAR EL DISCO [Y/N]"
}

func (d AdminDisk) Fdisk(size string, paths string, name string, typed string, unit string, fit string) string {
	startValue = 0
	tamano, _ := strconv.Atoi(size)
	if tamano <= 0 {
		return "~~~ ERROR [FDISK] ---- EL TAMANO DE PARTICION DEBE SER MAYOR A 0"
	}
	var tipo_part byte
	is_type := false
	if typed == "e" {
		tipo_part = 'e'
	} else if typed == "l" {
		is_type = true
		tipo_part = 'l'
	} else if typed == "p" {
		tipo_part = 'p'
	}
	var ajust byte
	if fit == "ff" {
		ajust = 'f'
	} else if fit == "wf" {
		ajust = 'w'
	} else if fit == "bf" {
		ajust = 'b'
	}
	if unit == "m" {
		tamano = tamano * 1024 * 1024
	} else if unit == "k" {
		tamano = tamano * 1024
	} else if unit == "" {
		tamano = tamano * 1024
	}
	var Disco Mbr
	archivo, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0666)
	archivo.Seek(0, 0)
	binary.Read(archivo, binary.LittleEndian, &Disco)
	defer archivo.Close()
	partitions := d.List_Partition(Disco)
	between := []Transition{}
	used := 0
	ext := 0
	c := int32(0)
	base := int32(unsafe.Sizeof(Disco))
	var extended Partition
	for _, prttn := range partitions {
		if prttn.PART_status == '1' {
			var trn Transition
			trn.partition = c
			trn.start = prttn.PART_start
			trn.end = prttn.PART_start + prttn.PART_size
			trn.before = trn.start - base
			base = trn.end
			if used != 0 {
				between[used-1].after = trn.start - (between[used-1].end)
			}
			between = append(between, trn)
			used++

			if prttn.PART_type == 'e' {
				ext++
				extended = prttn
			}
		}
		if used == 4 && !is_type {
			return "~~~ ERROR [FDISK] ---- NO SE PUEDE CREAR MAS PARTICIONES"
		} else if ext == 1 && tipo_part == 'e' {
			return "~~~ ERROR [FDISK] ---- YA EXISTE UNA PARTICION EXTENDIDA"
		}
		c++
	}
	if ext == 0 && tipo_part == 'l' {
		return "~~~ ERROR [FDISK] ---- NO EXISTE UNA PARTICION EXTENDIDA"
	}

	if used != 0 {
		between[len(between)-1].after = Disco.MBR_size - (between[len(between)-1].end)
	}

	_, err := d.findby(Disco, name, paths)
	if err == nil {
		return "~~~ ERROR [FDISK] ---- YA EXISTE UNA PARTICION CON ESE NOMBRE"
	}

	transitions := NewPartition()
	transitions.PART_status = '1'
	transitions.PART_fit = ajust
	copy(transitions.PART_name[:], name)
	transitions.PART_size = int32(tamano)
	transitions.PART_type = tipo_part
	fmt.Println(transitions)
	if is_type {
		return d.logic(transitions, extended, paths)
	}

	Disco, err = d.adjust(Disco, transitions, between, partitions, used)
	if err != nil {
		return "~~~ ERROR [FDISK] ---- NO HAY MAS ESPACIO, NO SE PUDO CREAR LA PARTICION"
	}

	bfile, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0666)
	defer bfile.Close()
	binary.Write(bfile, binary.LittleEndian, &Disco)
	if tipo_part == 'p' {
		return "█████████ [FDISK] ----  PARTICION PRIMARIA CREADA CON EXITO █████████"
	}
	if tipo_part == 'e' {
		ebr := NewEBR()
		ebr.EBR_start = int32(startValue)
		bfile.Seek(int64(startValue), 0)
		binary.Write(bfile, binary.LittleEndian, &ebr)
		return "█████████ [FDISK] ----  PARTICION EXTENDIDA CREADA CON EXITO █████████"
	}
	return ""
}

func (d AdminDisk) LeerMMR(paths string) {

	archivo, _ := os.Open(paths)
	defer archivo.Close()
	var read_MBR Mbr
	binary.Read(archivo, binary.LittleEndian, &read_MBR)
	fmt.Println("---------------------MBR---------------------")
	fmt.Println("mbr_size: ", read_MBR.MBR_size)
	fmt.Println("mbr_fit: ", string(read_MBR.MBR_fit))
	fmt.Println("mbr_time: ", read_MBR.MBR_time)
	fmt.Println("mbr_asigndisk: ", read_MBR.MBR_asigndisk)
	fmt.Println("tamano del mbr", unsafe.Sizeof(read_MBR))

	List_read := d.List_Partition(read_MBR)
	for i := 0; i < 4; i++ {
		if List_read[i].PART_status == '1' {
			if List_read[i].PART_type == 'p' {
				fmt.Println("--------------PARTICION PRIMARIA --------------")
				fmt.Println("	part_status: ", string(List_read[i].PART_status))
				fmt.Println("	part_type:   ", string(List_read[i].PART_type))
				fmt.Println("	part_fit: ", string(List_read[i].PART_fit))
				fmt.Println("	part_start: ", List_read[i].PART_start)
				fmt.Println("	part_size: ", List_read[i].PART_size)
				fmt.Println("	part_name: ", string(List_read[i].PART_name[:]))
			} else if List_read[i].PART_type == 'e' {
				fmt.Println("--------------PARTICION EXTENDIDA --------------")
				fmt.Println("	part_status: ", string(List_read[i].PART_status))
				fmt.Println("	part_type:   ", string(List_read[i].PART_type))
				fmt.Println("	part_fit: ", string(List_read[i].PART_fit))
				fmt.Println("	part_start: ", List_read[i].PART_start)
				fmt.Println("	part_size: ", List_read[i].PART_size)
				fmt.Println("	part_name: ", string(List_read[i].PART_name[:]))
				list_ext := d.getlogics(List_read[i], paths)
				for _, ebr := range list_ext {
					fmt.Println("--------------PARTICION LOGICA --------------")
					fmt.Println("	part_status: ", string(ebr.EBR_status))
					fmt.Println("	part_fit: ", string(ebr.EBR_fit))
					fmt.Println("	part_start: ", ebr.EBR_start)
					fmt.Println("	part_size: ", ebr.EBR_size)
					fmt.Println("	part_next: ", ebr.EBR_next)
					fmt.Println("	part_name: ", string(ebr.EBR_name[:]))
				}
			}
		}
	}

}

func (d AdminDisk) List_Partition(mbr Mbr) []Partition {
	List := []Partition{}
	List = append(List, mbr.MBR_Part_1)
	List = append(List, mbr.MBR_Part_2)
	List = append(List, mbr.MBR_Part_3)
	List = append(List, mbr.MBR_Part_4)
	return List
}

func ExisteArchivo(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, _ := os.Create(path)
		defer file.Close()
	} else {
		return true
	}
	return false
}

func (d AdminDisk) findby(mbr Mbr, name string, path string) (Partition, error) {
	var partitions [4]Partition
	partitions[0] = mbr.MBR_Part_1
	partitions[1] = mbr.MBR_Part_2
	partitions[2] = mbr.MBR_Part_3
	partitions[3] = mbr.MBR_Part_4

	ext := false
	var extended Partition
	var bytes [16]byte
	copy(bytes[:], []byte(name))
	for _, partition1 := range partitions {
		if partition1.PART_status == '1' {
			if partition1.PART_name == bytes {
				return partition1, nil
			} else if partition1.PART_type == 'e' {
				ext = true
				extended = partition1
			}
		}
	}
	if ext {
		var ebrs []EBR = d.getlogics(extended, path)
		for _, ebr := range ebrs {
			if ebr.EBR_status == '1' {
				if ebr.EBR_name == bytes {
					var tmp Partition
					tmp.PART_status = '1'
					tmp.PART_type = 'l'
					tmp.PART_fit = ebr.EBR_fit
					tmp.PART_start = ebr.EBR_start
					tmp.PART_size = ebr.EBR_size
					tmp.PART_name = ebr.EBR_name
					return tmp, nil
				}
			}
		}

	}

	return Partition{}, fmt.Errorf("[fdisk]---- La partición no existe")
}

func (d *AdminDisk) getlogics(partition Partition, path string) []EBR {
	var ebrs []EBR
	archivo, _ := os.OpenFile(path, os.O_RDWR, 0666)
	defer archivo.Close()
	archivo.Seek(0, 0)
	var tmp = NewEBR()
	archivo.Seek(int64(partition.PART_start), 0)
	binary.Read(archivo, binary.LittleEndian, &tmp)

	for {
		if !(tmp.EBR_status == '0' && tmp.EBR_next == -1) {
			if tmp.EBR_status != '0' {
				ebrs = append(ebrs, tmp)
			}
			archivo.Seek(int64(tmp.EBR_next), 0)
			binary.Read(archivo, binary.LittleEndian, &tmp)

		} else {
			break
		}

	}
	return ebrs
}

func (d AdminDisk) logic(partition Partition, ep Partition, p string) string {
	var nlogic EBR
	nlogic.EBR_status = '1'
	nlogic.EBR_fit = partition.PART_fit
	nlogic.EBR_size = partition.PART_size
	copy(nlogic.EBR_name[:], partition.PART_name[:])
	nlogic.EBR_next = -1

	archivo, _ := os.OpenFile(p, os.O_RDWR, 0666)
	defer archivo.Close()
	archivo.Seek(0, 0)

	var tmp EBR
	archivo.Seek(int64(ep.PART_start), 0)
	binary.Read(archivo, binary.LittleEndian, &tmp)
	size := 0
	for {
		size += int(tmp.EBR_size) + binary.Size(EBR{})
		if tmp.EBR_status == '0' && tmp.EBR_next == -1 {
			nlogic.EBR_start = tmp.EBR_start
			nlogic.EBR_next = nlogic.EBR_start + nlogic.EBR_size + int32(binary.Size(EBR{}))
			if (ep.PART_size - int32(size)) <= nlogic.EBR_size {
				return " ~~~  ERROR [FDISK] --- NO SE PUEDE CREAR MAS PARTICIONES LOGICA "

			}
			archivo.Seek(int64(nlogic.EBR_start), 0)
			binary.Write(archivo, binary.LittleEndian, &nlogic)
			archivo.Seek(int64(nlogic.EBR_next), 0)
			var addLogic EBR
			addLogic.EBR_status = '0'
			addLogic.EBR_next = -1
			addLogic.EBR_start = nlogic.EBR_next
			archivo.Seek(int64(addLogic.EBR_start), 0)
			binary.Write(archivo, binary.LittleEndian, &addLogic)
			return "███████   [FDISK] ---- PARTICION LOGICA CREADA CORRECTAMENTE ███████"
		}
		archivo.Seek(int64(tmp.EBR_next), 0)
		binary.Read(archivo, binary.LittleEndian, &tmp)

	}
}

func (d *AdminDisk) adjust(mbr Mbr, p Partition, t []Transition, ps []Partition, u int) (Mbr, error) {
	if u == 0 {
		p.PART_start = int32(unsafe.Sizeof(mbr))
		startValue = int(p.PART_start)
		mbr.MBR_Part_1 = p
		return mbr, nil
	} else {
		var toUse Transition
		var c int = 0
		for _, tr := range t {
			if c == 0 {
				toUse = tr
				c++
				continue
			}
			if mbr.MBR_fit == 'f' {
				if toUse.before >= p.PART_size || toUse.after >= p.PART_size {
					break
				}
				toUse = tr
			} else if mbr.MBR_fit == 'b' {
				if toUse.before < p.PART_size || toUse.after <= p.PART_size {
					toUse = tr
				} else {
					if tr.before >= p.PART_size || tr.after >= p.PART_size {
						b1 := toUse.before - p.PART_size
						a1 := toUse.after - p.PART_size
						b2 := tr.before - p.PART_size
						a2 := tr.after - p.PART_size

						if (b1 < b2 && b1 < a2) || (a1 < b2 && a1 < a2) {
							c++
							continue
						}
						toUse = tr
					}
				}

			} else if mbr.MBR_fit == 'w' {

				if !(toUse.before >= p.PART_size) || !(toUse.after >= p.PART_size) {
					toUse = tr
				} else {
					if tr.before >= p.PART_size || tr.after >= p.PART_size {
						b1 := toUse.before - p.PART_size
						a1 := toUse.after - p.PART_size
						b2 := tr.before - p.PART_size
						a2 := tr.after - p.PART_size

						if (b1 > b2 && b1 > a2) || (a1 > b2 && a1 > a2) {
							c++
							continue
						}
						toUse = tr
					}
				}
			}
			c++
		}

		if toUse.before >= p.PART_size || toUse.after >= p.PART_size {
			if mbr.MBR_fit == 'f' {
				if toUse.before >= p.PART_size {
					p.PART_start = toUse.start - toUse.before
					startValue = int(p.PART_start)
				} else {
					p.PART_start = toUse.end
					startValue = int(p.PART_start)
				}
			} else if mbr.MBR_fit == 'b' {
				b1 := toUse.before - p.PART_size
				a1 := toUse.after - p.PART_size
				if (toUse.before >= p.PART_size && b1 < a1) || !(toUse.after >= p.PART_start) {
					p.PART_start = toUse.start - toUse.before
					startValue = int(p.PART_start)
				} else {
					p.PART_start = toUse.end
					startValue = int(p.PART_start)
				}
			} else if mbr.MBR_fit == 'w' {
				b1 := toUse.before - p.PART_size
				a1 := toUse.after - p.PART_size
				if (toUse.before >= p.PART_size && b1 > a1) || !(toUse.after >= p.PART_start) {
					p.PART_start = toUse.start - toUse.before
					startValue = int(p.PART_start)
				} else {
					p.PART_start = toUse.end
					startValue = int(p.PART_start)
				}
			}

			var partitions [4]Partition
			for i := 0; i < len(ps); i++ {
				copy(partitions[:], ps[:])
			}

			for i, partition := range partitions {
				if partition.PART_status == '0' {
					partitions[i] = p
					break
				}
			}

			var aux Partition
			for i := 3; i >= 0; i-- {
				for j := 0; j < i; j++ {
					if partitions[j].PART_start > partitions[j+1].PART_start {
						aux = partitions[j+1]
						partitions[j+1] = partitions[j]
						partitions[j] = aux
					}
				}
			}

			for i := 3; i >= 0; i-- {
				for j := 0; j < i; j++ {
					if partitions[j].PART_status == '0' {
						aux = partitions[j]
						partitions[j] = partitions[j+1]
						partitions[j+1] = aux
					}
				}
			}

			mbr.MBR_Part_1 = partitions[0]
			mbr.MBR_Part_2 = partitions[1]
			mbr.MBR_Part_3 = partitions[2]
			mbr.MBR_Part_4 = partitions[3]
			return mbr, nil
		} else {
			return Mbr{}, errors.New("[fdisk]---- no hay suficiente espacio para realizar la particion")
		}
	}
}

func (d AdminDisk) MOUNT(paths string, name string) string {
	IdLIst := []byte{'1', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

	namedisk := strings.Replace(path.Base(paths), ".dsk", "", -1)

	var Disco Mbr
	file, _ := os.Open(paths)
	defer file.Close()
	file.Seek(0, 0)
	binary.Read(file, binary.LittleEndian, &Disco)

	var partitions [4]Partition
	partitions[0] = Disco.MBR_Part_1
	partitions[1] = Disco.MBR_Part_2
	partitions[2] = Disco.MBR_Part_3
	partitions[3] = Disco.MBR_Part_4
	encontrado_P := false

	for _, buscadoPart := range partitions {
		if buscadoPart.PART_type == 'p' {
			var bytes [16]byte
			copy(bytes[:], []byte(name))
			if buscadoPart.PART_name == bytes {
				encontrado_P = true
				break
			}
		} else if buscadoPart.PART_type == 'e' {
			var ebrs []EBR = d.getlogics(buscadoPart, paths)
			for _, buscadoLog := range ebrs {
				var bytes [16]byte
				copy(bytes[:], []byte(name))
				if buscadoLog.EBR_name == bytes {
					encontrado_P = true
					break
				}
			}

		}
	}

	parte1 := "73"
	if encontrado_P {
		es_mount := 0
		repetido := false
		var cont_L int
		for i := 0; i < len(List_mount); i++ {
			if List_mount[i].Disco == namedisk {
				repetido = true
				parte1 += strconv.Itoa(List_mount[i].Cont)
				terminar := false

				for n := 1; n < len(IdLIst); n++ {
					for y := 0; y < len(List_mount[i].ids); y++ {
						if List_mount[i].ids[y].Namedisk == name {
							return "~~~ ERROR [MOUNT] --- YA ESTA MONTADO"
						}
						es_mount = List_mount[i].ids[y].No
						es_mount++
						if n == es_mount {
							parte1 += string(IdLIst[n])
							terminar = true
							List_mount[i].ids[y].No = es_mount
							break
						}
					}
					if terminar {
						break
					}
				}
				List_mount[i].AddId(parte1, name, cont_L)
				break
			}
		}

		if !repetido {
			List_mount = append(List_mount, Mount{})
			List_mount[len(List_mount)-1].Disco = namedisk
			List_mount[len(List_mount)-1].Path = paths
			List_mount[len(List_mount)-1].Cont = aumento
			fmt.Println(aumento)
			parte1 += strconv.Itoa(List_mount[len(List_mount)-1].Cont)
			for i := 1; i < len(IdLIst); i++ {
				if i == 1 {
					parte1 += string(IdLIst[i])
					break
				}
			}
			aumento++
			List_mount[len(List_mount)-1].AddId(parte1, name, 1)

		}

	} else {
		return "~~~ ERROR [MOUNT] --- NO SE ENCONTRO LA PARTICION"
	}
	d.verVector()
	return "██████ [MOUNT] --- SE MONTÓ LA PARTICION CON EXITO ██████"
}

func (d AdminDisk) verVector() {
	for i := 0; i < len(List_mount); i++ {
		fmt.Println("Disco:", List_mount[i].Disco)
		fmt.Println("Path: ", List_mount[i].Path)
		for j := 0; j < len(List_mount[i].ids); j++ {
			fmt.Println(List_mount[i].ids[j].Id)
			fmt.Println("Name: ", List_mount[i].ids[j].Namedisk)

		}
	}
}

func (d AdminDisk) MKFS(types string, id string) string {
	envio := ""
	if types == "full" {
		envio = "██████ [MKFS] --- SE REALIZARA UN FORMATEO COMPLETO ██████"
	}
	paths := ""
	var particion Partition
	particion, err := d.EncontrarParticion(id, &paths)
	if err != nil {
		envio += "~~~ ERROR [MKFS] --- NO HAY DISCOS MONTADOS"
		return envio
	}

	ext2 := (particion.PART_size - int32(unsafe.Sizeof(Superblock{}))) / (4 + int32(unsafe.Sizeof(Inodes{})) + 3*int32(unsafe.Sizeof(Fileblock{})))
	fmt.Println("numero de bloqies: ", ext2)
	var superbloque Superblock
	superbloque.S_mtime = time.Now().Unix()
	superbloque.S_umtime = time.Now().Unix()
	superbloque.S_mnt_count = 1
	superbloque.S_filesystem_type = 2
	superbloque.S_inodes_count = ext2
	superbloque.S_blocks_count = ext2 * 3
	superbloque.S_free_blocks_count = ext2 * 3
	superbloque.S_free_inodes_count = ext2
	d.Format_ext2(superbloque, particion, int(ext2), paths, id)
	envio += "██████ [MKFS] --- SE REALIZO EL FORMATEO CON EXITO ██████"
	return envio
}

func (d AdminDisk) Format_ext2(superbloque Superblock, particion Partition, bloques int, paths string, ids string) {
	superbloque.S_bm_inode_start = particion.PART_start + int32(unsafe.Sizeof(Superblock{}))
	superbloque.S_bm_block_start = superbloque.S_bm_inode_start + int32(bloques)
	superbloque.S_inode_start = superbloque.S_bm_block_start + (3 * int32(bloques))
	superbloque.S_block_start = superbloque.S_inode_start + (int32(unsafe.Sizeof(Inodes{})) * int32(bloques))
	var tmp byte = '0'
	leer, _ := os.OpenFile(paths, os.O_RDWR|os.O_CREATE, 0666)
	defer leer.Close()
	leer.Seek(int64(particion.PART_start), 0)
	binary.Write(leer, binary.LittleEndian, &superbloque)
	leer.Seek(int64(superbloque.S_bm_inode_start), 0)
	for i := 0; i < bloques; i++ {
		binary.Write(leer, binary.LittleEndian, &tmp)
	}
	leer.Seek(int64(superbloque.S_bm_block_start), 0)
	for e := 0; e < (3 * bloques); e++ {
		binary.Write(leer, binary.LittleEndian, &tmp)
	}

	var inodos Inodes = NewInodes()
	leer.Seek(int64(superbloque.S_inode_start), 0)
	for i := 0; i < bloques; i++ {
		binary.Write(leer, binary.LittleEndian, &inodos)
	}
	var bloqueCarpetass Folderblock
	leer.Seek(int64(superbloque.S_block_start), 0)
	for i := 0; i < (3 * bloques); i++ {
		binary.Write(leer, binary.LittleEndian, &bloqueCarpetass)
	}
	var readsuper Superblock
	supblock, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer supblock.Close()
	supblock.Seek(int64(particion.PART_start), 0)
	binary.Read(supblock, binary.LittleEndian, &readsuper)
	var inodo Inodes = NewInodes()
	inodo.I_uid = 1
	inodo.I_gid = 1
	inodo.I_size = 0
	inodo.I_atime = superbloque.S_umtime
	inodo.I_ctime = superbloque.S_umtime
	inodo.I_mtime = superbloque.S_umtime
	inodo.I_block[0] = 0
	inodo.I_type = 48
	inodo.I_perm = 664

	var bloke Folderblock = NewFolder()
	copy(bloke.B_content[0].B_name[:], []byte("."))
	bloke.B_content[0].B_inodo = 0
	copy(bloke.B_content[1].B_name[:], []byte(".."))
	bloke.B_content[1].B_inodo = 0
	copy(bloke.B_content[2].B_name[:], []byte("users.txt"))
	bloke.B_content[2].B_inodo = 1
	copy(bloke.B_content[3].B_name[:], []byte("-"))
	bloke.B_content[3].B_inodo = -1

	data := "1,G,root\n1,U,root,root,123\n"
	var inodotemp Inodes = NewInodes()
	inodotemp.I_uid = 1
	inodotemp.I_gid = 1
	inodotemp.I_size = int32(len(data)) + int32(unsafe.Sizeof(Folderblock{}))
	inodotemp.I_atime = superbloque.S_umtime
	inodotemp.I_ctime = superbloque.S_umtime
	inodotemp.I_mtime = superbloque.S_umtime
	inodotemp.I_block[0] = 1
	inodotemp.I_type = 49
	inodotemp.I_perm = 664

	inodo.I_size = inodotemp.I_size + int32(unsafe.Sizeof(Folderblock{})) + int32(unsafe.Sizeof(Inodes{}))

	var fileb Fileblock
	copy(fileb.B_content[:], []byte(data))

	bfiles, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer bfiles.Close()
	var caracter byte = 49
	bfiles.Seek(int64(superbloque.S_bm_inode_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &caracter)
	binary.Write(bfiles, binary.LittleEndian, &caracter)

	bfiles.Seek(int64(superbloque.S_bm_block_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &caracter)
	binary.Write(bfiles, binary.LittleEndian, &caracter)

	bfiles.Seek(int64(superbloque.S_inode_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &inodo)
	bfiles.Seek(int64(superbloque.S_inode_start+int32(unsafe.Sizeof(Inodes{}))), 0)
	binary.Write(bfiles, binary.LittleEndian, &inodotemp)

	bfiles.Seek(int64(superbloque.S_block_start), 0)
	binary.Write(bfiles, binary.LittleEndian, &bloke)
	bfiles.Seek(int64(superbloque.S_block_start+int32(unsafe.Sizeof(Fileblock{}))), 0)
	binary.Write(bfiles, binary.LittleEndian, &fileb)

}

func (d AdminDisk) EncontrarParticion(id string, p *string) (Partition, error) {
	nombreParticion := ""
	paths := ""

	for i := 0; i < len(List_mount); i++ {
		for j := 0; j < len(List_mount[i].ids); j++ {
			if List_mount[i].ids[j].Id == id {
				nombreParticion = List_mount[i].ids[j].Namedisk
				paths = List_mount[i].Path
				break
			}

		}
	}

	*p = paths
	var mbr Mbr
	file, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer file.Close()
	file.Seek(0, 0)
	binary.Read(file, binary.LittleEndian, &mbr)
	return d.findby(mbr, nombreParticion, paths)
}

func (d AdminDisk) Esta_formateado(partition Partition, paths string) bool {
	var super Superblock = NewSuperblock()
	file, _ := os.OpenFile(paths, os.O_RDWR, 0666)
	defer file.Close()
	file.Seek(0, 0)
	file.Seek(int64(partition.PART_start), 0)
	binary.Read(file, binary.LittleEndian, &super)

	return super.S_filesystem_type == int32(2)
}
