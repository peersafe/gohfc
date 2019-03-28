// Code generated by protoc-gen-go. DO NOT EDIT.
// source: common/policies.proto

package common // import "github.com/hyperledger/fabric/protos/common"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import msp "github.com/hyperledger/fabric/protos/msp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Policy_PolicyType int32

const (
	Policy_UNKNOWN       Policy_PolicyType = 0
	Policy_SIGNATURE     Policy_PolicyType = 1
	Policy_MSP           Policy_PolicyType = 2
	Policy_IMPLICIT_META Policy_PolicyType = 3
	Policy_ROLE          Policy_PolicyType = 4
)

var Policy_PolicyType_name = map[int32]string{
	0: "UNKNOWN",
	1: "SIGNATURE",
	2: "MSP",
	3: "IMPLICIT_META",
	4: "ROLE",
}
var Policy_PolicyType_value = map[string]int32{
	"UNKNOWN":       0,
	"SIGNATURE":     1,
	"MSP":           2,
	"IMPLICIT_META": 3,
	"ROLE":          4,
}

func (x Policy_PolicyType) String() string {
	return proto.EnumName(Policy_PolicyType_name, int32(x))
}
func (Policy_PolicyType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{0, 0}
}

type ImplicitMetaPolicy_Rule int32

const (
	ImplicitMetaPolicy_ANY      ImplicitMetaPolicy_Rule = 0
	ImplicitMetaPolicy_ALL      ImplicitMetaPolicy_Rule = 1
	ImplicitMetaPolicy_MAJORITY ImplicitMetaPolicy_Rule = 2
)

var ImplicitMetaPolicy_Rule_name = map[int32]string{
	0: "ANY",
	1: "ALL",
	2: "MAJORITY",
}
var ImplicitMetaPolicy_Rule_value = map[string]int32{
	"ANY":      0,
	"ALL":      1,
	"MAJORITY": 2,
}

func (x ImplicitMetaPolicy_Rule) String() string {
	return proto.EnumName(ImplicitMetaPolicy_Rule_name, int32(x))
}
func (ImplicitMetaPolicy_Rule) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{3, 0}
}

// Policy expresses a policy which the orderer can evaluate, because there has been some desire expressed to support
// multiple policy engines, this is typed as a oneof for now
type Policy struct {
	Type                 int32    `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"`
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Policy) Reset()         { *m = Policy{} }
func (m *Policy) String() string { return proto.CompactTextString(m) }
func (*Policy) ProtoMessage()    {}
func (*Policy) Descriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{0}
}
func (m *Policy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Policy.Unmarshal(m, b)
}
func (m *Policy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Policy.Marshal(b, m, deterministic)
}
func (dst *Policy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Policy.Merge(dst, src)
}
func (m *Policy) XXX_Size() int {
	return xxx_messageInfo_Policy.Size(m)
}
func (m *Policy) XXX_DiscardUnknown() {
	xxx_messageInfo_Policy.DiscardUnknown(m)
}

var xxx_messageInfo_Policy proto.InternalMessageInfo

func (m *Policy) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *Policy) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

