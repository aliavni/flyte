// Code generated by go generate; DO NOT EDIT.
// This file was generated by robots.

package launchplan

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

var dereferencableKindsUpdateConfig = map[reflect.Kind]struct{}{
	reflect.Array: {}, reflect.Chan: {}, reflect.Map: {}, reflect.Ptr: {}, reflect.Slice: {},
}

// Checks if t is a kind that can be dereferenced to get its underlying type.
func canGetElementUpdateConfig(t reflect.Kind) bool {
	_, exists := dereferencableKindsUpdateConfig[t]
	return exists
}

// This decoder hook tests types for json unmarshaling capability. If implemented, it uses json unmarshal to build the
// object. Otherwise, it'll just pass on the original data.
func jsonUnmarshalerHookUpdateConfig(_, to reflect.Type, data interface{}) (interface{}, error) {
	unmarshalerType := reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
	if to.Implements(unmarshalerType) || reflect.PtrTo(to).Implements(unmarshalerType) ||
		(canGetElementUpdateConfig(to.Kind()) && to.Elem().Implements(unmarshalerType)) {

		raw, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Failed to marshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		res := reflect.New(to).Interface()
		err = json.Unmarshal(raw, &res)
		if err != nil {
			fmt.Printf("Failed to umarshal Data: %v. Error: %v. Skipping jsonUnmarshalHook", data, err)
			return data, nil
		}

		return res, nil
	}

	return data, nil
}

func decode_UpdateConfig(input, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			jsonUnmarshalerHookUpdateConfig,
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

func join_UpdateConfig(arr interface{}, sep string) string {
	listValue := reflect.ValueOf(arr)
	strs := make([]string, 0, listValue.Len())
	for i := 0; i < listValue.Len(); i++ {
		strs = append(strs, fmt.Sprintf("%v", listValue.Index(i)))
	}

	return strings.Join(strs, sep)
}

func testDecodeJson_UpdateConfig(t *testing.T, val, result interface{}) {
	assert.NoError(t, decode_UpdateConfig(val, result))
}

func testDecodeRaw_UpdateConfig(t *testing.T, vStringSlice, result interface{}) {
	assert.NoError(t, decode_UpdateConfig(vStringSlice, result))
}

func TestUpdateConfig_GetPFlagSet(t *testing.T) {
	val := UpdateConfig{}
	cmdFlags := val.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())
}

func TestUpdateConfig_SetFlags(t *testing.T) {
	actual := UpdateConfig{}
	cmdFlags := actual.GetPFlagSet("")
	assert.True(t, cmdFlags.HasFlags())

	t.Run("Test_activate", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("activate", testValue)
			if vBool, err := cmdFlags.GetBool("activate"); err == nil {
				testDecodeJson_UpdateConfig(t, fmt.Sprintf("%v", vBool), &actual.Activate)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_archive", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("archive", testValue)
			if vBool, err := cmdFlags.GetBool("archive"); err == nil {
				testDecodeJson_UpdateConfig(t, fmt.Sprintf("%v", vBool), &actual.Archive)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_deactivate", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("deactivate", testValue)
			if vBool, err := cmdFlags.GetBool("deactivate"); err == nil {
				testDecodeJson_UpdateConfig(t, fmt.Sprintf("%v", vBool), &actual.Deactivate)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_dryRun", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("dryRun", testValue)
			if vBool, err := cmdFlags.GetBool("dryRun"); err == nil {
				testDecodeJson_UpdateConfig(t, fmt.Sprintf("%v", vBool), &actual.DryRun)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_force", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("force", testValue)
			if vBool, err := cmdFlags.GetBool("force"); err == nil {
				testDecodeJson_UpdateConfig(t, fmt.Sprintf("%v", vBool), &actual.Force)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
	t.Run("Test_version", func(t *testing.T) {

		t.Run("Override", func(t *testing.T) {
			testValue := "1"

			cmdFlags.Set("version", testValue)
			if vString, err := cmdFlags.GetString("version"); err == nil {
				testDecodeJson_UpdateConfig(t, fmt.Sprintf("%v", vString), &actual.Version)

			} else {
				assert.FailNow(t, err.Error())
			}
		})
	})
}
