package gql

import "fmt"

type SimplifiedInterfaces map[string]*SimplifiedInterface

func NewSimplifiedInterfaces() SimplifiedInterfaces {
	return make(SimplifiedInterfaces)
}

func (m SimplifiedInterfaces) Put(simplifiedInterface *SimplifiedInterface) {
	m[simplifiedInterface.Name] = simplifiedInterface
}

func (m SimplifiedInterfaces) HasInterface(name string) bool {
	_, ok := m[name]
	return ok
}

func (m SimplifiedInterfaces) ApplyInterfaces(newType, oldType *SimplifiedType) error {
	for name, interf := range m {
		if oldType != nil {
			if oldType.HasInterface(name) {
				err := newType.AddInterface(interf)
				if err != nil {
					return err
				}
				continue
			}
		}
		if interf.ShouldImplement(newType) {
			err := newType.AddInterface(interf)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m SimplifiedInterfaces) GetObjectTypeFields(name string) []*SimplifiedField {
	objFields := make([]*SimplifiedField, 0)
	fields := m[name].GetObjectFields()
	for _, field := range fields {
		if !m.HasInterface(field.Name) {
			objFields = append(objFields, field)
		}
	}

	return objFields
}

type SimplifiedInterface struct {
	*SimplifiedBaseType
	SignatureFields []string
}

func NewSimplifiedInterface(name string, fields map[string]*SimplifiedField, signatureFields []string) *SimplifiedInterface {
	return &SimplifiedInterface{
		SimplifiedBaseType: NewSimplifiedBaseType(name, fields),
		SignatureFields:    signatureFields,
	}
}

func (m *SimplifiedInterface) ShouldImplement(simplifiedType *SimplifiedType) bool {
	for _, signatureField := range m.SignatureFields {
		if !simplifiedType.HasField(signatureField) {
			return false
		}
	}
	return true
}

func (m *SimplifiedInterface) Validate() error {
	if len(m.SignatureFields) == 0 {
		return fmt.Errorf("invalid interface: %v, it must have at least one signature field", m.Name)
	}
	return nil
}