// SignaturePolicyEnvelope wraps a SignaturePolicy and includes a version for future enhancements
type SignaturePolicyEnvelope struct {
	Version              int32               `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Rule                 *SignaturePolicy    `protobuf:"bytes,2,opt,name=rule,proto3" json:"rule,omitempty"`
	Identities           []*msp.MSPPrincipal `protobuf:"bytes,3,rep,name=identities,proto3" json:"identities,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *SignaturePolicyEnvelope) Reset()         { *m = SignaturePolicyEnvelope{} }
func (m *SignaturePolicyEnvelope) String() string { return proto.CompactTextString(m) }
func (*SignaturePolicyEnvelope) ProtoMessage()    {}
func (*SignaturePolicyEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{1}
}
func (m *SignaturePolicyEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignaturePolicyEnvelope.Unmarshal(m, b)
}
func (m *SignaturePolicyEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignaturePolicyEnvelope.Marshal(b, m, deterministic)
}
func (dst *SignaturePolicyEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignaturePolicyEnvelope.Merge(dst, src)
}
func (m *SignaturePolicyEnvelope) XXX_Size() int {
	return xxx_messageInfo_SignaturePolicyEnvelope.Size(m)
}
func (m *SignaturePolicyEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_SignaturePolicyEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_SignaturePolicyEnvelope proto.InternalMessageInfo

func (m *SignaturePolicyEnvelope) GetVersion() int32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *SignaturePolicyEnvelope) GetRule() *SignaturePolicy {
	if m != nil {
		return m.Rule
	}
	return nil
}

func (m *SignaturePolicyEnvelope) GetIdentities() []*msp.MSPPrincipal {
	if m != nil {
		return m.Identities
	}
	return nil
}

// SignaturePolicy is a recursive message structure which defines a featherweight DSL for describing
// policies which are more complicated than 'exactly this signature'.  The NOutOf operator is sufficent
// to express AND as well as OR, as well as of course N out of the following M policies
// SignedBy implies that the signature is from a valid certificate which is signed by the trusted
// authority specified in the bytes.  This will be the certificate itself for a self-signed certificate
// and will be the CA for more traditional certificates
type SignaturePolicy struct {
	// Types that are valid to be assigned to Type:
	//	*SignaturePolicy_SignedBy
	//	*SignaturePolicy_NOutOf_
	Type                 isSignaturePolicy_Type `protobuf_oneof:"Type"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *SignaturePolicy) Reset()         { *m = SignaturePolicy{} }
func (m *SignaturePolicy) String() string { return proto.CompactTextString(m) }
func (*SignaturePolicy) ProtoMessage()    {}
func (*SignaturePolicy) Descriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{2}
}
func (m *SignaturePolicy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignaturePolicy.Unmarshal(m, b)
}
func (m *SignaturePolicy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignaturePolicy.Marshal(b, m, deterministic)
}
func (dst *SignaturePolicy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignaturePolicy.Merge(dst, src)
}
func (m *SignaturePolicy) XXX_Size() int {
	return xxx_messageInfo_SignaturePolicy.Size(m)
}
func (m *SignaturePolicy) XXX_DiscardUnknown() {
	xxx_messageInfo_SignaturePolicy.DiscardUnknown(m)
}

var xxx_messageInfo_SignaturePolicy proto.InternalMessageInfo

type isSignaturePolicy_Type interface {
	isSignaturePolicy_Type()
}

type SignaturePolicy_SignedBy struct {
	SignedBy int32 `protobuf:"varint,1,opt,name=signed_by,json=signedBy,proto3,oneof"`
}

type SignaturePolicy_NOutOf_ struct {
	NOutOf *SignaturePolicy_NOutOf `protobuf:"bytes,2,opt,name=n_out_of,json=nOutOf,proto3,oneof"`
}

func (*SignaturePolicy_SignedBy) isSignaturePolicy_Type() {}

func (*SignaturePolicy_NOutOf_) isSignaturePolicy_Type() {}

func (m *SignaturePolicy) GetType() isSignaturePolicy_Type {
	if m != nil {
		return m.Type
	}
	return nil
}

func (m *SignaturePolicy) GetSignedBy() int32 {
	if x, ok := m.GetType().(*SignaturePolicy_SignedBy); ok {
		return x.SignedBy
	}
	return 0
}

func (m *SignaturePolicy) GetNOutOf() *SignaturePolicy_NOutOf {
	if x, ok := m.GetType().(*SignaturePolicy_NOutOf_); ok {
		return x.NOutOf
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*SignaturePolicy) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _SignaturePolicy_OneofMarshaler, _SignaturePolicy_OneofUnmarshaler, _SignaturePolicy_OneofSizer, []interface{}{
		(*SignaturePolicy_SignedBy)(nil),
		(*SignaturePolicy_NOutOf_)(nil),
	}
}

func _SignaturePolicy_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*SignaturePolicy)
	// Type
	switch x := m.Type.(type) {
	case *SignaturePolicy_SignedBy:
		b.EncodeVarint(1<<3 | proto.WireVarint)
		b.EncodeVarint(uint64(x.SignedBy))
	case *SignaturePolicy_NOutOf_:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.NOutOf); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("SignaturePolicy.Type has unexpected type %T", x)
	}
	return nil
}

func _SignaturePolicy_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*SignaturePolicy)
	switch tag {
	case 1: // Type.signed_by
		if wire != proto.WireVarint {
			return true, proto.ErrInternalBadWireType
		}
		x, err := b.DecodeVarint()
		m.Type = &SignaturePolicy_SignedBy{int32(x)}
		return true, err
	case 2: // Type.n_out_of
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(SignaturePolicy_NOutOf)
		err := b.DecodeMessage(msg)
		m.Type = &SignaturePolicy_NOutOf_{msg}
		return true, err
	default:
		return false, nil
	}
}

func _SignaturePolicy_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*SignaturePolicy)
	// Type
	switch x := m.Type.(type) {
	case *SignaturePolicy_SignedBy:
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(x.SignedBy))
	case *SignaturePolicy_NOutOf_:
		s := proto.Size(x.NOutOf)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type SignaturePolicy_NOutOf struct {
	N                    int32              `protobuf:"varint,1,opt,name=n,proto3" json:"n,omitempty"`
	Rules                []*SignaturePolicy `protobuf:"bytes,2,rep,name=rules,proto3" json:"rules,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *SignaturePolicy_NOutOf) Reset()         { *m = SignaturePolicy_NOutOf{} }
func (m *SignaturePolicy_NOutOf) String() string { return proto.CompactTextString(m) }
func (*SignaturePolicy_NOutOf) ProtoMessage()    {}
func (*SignaturePolicy_NOutOf) Descriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{2, 0}
}
func (m *SignaturePolicy_NOutOf) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SignaturePolicy_NOutOf.Unmarshal(m, b)
}
func (m *SignaturePolicy_NOutOf) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SignaturePolicy_NOutOf.Marshal(b, m, deterministic)
}
func (dst *SignaturePolicy_NOutOf) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SignaturePolicy_NOutOf.Merge(dst, src)
}
func (m *SignaturePolicy_NOutOf) XXX_Size() int {
	return xxx_messageInfo_SignaturePolicy_NOutOf.Size(m)
}
func (m *SignaturePolicy_NOutOf) XXX_DiscardUnknown() {
	xxx_messageInfo_SignaturePolicy_NOutOf.DiscardUnknown(m)
}

