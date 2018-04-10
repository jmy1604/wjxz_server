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
	MSGID_S2C_TEAMS_RESPONSE             MSGID = 10104
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
	10104: "S2C_TEAMS_RESPONSE",
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
	"S2C_TEAMS_RESPONSE":             10104,
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
	// 354 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x64, 0x92, 0xcb, 0x4e, 0x2a, 0x41,
	0x10, 0x86, 0xcf, 0x85, 0x73, 0x72, 0xd2, 0xc9, 0x49, 0x9a, 0x0e, 0x97, 0x80, 0x06, 0x17, 0xee,
	0x5c, 0xb0, 0xc0, 0x27, 0x68, 0x86, 0x62, 0x98, 0x64, 0xa6, 0x7a, 0x9c, 0xae, 0x21, 0x71, 0xd5,
	0xf1, 0x42, 0x08, 0x89, 0x88, 0x11, 0x9e, 0x40, 0x97, 0x6e, 0xbc, 0x3d, 0xaa, 0x17, 0x74, 0x65,
	0x9a, 0x99, 0x69, 0x11, 0xb7, 0xf5, 0xd7, 0xf7, 0x4d, 0x55, 0x4d, 0xb3, 0xfa, 0xc9, 0xd9, 0x64,
	0x74, 0xbe, 0x30, 0xd3, 0xd1, 0x7c, 0x7e, 0x34, 0x1e, 0x99, 0xc9, 0x69, 0xfb, 0xe2, 0x72, 0xb6,
	0x98, 0x89, 0xea, 0x74, 0x3e, 0x6e, 0x7f, 0x0b, 0xf7, 0x6e, 0x4a, 0xec, 0x4f, 0xa4, 0xfd, 0xa0,
	0x27, 0xfe, 0xb1, 0x12, 0x2a, 0x04, 0xfe, 0x43, 0x54, 0x18, 0xf7, 0x3a, 0xda, 0x10, 0x68, 0x32,
	0x9e, 0x8a, 0x22, 0x89, 0x3d, 0xfe, 0x53, 0x94, 0xd9, 0x7f, 0x5b, 0x1d, 0x80, 0x4c, 0xa8, 0x0b,
	0x92, 0xf8, 0x2f, 0xd1, 0x60, 0x55, 0xdd, 0xf1, 0x4c, 0x40, 0x10, 0x69, 0x13, 0x60, 0x5f, 0x99,
	0x34, 0xee, 0x49, 0x02, 0xfe, 0x5b, 0xd4, 0x58, 0xd9, 0x76, 0x87, 0xca, 0x0f, 0xd0, 0x24, 0x70,
	0x90, 0x82, 0x26, 0x7e, 0x8b, 0xa2, 0xce, 0x84, 0x45, 0x8a, 0xba, 0x8e, 0x15, 0x6a, 0xe0, 0x77,
	0x28, 0x9a, 0x99, 0x4b, 0xd1, 0x00, 0x12, 0x13, 0x87, 0xd2, 0x83, 0xac, 0x89, 0xdf, 0xa3, 0x68,
	0xb1, 0x86, 0x95, 0x69, 0x08, 0xc1, 0x23, 0xa3, 0x21, 0x19, 0x42, 0xe2, 0xa4, 0x0f, 0x28, 0x76,
	0x58, 0xd3, 0xb2, 0x9b, 0x79, 0x2e, 0x7f, 0x44, 0xb1, 0xc5, 0x6a, 0x56, 0x00, 0x48, 0x90, 0x18,
	0x5f, 0x46, 0xe0, 0xe8, 0x65, 0x2a, 0xb6, 0x59, 0xdd, 0xd2, 0x5f, 0xc2, 0x1c, 0x7d, 0x4b, 0xc5,
	0x2e, 0x6b, 0x6d, 0xa4, 0x9e, 0x8a, 0xe2, 0x10, 0x08, 0x0c, 0x2a, 0x0a, 0xfa, 0x87, 0xfc, 0x3d,
	0x2d, 0xb6, 0x4a, 0x54, 0x08, 0xfa, 0x93, 0xbe, 0x1a, 0x16, 0xee, 0x2c, 0xf0, 0x06, 0x12, 0x7d,
	0x87, 0x5d, 0x0f, 0x8b, 0xbd, 0xba, 0x92, 0x28, 0x5c, 0x7d, 0x35, 0x0d, 0xc9, 0x4d, 0xf6, 0xe4,
	0xf6, 0xda, 0xcc, 0x73, 0xfd, 0x33, 0x8a, 0x06, 0xab, 0x64, 0x87, 0x21, 0x43, 0x20, 0x23, 0xc7,
	0xbe, 0xb8, 0x7b, 0xae, 0x45, 0x39, 0xf6, 0xea, 0x7e, 0x82, 0xad, 0xaf, 0x8d, 0xbb, 0xc4, 0xe3,
	0xbf, 0xab, 0xb7, 0xb2, 0xff, 0x11, 0x00, 0x00, 0xff, 0xff, 0xc5, 0x1b, 0x32, 0x65, 0x46, 0x02,
	0x00, 0x00,
}
