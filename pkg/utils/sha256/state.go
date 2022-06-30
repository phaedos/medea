package sha256

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"hash"
	"reflect"
	"unsafe"
)

type State struct {
	H [8]uint32

	X [64]byte

	Nx int

	Len uint64
}

func (s *State) EncodeToString() (string, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(s); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func DecodeStringToState(cipherText string) (*State, error) {
	plainTextByte, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return nil, err
	}
	state := &State{}
	buf := bytes.Buffer{}
	buf.Write(plainTextByte)
	decoder := gob.NewDecoder(&buf)
	if err = decoder.Decode(state); err != nil {
		return nil, err
	}
	return state, nil
}

var ErrDigestType = errors.New("digest must be type of *sha256.digest")

func GetHashState(digest hash.Hash) (*State, error) {

	if reflect.TypeOf(digest).String() != "*sha256.digest" {
		return nil, ErrDigestType
	}

	digestElem := reflect.ValueOf(digest).Elem()

	var (
		h    [8]uint32
		x    [64]byte
		nx   int
		xLen uint64
	)

	rfh := digestElem.FieldByName("h")
	for i := 0; i < rfh.Len(); i++ {
		h[i] = *((*uint32)(unsafe.Pointer(rfh.Index(i).UnsafeAddr())))
	}

	rfx := digestElem.FieldByName("x")
	rfx = reflect.NewAt(rfx.Type(), unsafe.Pointer(rfx.UnsafeAddr())).Elem()
	x = rfx.Interface().([64]byte)

	rfnx := digestElem.FieldByName("nx")
	rfnx = reflect.NewAt(rfnx.Type(), unsafe.Pointer(rfnx.UnsafeAddr())).Elem()
	nx = rfnx.Interface().(int)

	rfxLen := digestElem.FieldByName("len")
	rfxLen = reflect.NewAt(rfxLen.Type(), unsafe.Pointer(rfxLen.UnsafeAddr())).Elem()
	xLen = rfxLen.Interface().(uint64)

	return &State{
		H:   h,
		X:   x,
		Nx:  nx,
		Len: xLen,
	}, nil
}

func SetHashState(digest hash.Hash, state *State) error {

	if reflect.TypeOf(digest).String() != "*sha256.digest" {
		return ErrDigestType
	}
	digestElem := reflect.ValueOf(digest).Elem()

	rfh := digestElem.FieldByName("h")
	rfh = reflect.NewAt(rfh.Type(), unsafe.Pointer(rfh.UnsafeAddr())).Elem()
	rfhp := (*[8]uint32)(unsafe.Pointer(rfh.UnsafeAddr()))
	*rfhp = state.H

	rfx := digestElem.FieldByName("x")
	rfx = reflect.NewAt(rfx.Type(), unsafe.Pointer(rfx.UnsafeAddr())).Elem()
	rfxp := (*[64]byte)(unsafe.Pointer(rfx.UnsafeAddr()))
	*rfxp = state.X

	rfnx := digestElem.FieldByName("nx")
	rfnx = reflect.NewAt(rfnx.Type(), unsafe.Pointer(rfnx.UnsafeAddr())).Elem()
	rfnxp := (*int)(unsafe.Pointer(rfnx.UnsafeAddr()))
	*rfnxp = state.Nx

	rfxLen := digestElem.FieldByName("len")
	rfxLen = reflect.NewAt(rfxLen.Type(), unsafe.Pointer(rfxLen.UnsafeAddr())).Elem()
	rfxLenP := (*uint64)(unsafe.Pointer(rfxLen.UnsafeAddr()))
	*rfxLenP = state.Len

	return nil
}

func NewHashWithStateText(stateCipherText string) (digest hash.Hash, err error) {
	state, err := DecodeStringToState(stateCipherText)
	if err != nil {
		return nil, err
	}
	digest = sha256.New()
	if err = SetHashState(digest, state); err != nil {
		return nil, err
	}
	return digest, nil
}

func GetHashStateText(digest hash.Hash) (string, error) {
	state, err := GetHashState(digest)
	if err != nil {
		return "", err
	}
	return state.EncodeToString()
}
