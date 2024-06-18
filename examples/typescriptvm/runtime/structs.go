package runtime

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type ResultJSON struct {
	Result      []byte          `json:"result"`
	Success     bool            `json:"success"`
	UpdatedKeys ContactStateMap `json:"updatedKeys"`
	ReadKeys    [][]byte        `json:"readKeys"`
	Error       string          `json:"error"`
}

type StateProvider func([]byte) ([]byte, error)

type JavyExecParams struct {
	MaxFuel       uint64
	MaxTime       time.Duration
	MaxMemory     int64
	Bytecode      *[]byte
	StateProvider StateProvider
	Payload       []byte
	FunctionName  string
	Actor         []byte
}

type JSPayload struct {
	CurrentState *ContactStateMap `json:"currentState"`
	Payload      []byte           `json:"payload"`
	FunctionName string           `json:"functionName"`
	Actor        []byte           `json:"actor"`
}

// Convert KeyPostfix to base64 string
func byteArrayToBase64String(arr []byte) string {
	return base64.StdEncoding.EncodeToString(arr)
}

// Convert base64 string to KeyPostfix
func base64StringToByteArray(str string) ([]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return []byte{}, err
	}
	var arr []byte
	copy(arr[:], bytes)
	return arr, nil
}

func NewContactStateMap() *ContactStateMap {
	return &ContactStateMap{
		data: make(map[string][]byte),
	}
}

type ContactStateMap struct {
	data map[string][]byte
}

func (m *ContactStateMap) Set(key []byte, value []byte) {
	m.data[string(key)] = value
}

func (m *ContactStateMap) Get(key []byte) ([]byte, bool) {
	val, ok := m.data[string(key)]

	return val, ok
}

func (m *ContactStateMap) Data() map[string][]byte {
	return m.data
}

func (m *ContactStateMap) MarshalJSON() ([]byte, error) {
	encodedData := make(map[string]string)
	for key, value := range m.data {
		keyBase64 := base64.StdEncoding.EncodeToString([]byte(key))
		encodedValue := base64.StdEncoding.EncodeToString(value)
		encodedData[keyBase64] = encodedValue
	}
	return json.Marshal(encodedData)
}

func (m *ContactStateMap) UnmarshalJSON(data []byte) error {
	var decodedData map[string]string
	if err := json.Unmarshal(data, &decodedData); err != nil {
		return err
	}

	m.data = make(map[string][]byte)
	for keyBase64, valueBase64 := range decodedData {
		key, err := base64.StdEncoding.DecodeString(keyBase64)
		if err != nil {
			return err
		}
		value, err := base64.StdEncoding.DecodeString(valueBase64)
		if err != nil {
			return err
		}
		m.data[string(key)] = value
	}
	return nil
}

type DummyStateProvider struct {
	state ContactStateMap
}

func NewDummyStateProvider() *DummyStateProvider {
	return &DummyStateProvider{
		state: *NewContactStateMap(),
	}
}

func (d *DummyStateProvider) StateProvider(key []byte) ([]byte, error) {
	value, exists := d.state.Get(key)
	if !exists {
		return []byte{}, nil
	}
	return value, nil
}

func (d *DummyStateProvider) Update(newVals ContactStateMap) {
	for k, v := range newVals.data {
		d.state.Set([]byte(k), v)
	}
}
