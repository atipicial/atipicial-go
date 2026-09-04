package standard

import (
	"testing"

	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/stretchr/testify/require"
)

func fooMethodBarEvent() *manifest.Manifest {
	return &manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name: "foo",
					Parameters: []manifest.Parameter{
						{Type: smartcontract.ByteArrayType},
						{Type: smartcontract.PublicKeyType},
					},
					ReturnType: smartcontract.IntegerType,
					Safe:       true,
				},
			},
			Events: []manifest.Event{
				{
					Name: "bar",
					Parameters: []manifest.Parameter{
						{Type: smartcontract.StringType},
					},
				},
			},
		},
	}
}

func TestComplyMissingMethod(t *testing.T) {
	m := fooMethodBarEvent()
	m.ABI.GetMethod("foo", -1).Name = "notafoo"
	err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
	require.ErrorIs(t, err, ErrMethodMissing)
}

func TestComplyInvalidReturnType(t *testing.T) {
	m := fooMethodBarEvent()
	m.ABI.GetMethod("foo", -1).ReturnType = smartcontract.VoidType
	err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
	require.ErrorIs(t, err, ErrInvalidReturnType)
}

func TestComplyMethodParameterCount(t *testing.T) {
	t.Run("Method", func(t *testing.T) {
		m := fooMethodBarEvent()
		f := m.ABI.GetMethod("foo", -1)
		f.Parameters = append(f.Parameters, manifest.Parameter{Type: smartcontract.BoolType})
		err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
		require.ErrorIs(t, err, ErrMethodMissing)
	})
	t.Run("Event", func(t *testing.T) {
		m := fooMethodBarEvent()
		ev := m.ABI.GetEvent("bar")
		ev.Parameters = ev.Parameters[:0]
		err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
		require.ErrorIs(t, err, ErrInvalidParameterCount)
	})
}

func TestComplyParameterType(t *testing.T) {
	t.Run("Method", func(t *testing.T) {
		m := fooMethodBarEvent()
		m.ABI.GetMethod("foo", -1).Parameters[0].Type = smartcontract.InteropInterfaceType
		err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
		require.ErrorIs(t, err, ErrInvalidParameterType)
	})
	t.Run("Event", func(t *testing.T) {
		m := fooMethodBarEvent()
		m.ABI.GetEvent("bar").Parameters[0].Type = smartcontract.InteropInterfaceType
		err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
		require.ErrorIs(t, err, ErrInvalidParameterType)
	})
}

func TestComplyParameterName(t *testing.T) {
	t.Run("Method", func(t *testing.T) {
		m := fooMethodBarEvent()
		m.ABI.GetMethod("foo", -1).Parameters[0].Name = "hehe"
		s := &Standard{Manifest: *fooMethodBarEvent()}
		err := Comply(m, s)
		require.ErrorIs(t, err, ErrInvalidParameterName)
		require.NoError(t, ComplyABI(m, s))
	})
	t.Run("Event", func(t *testing.T) {
		m := fooMethodBarEvent()
		m.ABI.GetEvent("bar").Parameters[0].Name = "hehe"
		s := &Standard{Manifest: *fooMethodBarEvent()}
		err := Comply(m, s)
		require.ErrorIs(t, err, ErrInvalidParameterName)
		require.NoError(t, ComplyABI(m, s))
	})
}

func TestMissingEvent(t *testing.T) {
	m := fooMethodBarEvent()
	m.ABI.GetEvent("bar").Name = "notabar"
	err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
	require.ErrorIs(t, err, ErrEventMissing)
}

func TestSafeFlag(t *testing.T) {
	m := fooMethodBarEvent()
	m.ABI.GetMethod("foo", -1).Safe = false
	err := Comply(m, &Standard{Manifest: *fooMethodBarEvent()})
	require.ErrorIs(t, err, ErrSafeMethodMismatch)
}

