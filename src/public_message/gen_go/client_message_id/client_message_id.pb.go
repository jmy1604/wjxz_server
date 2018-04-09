// Code generated by protoc-gen-go. DO NOT EDIT.
// source: client_message_id.proto

/*
Package msg_client_message_id is a generated protocol buffer package.

It is generated from these files:
	client_message_id.proto

It has these top-level messages:
*/
package msg_client_message_id

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type MSGID int32

const (
	MSGID_NONE                           MSGID = 0
	MSGID_C2S_TEST_COMMAND               MSGID = 1
	MSGID_C2S_HEARTBEAT                  MSGID = 2
	MSGID_S2C_ITEMS_INFO_UPDATE          MSGID = 3
	MSGID_C2S_LOGIN_REQUEST              MSGID = 10000
	MSGID_S2C_LOGIN_RESPONSE             MSGID = 10001
	MSGID_S2C_OTHER_PLACE_LOGIN          MSGID = 10002
	MSGID_C2S_SELECT_SERVER_REQUEST      MSGID = 10003
	MSGID_S2C_SELECT_SERVER_RESPONSE     MSGID = 10004
	MSGID_C2S_ENTER_GAME_REQUEST         MSGID = 11000
	MSGID_S2C_ENTER_GAME_RESPONSE        MSGID = 11001
	MSGID_S2C_ENTER_GAME_COMPLETE_NOTIFY MSGID = 11002
	MSGID_S2C_ROLES_RESPONSE             MSGID = 11010
	MSGID_S2C_ROLES_CHANGE_NOTIFY        MSGID = 11011
	MSGID_C2S_BATTLE_RESULT_REQUEST      MSGID = 10100
	MSGID_S2C_BATTLE_RESULT_RESPONSE     MSGID = 10101
	MSGID_C2S_SET_TEAM_REQUEST           MSGID = 10102
	MSGID_S2C_SET_TEAM_RESPONSE          MSGID = 10103
)

var MSGID_name = map[int32]string{
	0:     "NONE",
	1:     "C2S_TEST_COMMAND",
	2:     "C2S_HEARTBEAT",
	3:     "S2C_ITEMS_INFO_UPDATE",
	10000: "C2S_LOGIN_REQUEST",
	10001: "S2C_LOGIN_RESPONSE",
	10002: "S2C_OTHER_PLACE_LOGIN",
	10003: "C2S_SELECT_SERVER_REQUEST",
	10004: "S2C_SELECT_SERVER_RESPONSE",
	11000: "C2S_ENTER_GAME_REQUEST",
	11001: "S2C_ENTER_GAME_RESPONSE",
	11002: "S2C_ENTER_GAME_COMPLETE_NOTIFY",
	11010: "S2C_ROLES_RESPONSE",
	11011: "S2C_ROLES_CHANGE_NOTIFY",
	10100: "C2S_BATTLE_RESULT_REQUEST",
	10101: "S2C_BATTLE_RESULT_RESPONSE",
	10102: "C2S_SET_TEAM_REQUEST",
	10103: "S2C_SET_TEAM_RESPONSE",
}
var MSGID_value = map[string]int32{
	"NONE":                           0,
	"C2S_TEST_COMMAND":               1,
	"C2S_HEARTBEAT":                  2,
	"S2C_ITEMS_INFO_UPDATE":          3,
	"C2S_LOGIN_REQUEST":              10000,
	"S2C_LOGIN_RESPONSE":             10001,
	"S2C_OTHER_PLACE_LOGIN":          10002,
	"C2S_SELECT_SERVER_REQUEST":      10003,
	"S2C_SELECT_SERVER_RESPONSE":     10004,
	"C2S_ENTER_GAME_REQUEST":         11000,
	"S2C_ENTER_GAME_RESPONSE":        11001,
	"S2C_ENTER_GAME_COMPLETE_NOTIFY": 11002,
	"S2C_ROLES_RESPONSE":             11010,
	"S2C_ROLES_CHANGE_NOTIFY":        11011,
	"C2S_BATTLE_RESULT_REQUEST":      10100,
	"S2C_BATTLE_RESULT_RESPONSE":     10101,
	"C2S_SET_TEAM_REQUEST":           10102,
	"S2C_SET_TEAM_RESPONSE":          10103,
}

func (x MSGID) String() string {
	return proto.EnumName(MSGID_name, int32(x))
}
func (MSGID) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func init() {
	proto.RegisterEnum("msg.client_message_id.MSGID", MSGID_name, MSGID_value)
}

