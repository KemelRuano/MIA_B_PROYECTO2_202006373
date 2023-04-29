package process

import (
	"unsafe"
)

type Mbr struct {
	MBR_size      int32
	MBR_time      int64
	MBR_asigndisk int32
	MBR_fit       byte
	MBR_Part_1    Partition
	MBR_Part_2    Partition
	MBR_Part_3    Partition
	MBR_Part_4    Partition
}

type Partition struct {
	PART_status byte
	PART_type   byte
	PART_fit    byte
	PART_start  int32
	PART_size   int32
	PART_name   [16]byte
}

func NewPartition() Partition {
	return Partition{
		PART_status: '0',
		PART_type:   '-',
		PART_fit:    '-',
		PART_start:  -1,
		PART_size:   0,
		PART_name:   [16]byte{},
	}
}

type EBR struct {
	EBR_status byte
	EBR_fit    byte
	EBR_start  int32
	EBR_size   int32
	EBR_next   int32
	EBR_name   [16]byte
}

func NewEBR() EBR {
	return EBR{
		EBR_status: '0',
		EBR_fit:    '-',
		EBR_start:  -1,
		EBR_size:   0,
		EBR_next:   -1,
		EBR_name:   [16]byte{},
	}
}

type Mount_id struct {
	Id       string
	Namedisk string
	No       int
	Mkfs     bool
}
type Mount struct {
	Disco string
	Path  string
	Cont  int
	ids   []Mount_id
}

func (m *Mount) AddId(id string, namedisk string, no int) {
	m.ids = append(m.ids, Mount_id{Id: id, Namedisk: namedisk, No: no})
}

type Inodes struct {
	I_uid   int32
	I_gid   int32
	I_size  int32
	I_atime int64
	I_ctime int64
	I_mtime int64
	I_block [16]int32
	I_type  byte
	I_perm  int32
}

func NewInodes() Inodes {
	return Inodes{
		I_uid:   -1,
		I_gid:   -1,
		I_size:  0,
		I_atime: 0,
		I_ctime: 0,
		I_mtime: 0,
		I_block: [16]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  '-',
		I_perm:  -1,
	}
}

type Superblock struct {
	S_filesystem_type   int32
	S_inodes_count      int32
	S_blocks_count      int32
	S_free_blocks_count int32
	S_free_inodes_count int32
	S_mtime             int64
	S_umtime            int64
	S_mnt_count         int32
	S_magic             int32
	S_inode_size        int32
	S_block_size        int32
	S_first_ino         int32
	S_first_blo         int32
	S_bm_inode_start    int32
	S_bm_block_start    int32
	S_inode_start       int32
	S_block_start       int32
}

func NewSuperblock() Superblock {
	return Superblock{
		S_filesystem_type:   0,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_blocks_count: 0,
		S_free_inodes_count: 0,
		S_magic:             0xEF53,
		S_inode_size:        int32(unsafe.Sizeof(Inodes{})),
		S_block_size:        int32(unsafe.Sizeof(Folderblock{})),
		S_first_ino:         0,
		S_first_blo:         0,
		S_bm_inode_start:    0,
		S_bm_block_start:    0,
		S_inode_start:       0,
		S_block_start:       0,
	}
}

type Content struct {
	B_name  [12]byte
	B_inodo int32
}

func NewFolder() Folderblock {
	return Folderblock{B_content: [4]Content{NewContent(), NewContent(), NewContent(), NewContent()}}
}
func NewContent() Content {
	return Content{B_name: [12]byte{}, B_inodo: -1}
}

type Folderblock struct {
	B_content [4]Content
}

type Fileblock struct {
	B_content [64]byte
}

type ActionList struct {
	CreaMkdisk string
}
