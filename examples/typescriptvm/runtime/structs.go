package runtime

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type ResultJSON struct {
	Result      []byte             `json:"result"`
	Success     bool               `json:"success"`
	UpdatedKeys map[[4]byte][]byte `json:"updatedKeys"`
	ReadKeys    [][4]byte          `json:"readKeys"`
	Error       string             `json:"error"`
}

type StateProvider func([4]byte) ([]byte, error)

type JavyExecParams struct {
	MaxFuel       uint64
	MaxTime       time.Duration
	MaxMemory     int64
	Bytecode      *[]byte
	StateProvider StateProvider
	Payload       []byte
	Actor         []byte
}

type JSPayload struct {
	CurrentState map[[4]byte][]byte `json:"currentState"`
	Payload      []byte             `json:"payload"`
	Actor        []byte             `json:"actor"`
}

// Convert [4]byte to base64 string
func byteArrayToBase64String(arr [4]byte) string {
	return base64.StdEncoding.EncodeToString(arr[:])
}

// Convert base64 string to [4]byte
func base64StringToByteArray(str string) ([4]byte, error) {
	bytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return [4]byte{}, err
	}
	var arr [4]byte
	copy(arr[:], bytes)
	return arr, nil
}

// Custom JSON marshalling for ResultJSON
func (r ResultJSON) MarshalJSON() ([]byte, error) {
	updatedKeys := make(map[string]string)
	for k, v := range r.UpdatedKeys {
		updatedKeys[byteArrayToBase64String(k)] = base64.StdEncoding.EncodeToString(v)
	}

	readKeys := make([]string, len(r.ReadKeys))
	for i, key := range r.ReadKeys {
		readKeys[i] = byteArrayToBase64String(key)
	}

	return json.Marshal(&struct {
		Result      string            `json:"result"`
		Success     bool              `json:"success"`
		UpdatedKeys map[string]string `json:"updatedKeys"`
		ReadKeys    []string          `json:"readKeys"`
		Error       string            `json:"error"`
	}{
		Result:      base64.StdEncoding.EncodeToString(r.Result),
		Success:     r.Success,
		UpdatedKeys: updatedKeys,
		ReadKeys:    readKeys,
		Error:       r.Error,
	})
}

// Custom JSON unmarshalling for ResultJSON
func (r *ResultJSON) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Result      string            `json:"result"`
		Success     bool              `json:"success"`
		UpdatedKeys map[string]string `json:"updatedKeys"`
		ReadKeys    []string          `json:"readKeys"`
		Error       string            `json:"error"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	result, err := base64.StdEncoding.DecodeString(aux.Result)
	if err != nil {
		return err
	}
	r.Result = result
	r.Success = aux.Success
	r.Error = aux.Error

	updatedKeys := make(map[[4]byte][]byte)
	for k, v := range aux.UpdatedKeys {
		key, err := base64StringToByteArray(k)
		if err != nil {
			return err
		}
		value, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		updatedKeys[key] = value
	}
	r.UpdatedKeys = updatedKeys

	readKeys := make([][4]byte, len(aux.ReadKeys))
	for i, key := range aux.ReadKeys {
		readKey, err := base64StringToByteArray(key)
		if err != nil {
			return err
		}
		readKeys[i] = readKey
	}
	r.ReadKeys = readKeys

	return nil
}

// Custom JSON marshalling for JSPayload
func (j JSPayload) MarshalJSON() ([]byte, error) {
	currentState := make(map[string]string)
	for k, v := range j.CurrentState {
		currentState[byteArrayToBase64String(k)] = base64.StdEncoding.EncodeToString(v)
	}

	return json.Marshal(&struct {
		CurrentState map[string]string `json:"currentState"`
		Payload      string            `json:"payload"`
		Actor        string            `json:"actor"`
	}{
		CurrentState: currentState,
		Payload:      base64.StdEncoding.EncodeToString(j.Payload),
		Actor:        base64.StdEncoding.EncodeToString(j.Actor),
	})
}

// Custom JSON unmarshalling for JSPayload
func (j *JSPayload) UnmarshalJSON(data []byte) error {
	aux := &struct {
		CurrentState map[string]string `json:"currentState"`
		Payload      string            `json:"payload"`
		Actor        string            `json:"actor"`
	}{}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	currentState := make(map[[4]byte][]byte)
	for k, v := range aux.CurrentState {
		key, err := base64StringToByteArray(k)
		if err != nil {
			return err
		}
		value, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		currentState[key] = value
	}
	j.CurrentState = currentState

	payload, err := base64.StdEncoding.DecodeString(aux.Payload)
	if err != nil {
		return err
	}
	j.Payload = payload

	actor, err := base64.StdEncoding.DecodeString(aux.Actor)
	if err != nil {
		return err
	}
	j.Actor = actor

	return nil
}