func init() { proto.RegisterFile("client_message_id.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 345 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x64, 0x92, 0xcb, 0x4e, 0x2a, 0x41,
	0x10, 0x40, 0xef, 0xbd, 0x5c, 0x8d, 0xe9, 0xc4, 0xa4, 0xa9, 0xf0, 0x08, 0x68, 0x70, 0xe1, 0xce,
	0x05, 0x0b, 0xfc, 0x82, 0x66, 0x28, 0x86, 0x49, 0x66, 0xaa, 0xc7, 0xe9, 0x1a, 0x12, 0x57, 0x1d,
	0x1f, 0x84, 0x90, 0x88, 0x18, 0xe1, 0x0b, 0xf4, 0x07, 0x7c, 0x7d, 0xa9, 0xf1, 0xbd, 0x32, 0xc3,
	0x3c, 0x54, 0xdc, 0xf6, 0xe9, 0x73, 0x66, 0xaa, 0xd2, 0xa2, 0x7e, 0x72, 0x36, 0x19, 0x9d, 0x2f,
	0xec, 0x74, 0x34, 0x9f, 0x1f, 0x8d, 0x47, 0x76, 0x72, 0xda, 0xbe, 0xb8, 0x9c, 0x2d, 0x66, 0x50,
	0x9d, 0xce, 0xc7, 0xed, 0x5f, 0x70, 0xef, 0xb1, 0x24, 0xd6, 0x02, 0xe3, 0x7a, 0x3d, 0xd8, 0x10,
	0xff, 0x49, 0x13, 0xca, 0x3f, 0x50, 0x11, 0xd2, 0xe9, 0x18, 0xcb, 0x68, 0xd8, 0x3a, 0x3a, 0x08,
	0x14, 0xf5, 0xe4, 0x5f, 0x28, 0x8b, 0xcd, 0xe4, 0x74, 0x80, 0x2a, 0xe2, 0x2e, 0x2a, 0x96, 0xff,
	0xa0, 0x21, 0xaa, 0xa6, 0xe3, 0x58, 0x8f, 0x31, 0x30, 0xd6, 0xa3, 0xbe, 0xb6, 0x71, 0xd8, 0x53,
	0x8c, 0xb2, 0x04, 0x35, 0x51, 0x4e, 0x6e, 0xfb, 0xda, 0xf5, 0xc8, 0x46, 0x78, 0x10, 0xa3, 0x61,
	0x79, 0x43, 0x50, 0x17, 0x90, 0x28, 0xf9, 0xb9, 0x09, 0x35, 0x19, 0x94, 0xb7, 0x04, 0xcd, 0xb4,
	0xa5, 0x79, 0x80, 0x91, 0x0d, 0x7d, 0xe5, 0x60, 0x7a, 0x49, 0xde, 0x11, 0xb4, 0x44, 0x23, 0x89,
	0x19, 0xf4, 0xd1, 0x61, 0x6b, 0x30, 0x1a, 0x62, 0x54, 0x44, 0xef, 0x09, 0x76, 0x44, 0x33, 0x71,
	0x57, 0x79, 0x16, 0x7f, 0x20, 0xd8, 0x12, 0xb5, 0x24, 0x80, 0xc4, 0x18, 0x59, 0x57, 0x05, 0x58,
	0xd8, 0x6f, 0x31, 0x6c, 0x8b, 0x7a, 0x62, 0xff, 0x80, 0x99, 0xfa, 0x1e, 0xc3, 0xae, 0x68, 0xad,
	0x50, 0x47, 0x07, 0xa1, 0x8f, 0x8c, 0x96, 0x34, 0x7b, 0xfd, 0x43, 0xf9, 0x11, 0xe7, 0x53, 0x45,
	0xda, 0x47, 0xf3, 0x65, 0x5f, 0x0d, 0xf3, 0x76, 0x0a, 0x9c, 0x81, 0x22, 0xb7, 0xd0, 0xae, 0x87,
	0xf9, 0x5c, 0x5d, 0xc5, 0xec, 0x2f, 0xbf, 0x1a, 0xfb, 0x5c, 0xfc, 0xd9, 0x53, 0x31, 0xd7, 0x2a,
	0xcf, 0xf2, 0xcf, 0x04, 0x0d, 0x51, 0x49, 0x17, 0xc3, 0x96, 0x51, 0x05, 0x85, 0xfb, 0x52, 0xec,
	0xf3, 0x1b, 0xca, 0xb4, 0x57, 0x3a, 0x5e, 0x5f, 0x3e, 0x89, 0xfd, 0xcf, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x9a, 0xa5, 0xee, 0x05, 0x2d, 0x02, 0x00, 0x00,
}
