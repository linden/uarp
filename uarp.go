package uarp

import (
	"encoding/binary"
	"fmt"
)

type Metadata struct {
	Type  string
	Value int
}

type Row struct {
	Metadata []Metadata
	Version  string
	Payload  []byte
	Type     string
}

type Table struct {
	Metadata []Metadata
	Version  string
	Rows     []Row
}

type Reader struct {
	Value []byte
}

var MetadataTypes = map[uint32]string{
	3436347648: "Payload Filepath",
	3436347649: "Payload Long Name",
	3436347650: "Minimum Required Version",
	3436347651: "Ignore Version",
	3436347652: "Urgent Update",
	3436347653: "Payload Certificate",
	3436347654: "Payload Signature",
	3436347655: "Payload Hash",
	3436347656: "Payload Digest",
	3436347657: "Minimum Battery Level",
	3436347658: "Trigger Battery Level",
	3436347659: "Payload Compression ChunkSize",
	3436347660: "Payload Compression Algorithm",
	3436347663: "Compressed Headers Payload Index",
	1619725824: "HeySiri Model Type",
	1619725825: "HeySiri Model Locale",
	1619725826: "HeySiri Model Hash",
	1619725827: "HeySiri Model Role",
	1619725828: "HeySiri Model Digest",
	1619725829: "HeySiri Model Signature",
	1619725830: "HeySiri Model Certificate",
	1619725831: "HeySiri Model Engine Version",
	1619725832: "HeySiri Model Engine Type",
	2293403904: "Personalization Required",
	2293403905: "Personalization Payload Tag",
	2293403906: "Personalization SuperBinary AssetID",
	2293403907: "Personalization Manifest Prefix",
	3291140096: "Host Minimum Battery Level",
	3291140097: "Host Inactive To Stage Asset",
	3291140098: "Host Inactive To Apply Asset",
	3291140099: "Host Network Delay",
	3291140100: "Host Reconnect After Apply",
	3291140101: "Minimum iOS Version",
	3291140102: "Minimum macOS Version",
	3291140103: "Minimum tvOS Version",
	3291140104: "Minimum watchOS Version",
	3291140105: "Host Trigger Battery Level",
	76079616:   "Voice Assist Type",
	76079617:   "Voice Assist Locale",
	76079618:   "Voice Assist Hash",
	76079619:   "Voice Assist Role",
	76079620:   "Voice Assist Digest",
	76079621:   "Voice Assist Signature",
	76079622:   "Voice Assist Certificate",
	76079623:   "Voice Assist Engine Version",
}

var PayloadTypes = map[string]string{
	"FOTA": "Firmware Over the Air (FOTA)",
	"P1FW": "PROTO1 (P1FW)",
	"P2FW": "PROTO2 (P2FW)",
	"EVTF": "Engineering Validation Test (EVTF)",
	"PVTF": "Production Validation Test (PVTF)",
	"MPFW": "Mainline Production Firmware (MPFW)",
	"STFW": "Storage Firmware (STFW)",
	"DTTX": "Data Transmit (DTTX)",
	"DTRX": "Data Receive (DTRX)",
	"DMTP": "Test Point (DMTP)",
	"PDFW": "USB-C Power Delivery (PDFW)",
	"ULPD": "Upload (ULPD)",
	"CHDR": "Charge Direction (CHDR)",
}

func (reader *Reader) Read(length int) []byte {
	cursor := reader.Value[:length]
	reader.Value = reader.Value[length:]

	return cursor
}

func (reader *Reader) ReadVersion() string {
	major := binary.BigEndian.Uint32(reader.Read(4))
	minor := binary.BigEndian.Uint32(reader.Read(4))
	release := binary.BigEndian.Uint32(reader.Read(4))
	build := binary.BigEndian.Uint32(reader.Read(4))

	return fmt.Sprintf("%d.%d.%d.%d", major, minor, release, build)
}

func ParseMetadata(raw []byte) []Metadata {
	reader := Reader{raw}

	var metadata []Metadata

	for index := 0; len(reader.Value) > 10; index++ {
		var _type string

		if value, known := MetadataTypes[binary.BigEndian.Uint32(reader.Read(4))]; known {
			_type = value
		} else {
			_type = "Unknown Metadata Type"
		}

		length := binary.BigEndian.Uint32(reader.Read(4))

		var value int

		switch length {
		case 2:
			value = int(binary.BigEndian.Uint16(reader.Read(2)))

		case 4:
			value = int(binary.BigEndian.Uint32(reader.Read(4)))

		default:
			fmt.Println("invalid length", length)
		}

		metadata = append(metadata, Metadata{_type, value})
	}

	return metadata
}

func ParseRows(raw []byte, row []byte) []Row {
	reader := Reader{row}

	var rows []Row

	for index := 0; index < len(reader.Value); index++ {
		_ = binary.BigEndian.Uint32(reader.Read(4)) // size

		var _type string

		if value, known := PayloadTypes[string(reader.Read(4))]; known {
			_type = value
		} else {
			_type = "Unknown Payload Type"
		}

		version := reader.ReadVersion()

		metadataOffset := binary.BigEndian.Uint32(reader.Read(4))
		metadataLength := binary.BigEndian.Uint32(reader.Read(4))

		payloadOffset := binary.BigEndian.Uint32(reader.Read(4))
		payloadLength := binary.BigEndian.Uint32(reader.Read(4))

		metadata := ParseMetadata(raw[metadataOffset:(metadataOffset + metadataLength)])
		payload := raw[payloadOffset:(payloadOffset + payloadLength)]

		rows = append(rows, Row{
			Metadata: metadata,
			Version:  version,
			Payload:  payload,
			Type:     _type,
		})
	}

	return rows
}

func ParseTable(raw []byte) Table {
	reader := Reader{raw}

	_ = binary.BigEndian.Uint32(reader.Read(4)) // format
	_ = binary.BigEndian.Uint32(reader.Read(4)) // size
	_ = binary.BigEndian.Uint32(reader.Read(4)) // binarySize

	version := reader.ReadVersion()

	metadataOffset := binary.BigEndian.Uint32(reader.Read(4))
	metadataLength := binary.BigEndian.Uint32(reader.Read(4))

	rowOffset := binary.BigEndian.Uint32(reader.Read(4))
	rowLength := binary.BigEndian.Uint32(reader.Read(4))

	metadata := ParseMetadata(raw[metadataOffset:(metadataOffset + metadataLength)])
	rows := ParseRows(raw, raw[rowOffset:(rowOffset+rowLength)])

	return Table{
		Metadata: metadata,
		Version:  version,
		Rows:     rows,
	}
}