func TestComplyValid(t *testing.T) {
	m := fooMethodBarEvent()
	m.ABI.Methods = append(m.ABI.Methods, manifest.Method{
		Name:       "newmethod",
		Offset:     123,
		ReturnType: smartcontract.ByteArrayType,
	})
	m.ABI.Events = append(m.ABI.Events, manifest.Event{
		Name: "otherevent",
		Parameters: []manifest.Parameter{{
			Name: "names do not matter",
			Type: smartcontract.IntegerType,
		}},
	})
	require.NoError(t, Comply(m, &Standard{Manifest: *fooMethodBarEvent()}))
}

func TestCheck(t *testing.T) {
	m := manifest.NewManifest("Test")
	require.Error(t, Check(m, manifest.AEP17StandardName))

	m.ABI.Methods = append(m.ABI.Methods, DecimalTokenBase.ABI.Methods...)
	m.ABI.Methods = append(m.ABI.Methods, Aep17.ABI.Methods...)
	m.ABI.Events = append(m.ABI.Events, Aep17.ABI.Events...)
	require.NoError(t, Check(m, manifest.AEP17StandardName))
	require.NoError(t, CheckABI(m, manifest.AEP17StandardName))
}

func TestCheck_AEP22(t *testing.T) {
	m := manifest.NewManifest("Test")
	require.Error(t, Check(m, manifest.AEP22StandardName))

	m.ABI.Methods = append(m.ABI.Methods, Aep22.ABI.Methods...)
	m.ABI.Events = append(m.ABI.Events, Aep22.ABI.Events...)
	require.NoError(t, Check(m, manifest.AEP22StandardName))
}

func TestCheck_AEP29(t *testing.T) {
	m := manifest.NewManifest("Test")
	require.Error(t, Check(m, manifest.AEP29StandardName))

	m.ABI.Methods = append(m.ABI.Methods, Aep29.ABI.Methods...)
	require.NoError(t, Check(m, manifest.AEP29StandardName))
}

func TestCheck_AEP30(t *testing.T) {
	m := manifest.NewManifest("Test")
	require.Error(t, Check(m, manifest.AEP30StandardName))

	m.ABI.Methods = append(m.ABI.Methods, Aep30.ABI.Methods...)
	require.NoError(t, Check(m, manifest.AEP30StandardName))
}

func TestCheck_AEP31(t *testing.T) {
	m := manifest.NewManifest("Test")
	require.Error(t, Check(m, manifest.AEP31StandardName))

	m.ABI.Methods = append(m.ABI.Methods, Aep31.ABI.Methods...)
	m.ABI.Events = append(m.ABI.Events, Aep31.ABI.Events...)
	require.NoError(t, Check(m, manifest.AEP31StandardName))
}

func TestOptional(t *testing.T) {
	var m Standard
	m.Optional = []manifest.Method{{
		Name:       "optMet",
		Parameters: []manifest.Parameter{{Type: smartcontract.ByteArrayType}},
		ReturnType: smartcontract.IntegerType,
	}}

	t.Run("wrong parameter count, do not check", func(t *testing.T) {
		var actual manifest.Manifest
		actual.ABI.Methods = []manifest.Method{{
			Name:       "optMet",
			ReturnType: smartcontract.IntegerType,
		}}
		require.NoError(t, Comply(&actual, &m))
	})
	t.Run("good parameter count, bad return", func(t *testing.T) {
		var actual manifest.Manifest
		actual.ABI.Methods = []manifest.Method{{
			Name:       "optMet",
			Parameters: []manifest.Parameter{{Type: smartcontract.ArrayType}},
			ReturnType: smartcontract.IntegerType,
		}}
		require.Error(t, Comply(&actual, &m))
	})
	t.Run("good parameter count, good return", func(t *testing.T) {
		var actual manifest.Manifest
		actual.ABI.Methods = []manifest.Method{{
			Name:       "optMet",
			Parameters: []manifest.Parameter{{Type: smartcontract.ByteArrayType}},
			ReturnType: smartcontract.IntegerType,
		}}
		require.NoError(t, Comply(&actual, &m))
	})
}
