package wasm

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/go-interpreter/wagon/wasm/internal/readpos"
	"github.com/go-interpreter/wagon/wasm/leb128"
)

// SectionID is a 1-byte code that encodes the section code of both known and custom sections.
type SectionID uint8

const (
	SectionIDCustom   SectionID = 0
	SectionIDType     SectionID = 1
	SectionIDImport   SectionID = 2
	SectionIDFunction SectionID = 3
	SectionIDTable    SectionID = 4
	SectionIDMemory   SectionID = 5
	SectionIDGlobal   SectionID = 6
	SectionIDExport   SectionID = 7
	SectionIDStart    SectionID = 8
	SectionIDElement  SectionID = 9
	SectionIDCode     SectionID = 10
	SectionIDData     SectionID = 11
)

// Section is a declared section in a WASM module.
type Section struct {
	Start int64
	End   int64

	ID SectionID
	// Size of this section in bytes
	PayloadLen uint32
	// Section name, empty if id != 0
	Name  string
	Bytes []byte
}

type ErrInvalidSectionID SectionID

func (e ErrInvalidSectionID) Error() string {
	return fmt.Sprintf("wasm: invalid section ID %d", e)
}

var ErrUnsupportedSection = errors.New("wasm: unsupported section")

func (m *Module) readSection(r *readpos.ReadPos) (bool, error) {
	s := Section{}
	var err error
	var id uint32

	if id, err = leb128.ReadVarUint32(r); err != nil {
		if err == io.EOF {
			return true, nil
		}
		return false, err
	}
	s.ID = SectionID(id)

	if s.PayloadLen, err = leb128.ReadVarUint32(r); err != nil {
		return false, nil
	}

	payloadDataLen := s.PayloadLen

	if s.ID == SectionIDCustom {
		nameLen, nameLenSize, err := leb128.ReadVarUint32Size(r)
		if err != nil {
			return false, err
		}
		payloadDataLen -= uint32(nameLenSize)
		if s.Name, err = readString(r, uint(nameLen)); err != nil {
			return false, err
		}

		payloadDataLen -= uint32(len(s.Name))
	}

	logger.Printf("Section payload length: %d", payloadDataLen)

	s.Start = r.CurPos

	s.Bytes = make([]byte, payloadDataLen)
	_, err = io.ReadFull(r, s.Bytes)
	if err != nil {
		return false, err
	}

	s.End = r.CurPos

	sectionReader := bytes.NewReader(s.Bytes)

	switch s.ID {
	case SectionIDCustom:
		logger.Println("section custom")
		if s.Name == "name" {

		}
		break
		// TODO: Read custom sections
	case SectionIDType:
		logger.Println("section type")
		if err = m.readSectionTypes(sectionReader); err == nil {
			m.Types.Section = s
		}
	case SectionIDImport:
		logger.Println("section import")
		if err = m.readSectionImports(sectionReader); err == nil {
			m.Import.Section = s
		}
	case SectionIDFunction:
		logger.Println("section function")
		if err = m.readSectionFunctions(sectionReader); err == nil {
			m.Function.Section = s
		}
	case SectionIDTable:
		logger.Println("section table")
		if err = m.readSectionTables(sectionReader); err == nil {
			m.Table.Section = s
		}
	case SectionIDMemory:
		logger.Println("section memory")
		if err = m.readSectionMemories(sectionReader); err == nil {
			m.Memory.Section = s
		}
	case SectionIDGlobal:
		logger.Println("section global")
		if err = m.readSectionGlobals(sectionReader); err == nil {
			m.Global.Section = s
		}
	case SectionIDExport:
		logger.Println("section export")
		if err = m.readSectionExports(sectionReader); err == nil {
			m.Export.Section = s
		}
	case SectionIDStart:
		logger.Println("section start")
		if err = m.readSectionStart(sectionReader); err == nil {
			m.Start.Section = s
		}
	case SectionIDElement:
		logger.Println("section element")
		if err = m.readSectionElements(sectionReader); err == nil {
			m.Elements.Section = s
		}
	case SectionIDCode:
		logger.Println("section code")
		if err = m.readSectionCode(sectionReader); err == nil {
			m.Code.Section = s
		}
	case SectionIDData:
		logger.Println("section data")
		if err = m.readSectionData(sectionReader); err == nil {
			m.Data.Section = s
		}
	default:
		return false, ErrInvalidSectionID(s.ID)
	}

	return false, err
}

// SectionTypes declares all function signatures that will be used in a module.
type SectionTypes struct {
	Section
	Entries []FunctionSig
}

