package pkt

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/klintcheng/kim/wire"
	"github.com/klintcheng/kim/wire/endian"
	"google.golang.org/protobuf/proto"
)

// LogicPkt 逻辑协议
type LogicPkt struct {
	// Header 使用protobuf序列化
	Header
	Body []byte `json:"body,omitempty"`
}

// HeaderOption
type HeaderOption func(*Header)

// WithStatus
func WithStatus(status Status) HeaderOption {
	return func(h *Header) {
		h.Status = status
	}
}

// WithSeq
func WithSeq(seq uint32) HeaderOption {
	return func(h *Header) {
		h.Sequence = seq
	}
}

// WithChannel set channelID
func WithChannel(channelID string) HeaderOption {
	return func(h *Header) {
		h.ChannelId = channelID
	}
}

// WithDest
func WithDest(dest string) HeaderOption {
	return func(h *Header) {
		h.Dest = dest
	}
}

// New an empty payload message
func New(command string, options ...HeaderOption) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Command = command

	for _, option := range options {
		option(&pkt.Header)
	}
	if pkt.Sequence == 0 {
		pkt.Sequence = wire.Seq.Next()
	}
	return pkt
}

// NewLogicPkt new LogicPacket from a header
func NewLogicPkt(header *Header) *LogicPkt {
	pkt := &LogicPkt{}
	pkt.Header = Header{
		Command:   header.Command,
		Sequence:  header.Sequence,
		ChannelId: header.ChannelId,
		Status:    header.Status,
		Dest:      header.Dest,
	}
	return pkt
}

// Decode ReadPkt read bytes to LogicPkt from a reader
func (p *LogicPkt) Decode(r io.Reader) error {
	headerBytes, err := endian.ReadBytes(r)
	if err != nil {
		return err
	}
	if err := proto.Unmarshal(headerBytes, &p.Header); err != nil {
		return err
	}
	// read body
	p.Body, err = endian.ReadBytes(r)
	if err != nil {
		return err
	}
	return nil
}

// Encode Header to a writer
func (p *LogicPkt) Encode(w io.Writer) error {
	headerBytes, err := proto.Marshal(&p.Header)
	if err != nil {
		return err
	}
	if err := endian.WriteBytes(w, headerBytes); err != nil {
		return err
	}
	if err := endian.WriteBytes(w, p.Body); err != nil {
		return err
	}
	return nil
}

// ReadBody val must be a pointer
func (p *LogicPkt) ReadBody(val proto.Message) error {
	return json.Unmarshal(p.Body, val)
}

// WriteBody WritePb
func (p *LogicPkt) WriteBody(val proto.Message) *LogicPkt {
	if val == nil {
		return p
	}
	p.Body, _ = json.Marshal(val)
	return p
}

// StringBody return string body
func (p *LogicPkt) StringBody() string {
	return string(p.Body)
}

func (p *LogicPkt) String() string {
	return fmt.Sprintf("header:%v body:%dbits", &p.Header, len(p.Body))
}

// ServiceName
// 这里涉及一个服务定位逻辑，
func (h *Header) ServiceName() string {
	arr := strings.SplitN(h.Command, ".", 2)
	if len(arr) <= 1 {
		return "default"
	}
	return arr[0]
}

// AddMeta AddMeta
func (p *LogicPkt) AddMeta(m ...*Meta) {
	p.Meta = append(p.Meta, m...)
}

// AddStringMeta AddStringMeta
func (p *LogicPkt) AddStringMeta(key, value string) {
	p.AddMeta(&Meta{
		Key:   key,
		Value: value,
		Type:  MetaType_string,
	})
}

// GetMeta extra value
func (p *LogicPkt) GetMeta(key string) (interface{}, bool) {
	for _, m := range p.Meta {
		if m.Key == key {
			switch m.Type {
			case MetaType_int:
				v, _ := strconv.Atoi(m.Value)
				return v, true
			case MetaType_float:
				v, _ := strconv.ParseFloat(m.Value, 64)
				return v, true
			}
			return m.Value, true
		}
	}
	return nil, false
}

// DelMeta DelMeta
func (p *LogicPkt) DelMeta(key string) {
	for i, m := range p.Meta {
		if m.Key == key {
			length := len(p.Meta)
			if i < length-1 {
				copy(p.Meta[i:], p.Meta[i+1:])
			}
			p.Meta = p.Meta[:length-1]
		}
	}
}
