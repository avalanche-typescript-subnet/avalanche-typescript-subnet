package runtime

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type StateProvider func(string) ([]byte, error)

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

// payload

type JSPayload struct {
	CurrentState *map[string][]byte `json:"currentState"`
	Payload      []byte             `json:"payload"`
	FunctionName string             `json:"functionName"`
	Actor        []byte             `json:"actor"`
}

func (p JSPayload) MarshalJSON() ([]byte, error) {
	type Alias JSPayload
	encodedState := make(map[string][]byte)
	for k, v := range *p.CurrentState {
		encodedKey := base64.StdEncoding.EncodeToString([]byte(k))
		encodedState[encodedKey] = v
	}
	return json.Marshal(&struct {
		CurrentState map[string][]byte `json:"currentState"`
		*Alias
	}{
		CurrentState: encodedState,
		Alias:        (*Alias)(&p),
	})
}

// state provider

type DummyStateProvider struct {
	state map[string][]byte
}

func NewDummyStateProvider() *DummyStateProvider {
	return &DummyStateProvider{
		state: make(map[string][]byte),
	}
}

func (d *DummyStateProvider) StateProvider(key string) ([]byte, error) {
	value, exists := d.state[key]
	if exists {
		return value, nil
	} else {
		return []byte{}, nil
	}
}

func (d *DummyStateProvider) Update(newVals map[string][]byte) {
	for k, v := range newVals {
		d.state[k] = v
	}
}

//result json

type ResultJSON struct {
	Result      []byte            `json:"result"`
	Success     bool              `json:"success"`
	UpdatedKeys map[string][]byte `json:"updatedKeys"`
	ReadKeys    [][]byte          `json:"readKeys"`
	Error       string            `json:"error"`
}

func (r *ResultJSON) UnmarshalJSON(data []byte) error {
	type Alias ResultJSON
	aux := &struct {
		UpdatedKeys map[string][]byte `json:"updatedKeys"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	decodedKeys := make(map[string][]byte)
	for k, v := range aux.UpdatedKeys {
		decodedKey, err := base64.StdEncoding.DecodeString(k)
		if err != nil {
			return err
		}
		decodedKeys[string(decodedKey)] = v
	}
	r.UpdatedKeys = decodedKeys

	return nil
}
