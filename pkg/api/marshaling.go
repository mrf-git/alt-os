package api

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path"
	"strings"

	api_api_v0 "alt-os/api/api/v0"
	"encoding/json"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"gopkg.in/yaml.v3"
)

// ApiProtoMessage represents a single versioned API message.
type ApiProtoMessage struct {
	Kind    string
	Version string
	Def     proto.Message
}

// UnmarshalApiProtoMessages reads the specified file and unmarshals the protobuf API
// messages it contains. If format is empty it is inferred from the file extension.
func UnmarshalApiProtoMessages(filename, format string) ([]*ApiProtoMessage, error) {

	// Determine how to read the input file.
	if format == "" {
		format = strings.TrimPrefix(path.Ext(filename), ".")
	}
	var makeJsonReader func(filename string) (io.ReadCloser, error)
	switch format {
	default:
		return nil, errors.New("unrecognized format: " + format)
	case "json":
		makeJsonReader = func(filename string) (io.ReadCloser, error) { return os.Open(filename) }
	case "yml", "yaml":
		makeJsonReader = openYamlAsJson
	}

	// Read the message list structure from the file.
	var msgList api_api_v0.ApiMessageList
	if r, err := makeJsonReader(filename); err != nil {
		return nil, err
	} else {
		unmarshaler := &jsonpb.Unmarshaler{}
		err := unmarshaler.Unmarshal(r, &msgList)
		r.Close()
		if err != nil {
			return nil, err
		}
	}

	// Unmarshal each individual specific kind of message.
	var apiMessages []*ApiProtoMessage
	for _, msg := range msgList.Messages {
		if msg.Kind != "api."+msg.Def.TypeUrl {
			return nil, errors.New("kind/type mismatch")
		}
		if protoMsg, err := unmarshalKind(msg.Kind, msg.Version, msg.Def.Value); err != nil {
			return nil, err
		} else {
			apiMsg := &ApiProtoMessage{
				Kind:    msg.Kind,
				Version: msg.Version,
				Def:     protoMsg,
			}
			apiMessages = append(apiMessages, apiMsg)
		}
	}

	return apiMessages, nil
}

// MarshalApiProtoMessages marshals the specified messages and writes them to the specified
// output file. If format is empty it is inferred from the file extension.
func MarshalApiProtoMessages(messages []*ApiProtoMessage, filename, format string) error {
	// Determine how to write the output file.
	if format == "" {
		format = strings.TrimPrefix(path.Ext(filename), ".")
	}
	var makeJsonWriter func(filename string) (io.WriteCloser, error)
	switch format {
	default:
		return errors.New("unrecognized format: " + format)
	case "json":
		makeJsonWriter = func(filename string) (io.WriteCloser, error) { return os.Create(filename) }
	case "yml", "yaml":
		makeJsonWriter = createYamlAsJson
	}

	// Marshal each individual specific kind of message.
	var msgList api_api_v0.ApiMessageList
	for _, msg := range messages {
		kind, version, bytes, err := marshalKind(msg.Def)
		if err != nil {
			return nil
		}
		if kind != msg.Kind {
			return errors.New("kind/type mismatch")
		}
		if version != msg.Version {
			return errors.New("version mismatch")
		}
		apiMsg := &api_api_v0.ApiMessage{
			Kind:    kind,
			Version: version,
			Def: &types.Any{
				TypeUrl: strings.TrimPrefix(kind, "api."),
				Value:   bytes,
			},
		}
		msgList.Messages = append(msgList.Messages, apiMsg)
	}

	// Write the message list structure to the file.
	if w, err := makeJsonWriter(filename); err != nil {
		return err
	} else {
		marshaler := &jsonpb.Marshaler{}
		err := marshaler.Marshal(w, &msgList)
		w.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// openYamlAsJson reads the specified yaml file, converts it to json, and returns a
// reader for the json bytes.
func openYamlAsJson(filename string) (io.ReadCloser, error) {
	jsonMap := map[string][]map[string]interface{}{
		"messages": {},
	}
	if f, err := os.Open(filename); err != nil {
		return nil, err
	} else {
		decoder := yaml.NewDecoder(f)
		for {
			yamlMap := make(map[string]interface{})
			if err := decoder.Decode(&yamlMap); errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return nil, err
			}
			protoType := strings.TrimPrefix(yamlMap["kind"].(string), "api.")
			defMap := yamlMap["def"].(map[string]interface{})
			defMap["@type"] = protoType
			jsonMap["messages"] = append(jsonMap["messages"], yamlMap)
		}
		f.Close()
	}
	if data, err := json.Marshal(jsonMap); err != nil {
		return nil, err
	} else {
		return io.NopCloser(bytes.NewReader(data)), nil
	}
}

// createYamlAsJson creates and returns a new writer that accepts json bytes, converts
// json into yaml, and then writes yaml bytes to the output file.
func createYamlAsJson(filename string) (io.WriteCloser, error) {
	if f, err := os.Create(filename); err != nil {
		return nil, err
	} else {
		return &jsonYamlWriter{
			File:     f,
			JsonData: bytes.NewBuffer(nil),
		}, nil
	}
}

// jsonYamlWriter implements an io.WriteCloser for writing json as yaml.
type jsonYamlWriter struct {
	File     *os.File
	JsonData *bytes.Buffer
}

func (w *jsonYamlWriter) Write(p []byte) (n int, err error) {
	return w.JsonData.Write(p)
}

func (w *jsonYamlWriter) Close() (err error) {
	type yamlMsgStruct struct {
		Kind    string                 `yaml:"kind"`
		Version string                 `yaml:"version"`
		Def     map[string]interface{} `yaml:"def"`
	}
	jsonMap := make(map[string][]*yamlMsgStruct)
	err = json.Unmarshal(w.JsonData.Bytes(), &jsonMap)
	if err == nil {
		encoder := yaml.NewEncoder(w.File)
		encoder.SetIndent(2)
		for _, yamlMsg := range jsonMap["messages"] {
			delete(yamlMsg.Def, "@type")
			err = encoder.Encode(yamlMsg)
			if err != nil {
				break
			}
		}
	}
	if err == nil {
		err = w.File.Close()

	} else {
		w.File.Close()
	}
	return err
}
