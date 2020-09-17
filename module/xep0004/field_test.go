package xep0004

import (
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func TestModule_XEP0004_Field_FromElement(t *testing.T) {
	elem := xmpp.NewElementName("")
	_, err := NewFieldFromElement(elem)
	require.NotNil(t, err)

	elem.SetName("field")
	elem.SetAttribute("var", "name")
	elem.SetAttribute("type", "integer")
	_, err = NewFieldFromElement(elem)
	require.NotNil(t, err)

	elem.SetAttribute("type", TextSingle)
	_, err = NewFieldFromElement(elem)
	require.Nil(t, err)

	desc := "A description"
	descElem := xmpp.NewElementName("desc")
	descElem.SetText(desc)
	elem.AppendElement(descElem)
	elem.AppendElement(xmpp.NewElementName("required"))
	f, err := NewFieldFromElement(elem)
	require.Nil(t, err)
	require.Equal(t, desc, f.Description)
	require.True(t, f.Required)

	value := "A value"
	valueElem := xmpp.NewElementName("value")
	valueElem.SetText(value)
	elem.AppendElement(valueElem)
	f, err = NewFieldFromElement(elem)
	require.Nil(t, err)
	require.Len(t, f.Values, 1)
	require.Equal(t, value, f.Values[0])
	elem.RemoveElements("value")

	optValue := "An option value"
	valueElem.SetText(optValue)
	optElem := xmpp.NewElementName("option")
	optElem.SetAttribute("label", "news")
	optElem.AppendElement(valueElem)
	elem.AppendElement(optElem)
	f, err = NewFieldFromElement(elem)
	require.Nil(t, err)
	require.Len(t, f.Options, 1)
	require.Equal(t, "news", f.Options[0].Label)
	require.Equal(t, optValue, f.Options[0].Value)
}

func TestModule_XEP0004_Field_Element(t *testing.T) {
	f := Field{Var: "a_var"}
	f.Type = "a_type"
	f.Label = "a_label"
	f.Required = true
	f.Description = "A description"
	f.Values = []string{"A value"}
	f.Options = []Option{{"opt_label", "An option value"}}
	elem := f.Element()

	require.Equal(t, "field", elem.Name())
	require.Equal(t, "a_var", elem.Attributes().Get("var"))
	require.Equal(t, "a_type", elem.Attributes().Get("type"))
	require.Equal(t, "a_label", elem.Attributes().Get("label"))

	valElem := elem.Elements().Child("value")
	require.NotNil(t, valElem)
	require.Equal(t, "A value", valElem.Text())

	optElems := elem.Elements().Children("option")
	require.Len(t, optElems, 1)
	optElem := optElems[0]
	require.Equal(t, "opt_label", optElem.Attributes().Get("label"))

	valElem = optElem.Elements().Child("value")
	require.Equal(t, "An option value", valElem.Text())
}