var xxx_messageInfo_SignaturePolicy_NOutOf proto.InternalMessageInfo

func (m *SignaturePolicy_NOutOf) GetN() int32 {
	if m != nil {
		return m.N
	}
	return 0
}

func (m *SignaturePolicy_NOutOf) GetRules() []*SignaturePolicy {
	if m != nil {
		return m.Rules
	}
	return nil
}

// ImplicitMetaPolicy is a policy type which depends on the hierarchical nature of the configuration
// It is implicit because the rule is generate implicitly based on the number of sub policies
// It is meta because it depends only on the result of other policies
// When evaluated, this policy iterates over all immediate child sub-groups, retrieves the policy
// of name sub_policy, evaluates the collection and applies the rule.
// For example, with 4 sub-groups, and a policy name of "foo", ImplicitMetaPolicy retrieves
// each sub-group, retrieves policy "foo" for each subgroup, evaluates it, and, in the case of ANY
// 1 satisfied is sufficient, ALL would require 4 signatures, and MAJORITY would require 3 signatures.
type ImplicitMetaPolicy struct {
	SubPolicy            string                  `protobuf:"bytes,1,opt,name=sub_policy,json=subPolicy,proto3" json:"sub_policy,omitempty"`
	Rule                 ImplicitMetaPolicy_Rule `protobuf:"varint,2,opt,name=rule,proto3,enum=common.ImplicitMetaPolicy_Rule" json:"rule,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *ImplicitMetaPolicy) Reset()         { *m = ImplicitMetaPolicy{} }
func (m *ImplicitMetaPolicy) String() string { return proto.CompactTextString(m) }
func (*ImplicitMetaPolicy) ProtoMessage()    {}
func (*ImplicitMetaPolicy) Descriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{3}
}
func (m *ImplicitMetaPolicy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ImplicitMetaPolicy.Unmarshal(m, b)
}
func (m *ImplicitMetaPolicy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ImplicitMetaPolicy.Marshal(b, m, deterministic)
}
func (dst *ImplicitMetaPolicy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ImplicitMetaPolicy.Merge(dst, src)
}
func (m *ImplicitMetaPolicy) XXX_Size() int {
	return xxx_messageInfo_ImplicitMetaPolicy.Size(m)
}
func (m *ImplicitMetaPolicy) XXX_DiscardUnknown() {
	xxx_messageInfo_ImplicitMetaPolicy.DiscardUnknown(m)
}

var xxx_messageInfo_ImplicitMetaPolicy proto.InternalMessageInfo

func (m *ImplicitMetaPolicy) GetSubPolicy() string {
	if m != nil {
		return m.SubPolicy
	}
	return ""
}

func (m *ImplicitMetaPolicy) GetRule() ImplicitMetaPolicy_Rule {
	if m != nil {
		return m.Rule
	}
	return ImplicitMetaPolicy_ANY
}

// RolePolicy is a policy type which depends on the MSPOrgRole in the channel, every MSP has a role in core, main
// or normal and role defined in MSPCONFIG. Percent is a integer between 0 and 100 which defines how many Signatures
// of MSP-with-right-roles  defined in policy is valid.
// 0 needs no signature
// 100 needs all signature
// 1~99 needs ceil(percent/100*MSPs)
type RolePolicy struct {
	Percent              int32          `protobuf:"varint,1,opt,name=percent,proto3" json:"percent,omitempty"`
	Role                 msp.MSPOrgRole `protobuf:"varint,2,opt,name=role,proto3,enum=msp.MSPOrgRole" json:"role,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *RolePolicy) Reset()         { *m = RolePolicy{} }
func (m *RolePolicy) String() string { return proto.CompactTextString(m) }
func (*RolePolicy) ProtoMessage()    {}
func (*RolePolicy) Descriptor() ([]byte, []int) {
	return fileDescriptor_policies_979e39cfc9446621, []int{4}
}
func (m *RolePolicy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RolePolicy.Unmarshal(m, b)
}
func (m *RolePolicy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RolePolicy.Marshal(b, m, deterministic)
}
func (dst *RolePolicy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RolePolicy.Merge(dst, src)
}
func (m *RolePolicy) XXX_Size() int {
	return xxx_messageInfo_RolePolicy.Size(m)
}
func (m *RolePolicy) XXX_DiscardUnknown() {
	xxx_messageInfo_RolePolicy.DiscardUnknown(m)
}

var xxx_messageInfo_RolePolicy proto.InternalMessageInfo

func (m *RolePolicy) GetPercent() int32 {
	if m != nil {
		return m.Percent
	}
	return 0
}

func (m *RolePolicy) GetRole() msp.MSPOrgRole {
	if m != nil {
		return m.Role
	}
	return msp.MSPOrgRole_NORMAL
}

func init() {
	proto.RegisterType((*Policy)(nil), "common.Policy")
	proto.RegisterType((*SignaturePolicyEnvelope)(nil), "common.SignaturePolicyEnvelope")
	proto.RegisterType((*SignaturePolicy)(nil), "common.SignaturePolicy")
	proto.RegisterType((*SignaturePolicy_NOutOf)(nil), "common.SignaturePolicy.NOutOf")
	proto.RegisterType((*ImplicitMetaPolicy)(nil), "common.ImplicitMetaPolicy")
	proto.RegisterType((*RolePolicy)(nil), "common.RolePolicy")
	proto.RegisterEnum("common.Policy_PolicyType", Policy_PolicyType_name, Policy_PolicyType_value)
	proto.RegisterEnum("common.ImplicitMetaPolicy_Rule", ImplicitMetaPolicy_Rule_name, ImplicitMetaPolicy_Rule_value)
}

func init() { proto.RegisterFile("common/policies.proto", fileDescriptor_policies_979e39cfc9446621) }

var fileDescriptor_policies_979e39cfc9446621 = []byte{
	// 537 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x52, 0x41, 0x6b, 0xdb, 0x4c,
	0x10, 0xb5, 0x6c, 0x45, 0xb6, 0x27, 0xc9, 0x17, 0x7d, 0x4b, 0x4a, 0x44, 0xa0, 0x6d, 0x50, 0x4b,
	0x09, 0x84, 0x4a, 0x90, 0xf4, 0xd4, 0x9b, 0x5d, 0x4c, 0xa3, 0xc6, 0x92, 0xcc, 0xda, 0xa1, 0xa4,
	0x17, 0x61, 0xc9, 0x6b, 0x65, 0x41, 0xda, 0x5d, 0x56, 0x92, 0xc1, 0xd7, 0xfe, 0x82, 0x9e, 0xfa,
	0x67, 0xfa, 0xe7, 0x8a, 0xb4, 0x56, 0x30, 0x29, 0xb9, 0xed, 0x9b, 0x79, 0x33, 0xfb, 0xe6, 0xcd,
	0xc0, 0xab, 0x84, 0xe7, 0x39, 0x67, 0xae, 0xe0, 0x19, 0x4d, 0x28, 0x29, 0x1c, 0x21, 0x79, 0xc9,
	0x91, 0xa1, 0xc2, 0xe7, 0x67, 0x79, 0x21, 0xdc, 0xbc, 0x10, 0x91, 0x90, 0x94, 0x25, 0x54, 0x2c,
	0x33, 0x45, 0x38, 0x3f, 0x6d, 0x13, 0x09, 0x67, 0x6b, 0x9a, 0xaa, 0xa8, 0xfd, 0x53, 0x03, 0x63,
	0x56, 0x77, 0xda, 0x22, 0x04, 0x7a, 0xb9, 0x15, 0xc4, 0xd2, 0x2e, 0xb4, 0xcb, 0x03, 0xdc, 0xbc,
	0xd1, 0x29, 0x1c, 0x6c, 0x96, 0x59, 0x45, 0xac, 0xee, 0x85, 0x76, 0x79, 0x84, 0x15, 0xb0, 0x03,
	0x00, 0x55, 0xb3, 0xa8, 0x39, 0x87, 0xd0, 0xbf, 0x0f, 0xee, 0x82, 0xf0, 0x7b, 0x60, 0x76, 0xd0,
	0x31, 0x0c, 0xe7, 0xde, 0xd7, 0x60, 0xb4, 0xb8, 0xc7, 0x13, 0x53, 0x43, 0x7d, 0xe8, 0xf9, 0xf3,
	0x99, 0xd9, 0x45, 0xff, 0xc3, 0xb1, 0xe7, 0xcf, 0xa6, 0xde, 0x17, 0x6f, 0x11, 0xf9, 0x93, 0xc5,
	0xc8, 0xec, 0xa1, 0x01, 0xe8, 0x38, 0x9c, 0x4e, 0x4c, 0xdd, 0xfe, 0xad, 0xc1, 0xd9, 0x9c, 0xa6,
	0x6c, 0x59, 0x56, 0x92, 0xa8, 0xce, 0x13, 0xb6, 0x21, 0x19, 0x17, 0x04, 0x59, 0xd0, 0xdf, 0x10,
	0x59, 0x50, 0xce, 0x76, 0xc2, 0x5a, 0x88, 0xae, 0x40, 0x97, 0x55, 0xa6, 0xa4, 0x1d, 0x5e, 0x9f,
	0x39, 0xca, 0x00, 0xe7, 0x59, 0x23, 0xdc, 0x90, 0xd0, 0x27, 0x00, 0xba, 0x22, 0xac, 0xa4, 0x25,
	0x25, 0x85, 0xd5, 0xbb, 0xe8, 0x5d, 0x1e, 0x5e, 0x9f, 0xb6, 0x25, 0xfe, 0x7c, 0x36, 0x6b, 0xdd,
	0xc2, 0x7b, 0x3c, 0xfb, 0x8f, 0x06, 0x27, 0xcf, 0xfa, 0xa1, 0xd7, 0x30, 0x2c, 0x68, 0xca, 0xc8,
	0x2a, 0x8a, 0xb7, 0x4a, 0xd2, 0x6d, 0x07, 0x0f, 0x54, 0x68, 0xbc, 0x45, 0x9f, 0x61, 0xc0, 0x22,
	0x5e, 0x95, 0x11, 0x5f, 0xef, 0x94, 0xbd, 0x79, 0x41, 0x99, 0x13, 0x84, 0x55, 0x19, 0xae, 0x6f,
	0x3b, 0xd8, 0x60, 0xcd, 0xeb, 0x7c, 0x02, 0x86, 0x8a, 0xa1, 0x23, 0xd0, 0xda, 0x79, 0x35, 0x86,
	0x3e, 0xc2, 0x41, 0x3d, 0x44, 0x61, 0x75, 0x1b, 0xdd, 0x2f, 0x8e, 0xaa, 0x58, 0x63, 0x03, 0xf4,
	0x7a, 0x31, 0xf6, 0x2f, 0x0d, 0x90, 0x97, 0x8b, 0xfa, 0x4c, 0x4a, 0x9f, 0x94, 0xcb, 0xa7, 0x01,
	0xa0, 0xa8, 0xe2, 0xa8, 0xb9, 0x1f, 0x35, 0xc1, 0x10, 0x0f, 0x8b, 0x2a, 0xde, 0xa5, 0x6f, 0xf6,
	0x6c, 0xfd, 0xef, 0xfa, 0x6d, 0xfb, 0xd7, 0xbf, 0x8d, 0x1c, 0x5c, 0x65, 0x44, 0xd9, 0x6b, 0x7f,
	0x00, 0xbd, 0x46, 0xf5, 0xbe, 0x47, 0xc1, 0x83, 0xd9, 0x69, 0x1e, 0xd3, 0xa9, 0xa9, 0xa1, 0x23,
	0x18, 0xf8, 0xa3, 0x6f, 0x21, 0xf6, 0x16, 0x0f, 0x66, 0xd7, 0xbe, 0x03, 0xc0, 0x3c, 0x6b, 0xad,
	0xb4, 0xa0, 0x2f, 0x88, 0x4c, 0x08, 0x2b, 0xdb, 0xdd, 0xee, 0x20, 0x7a, 0x07, 0xba, 0xe4, 0x4f,
	0x22, 0x4e, 0x9c, 0xbc, 0x10, 0xf5, 0x96, 0x42, 0x99, 0xd6, 0xe5, 0xb8, 0x49, 0x8e, 0xe7, 0xf0,
	0x9e, 0xcb, 0xd4, 0x79, 0xdc, 0x0a, 0x22, 0x33, 0xb2, 0x4a, 0x89, 0x74, 0xd6, 0xcb, 0x58, 0xd2,
	0x44, 0xdd, 0x76, 0xb1, 0x93, 0xfe, 0xe3, 0x2a, 0xa5, 0xe5, 0x63, 0x15, 0xd7, 0xd0, 0xdd, 0x23,
	0xbb, 0x8a, 0xec, 0x2a, 0xb2, 0xab, 0xc8, 0xb1, 0xd1, 0xc0, 0x9b, 0xbf, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x50, 0xa6, 0xa8, 0x15, 0x67, 0x03, 0x00, 0x00,
}