func (m *Module) readSectionTypes(r io.Reader) error {
	s := &SectionTypes{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	s.Entries = make([]FunctionSig, int(count))

	for i := uint32(0); i < count; i++ {
		if s.Entries[i], err = readFunction(r); err != nil {
			return err
		}

	}

	m.Types = s

	return nil
}

// SectionImports declares all imports that will be used in the module.
type SectionImports struct {
	Section
	Entries []ImportEntry
}

func (m *Module) readSectionImports(r io.Reader) error {
	s := &SectionImports{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make([]ImportEntry, count)

	for i := uint32(0); i < count; i++ {
		var err error
		s.Entries[i], err = readImportEntry(r)
		if err != nil {
			return err
		}
	}

	m.Import = s
	return nil
}

func readImportEntry(r io.Reader) (ImportEntry, error) {
	i := ImportEntry{}

	modLen, err := leb128.ReadVarUint32(r)
	if err != nil {
		return i, err
	}

	if i.ModuleStr, err = readString(r, uint(modLen)); err != nil {
		return i, err
	}

	fieldLen, err := leb128.ReadVarUint32(r)
	if err != nil {
		return i, err
	}

	if i.FieldStr, err = readString(r, uint(fieldLen)); err != nil {
		return i, err
	}

	if i.Kind, err = readExternal(r); err != nil {
		return i, err
	}

	var ptr *GlobalVar

	switch i.Kind {
	case ExternalFunction:
		logger.Println("importing function")
		i.Type, err = leb128.ReadVarUint32(r)
	case ExternalTable:
		logger.Println("importing table")
		i.Type, err = readTable(r)
	case ExternalMemory:
		logger.Println("importing memory")
		i.Type, err = readMemory(r)
	case ExternalGlobal:
		logger.Println("importing global var")
		ptr, err = readGlobalVar(r)
		if ptr != nil {
			i.Type = *ptr
		}
	default:
		return i, ErrInvalidExternal(i.Kind)
	}

	return i, err
}

// SectionFunction declares the signature of all functions defined in the module (in the code section)
type SectionFunctions struct {
	Section
	// Sequences of indices into (FunctionSignatues).Entries
	Types []uint32
}

func (m *Module) readSectionFunctions(r io.Reader) error {
	s := &SectionFunctions{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	for i := uint32(0); i < count; i++ {
		t, err := leb128.ReadVarUint32(r)
		if err != nil {
			return err
		}
		s.Types = append(s.Types, t)
	}

	m.Function = s
	return nil
}

// SectionTables describes all tables declared by a module.
type SectionTables struct {
	Section
	Entries []Table
}

func (m *Module) readSectionTables(r io.Reader) error {
	s := &SectionTables{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make([]Table, count)

	for i := uint32(0); i < count; i++ {
		s.Entries[i], err = readTable(r)
		if err != nil {
			return err
		}
	}

	m.Table = s
	return err
}

// SectionMemories describes all linaer memories used by a module.
type SectionMemories struct {
	Section
	Entries []Memory
}

func (m *Module) readSectionMemories(r io.Reader) error {
	s := &SectionMemories{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	s.Entries = make([]Memory, count)

	for i := uint32(0); i < count; i++ {
		s.Entries[i], err = readMemory(r)
		if err != nil {
			return err
		}
	}

	m.Memory = s
	return err
}

// SectionGlobals defines the value of all global variables declared in a module.
type SectionGlobals struct {
	Section
	Globals []GlobalEntry
}

func (m *Module) readSectionGlobals(r io.Reader) error {
	s := &SectionGlobals{}

	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Globals = make([]GlobalEntry, count)

	logger.Printf("%d global entries\n", count)
	for i := uint32(0); i < count; i++ {
		s.Globals[i], err = readGlobalEntry(r)
		if err != nil {
			return err
		}
	}

	m.Global = s
	return nil
}

// GlobalEntry declares a global variable.
type GlobalEntry struct {
	Type *GlobalVar // GlobalVar value stores the type and mutability of the var
	Init []byte     // Init stores the initial value of the global variable
}

func readGlobalEntry(r io.Reader) (e GlobalEntry, err error) {
	logger.Println("reading global_type")
	e.Type, err = readGlobalVar(r)
	if err != nil {
		logger.Println("Error!")
		return
	}
	logger.Println("reading init expr")

	// init_expr is delimited by opcode "end" (0x0b)
	e.Init, err = readInitExpr(r)
	logger.Println("Value:", e.Init)
	return e, err
}

// SectionExports declares the export section of a module
type SectionExports struct {
	Section
	Entries map[string]ExportEntry
}

func (m *Module) readSectionExports(r io.Reader) error {
	s := &SectionExports{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make(map[string]ExportEntry, count)

	for i := uint32(0); i < count; i++ {
		entry, err := readExportEntry(r)
		if err != nil {
			return err
		}

		if _, exists := s.Entries[entry.FieldStr]; exists {
			return errors.New("Duplicate export entry.")
		}
		s.Entries[entry.FieldStr] = entry
	}

	m.Export = s
	return nil
}

// ExportEntry represents an exported entry by the module
type ExportEntry struct {
	FieldStr string
	Kind     External
	Index    uint32
}

func readExportEntry(r io.Reader) (ExportEntry, error) {
	e := ExportEntry{}
	fieldLen, err := leb128.ReadVarUint32(r)

	if e.FieldStr, err = readString(r, uint(fieldLen)); err != nil {
		return e, err
	}

	if e.Kind, err = readExternal(r); err != nil {
		return e, err
	}

	e.Index, err = leb128.ReadVarUint32(r)

	return e, err
}

// SectionStartFunction represents the start function section.
type SectionStartFunction struct {
	Section
	Index uint32 // The function index of the start function
}

func (m *Module) readSectionStart(r io.Reader) error {
	s := &SectionStartFunction{}
	var err error

	s.Index, err = leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	m.Start = s
	return nil
}

// SectionElements describes the initial contents of a table's elements.
type SectionElements struct {
	Section
	Entries []ElementSegment
}

func (m *Module) readSectionElements(r io.Reader) error {
	s := &SectionElements{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Entries = make([]ElementSegment, count)

	for i := uint32(0); i < count; i++ {
		s.Entries[i], err = readElementSegment(r)
		if err != nil {
			return err
		}
	}

	m.Elements = s
	return nil
}

type ElementSegment struct {
	// should be 0
	Index uint32
	// should return i32, computes the offset at which to place the elements
	Offset []byte
	Elems  []uint32
}

func readElementSegment(r io.Reader) (ElementSegment, error) {
	s := ElementSegment{}
	var err error

	if s.Index, err = leb128.ReadVarUint32(r); err != nil {
		return s, err
	}
	if s.Offset, err = readInitExpr(r); err != nil {
		return s, err
	}

	numElems, err := leb128.ReadVarUint32(r)
	if err != nil {
		return s, err
	}

	for i := uint32(0); i < numElems; i++ {
		e, err := leb128.ReadVarUint32(r)
		if err != nil {
			return s, err
		}
		s.Elems = append(s.Elems, e)
	}

	return s, nil
}

// SectionCode describes the body for every function declared inside a module.
type SectionCode struct {
	Section
	Bodies []FunctionBody
}

func (m *Module) readSectionCode(r io.Reader) error {
	s := &SectionCode{}

	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}
	s.Bodies = make([]FunctionBody, count)
	logger.Printf("%d function bodies\n", count)

	for i := uint32(0); i < count; i++ {
		logger.Printf("Reading function %d\n", i)
		if s.Bodies[i], err = readFunctionBody(r); err != nil {
			return err
		}
		s.Bodies[i].Module = m

	}

	m.Code = s
	if m.Function == nil || len(m.Function.Types) == 0 {
		return errors.New("Missing section: function")
	}
	if len(m.Function.Types) != len(s.Bodies) {
		return errors.New("The number of entries in the function and code section are unequal")
	}

	if m.Types == nil {
		return errors.New("Missing section: type")
	}

	for _, index := range m.Function.Types {
		if int(index) >= len(m.Types.Entries) {
			return errors.New("Invalid index")
		}

		m.Types.Entries[index].body = &s.Bodies[index]
	}

	return nil
}

type FunctionBody struct {
	Module *Module // The parent module containing this function body, for execution purposes
	Locals []LocalEntry
	Code   []byte
}

func readFunctionBody(r io.Reader) (FunctionBody, error) {
	f := FunctionBody{}

	bodySize, err := leb128.ReadVarUint32(r)
	if err != nil {
		return f, err
	}

	r = io.LimitReader(r, int64(bodySize))

	localCount, err := leb128.ReadVarUint32(r)
	if err != nil {
		return f, err
	}
	f.Locals = make([]LocalEntry, localCount)

	for i := uint32(0); i < localCount; i++ {
		if f.Locals[i], err = readLocalEntry(r); err != nil {
			return f, err
		}
	}

	logger.Printf("bodySize: %d, localCount: %d\n", bodySize, localCount)

	code, err := ioutil.ReadAll(r)

	if code[len(code)-1] != end {
		return f, errors.New("Function body does not end with 0x0b (end)")
	}

	f.Code = code[:len(code)-1]
	if err != nil {
		return f, err
	}

	return f, nil
}

type LocalEntry struct {
	Count uint32
	Type  ValueType
}

func readLocalEntry(r io.Reader) (LocalEntry, error) {
	l := LocalEntry{}
	var err error

	l.Count, err = leb128.ReadVarUint32(r)
	if err != nil {
		return l, err
	}

	l.Type, err = readValueType(r)
	if err != nil {
		return l, err
	}

	return l, nil
}

// SectionData describes the intial values of a module's linear memory
type SectionData struct {
	Section
	Entries []DataSegment
}

func (m *Module) readSectionData(r io.Reader) error {
	s := &SectionData{}
	count, err := leb128.ReadVarUint32(r)
	if err != nil {
		return err
	}

	s.Entries = make([]DataSegment, count)

	for i := uint32(0); i < count; i++ {
		if s.Entries[i], err = readDataSegment(r); err != nil {
			return err
		}
	}

	m.Data = s
	return err
}

type DataSegment struct {
	Index uint32
	// should return i32, computes the offset at which to place the elements
	Offset []byte
	Data   []byte
}

func readDataSegment(r io.Reader) (DataSegment, error) {
	s := DataSegment{}
	var err error

	if s.Index, err = leb128.ReadVarUint32(r); err != nil {
		return s, err
	}
	if s.Offset, err = readInitExpr(r); err != nil {
		return s, err
	}

	size, err := leb128.ReadVarUint32(r)
	if err != nil {
		return s, err
	}
	s.Data, err = readBytes(r, uint(size))

	return s, err
}
