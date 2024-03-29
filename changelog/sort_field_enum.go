// Code generated by "enumer -gqlgen -type SortField --output sort_field_enum.go --trimprefix SortField"; DO NOT EDIT.

package changelog

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

const _SortFieldName = "TimestampUserOperationEntity"

var _SortFieldIndex = [...]uint8{0, 9, 13, 22, 28}

const _SortFieldLowerName = "timestampuseroperationentity"

func (i SortField) String() string {
	if i < 0 || i >= SortField(len(_SortFieldIndex)-1) {
		return fmt.Sprintf("SortField(%d)", i)
	}
	return _SortFieldName[_SortFieldIndex[i]:_SortFieldIndex[i+1]]
}

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the stringer command to generate them again.
func _SortFieldNoOp() {
	var x [1]struct{}
	_ = x[SortFieldTimestamp-(0)]
	_ = x[SortFieldUser-(1)]
	_ = x[SortFieldOperation-(2)]
	_ = x[SortFieldEntity-(3)]
}

var _SortFieldValues = []SortField{SortFieldTimestamp, SortFieldUser, SortFieldOperation, SortFieldEntity}

var _SortFieldNameToValueMap = map[string]SortField{
	_SortFieldName[0:9]:        SortFieldTimestamp,
	_SortFieldLowerName[0:9]:   SortFieldTimestamp,
	_SortFieldName[9:13]:       SortFieldUser,
	_SortFieldLowerName[9:13]:  SortFieldUser,
	_SortFieldName[13:22]:      SortFieldOperation,
	_SortFieldLowerName[13:22]: SortFieldOperation,
	_SortFieldName[22:28]:      SortFieldEntity,
	_SortFieldLowerName[22:28]: SortFieldEntity,
}

var _SortFieldNames = []string{
	_SortFieldName[0:9],
	_SortFieldName[9:13],
	_SortFieldName[13:22],
	_SortFieldName[22:28],
}

// SortFieldString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func SortFieldString(s string) (SortField, error) {
	if val, ok := _SortFieldNameToValueMap[s]; ok {
		return val, nil
	}

	if val, ok := _SortFieldNameToValueMap[strings.ToLower(s)]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to SortField values", s)
}

// SortFieldValues returns all values of the enum
func SortFieldValues() []SortField {
	return _SortFieldValues
}

// SortFieldStrings returns a slice of all String values of the enum
func SortFieldStrings() []string {
	strs := make([]string, len(_SortFieldNames))
	copy(strs, _SortFieldNames)
	return strs
}

// IsASortField returns "true" if the value is listed in the enum definition. "false" otherwise
func (i SortField) IsASortField() bool {
	for _, v := range _SortFieldValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalGQL implements the graphql.Marshaler interface for SortField
func (i SortField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(i.String()))
}

// UnmarshalGQL implements the graphql.Unmarshaler interface for SortField
func (i *SortField) UnmarshalGQL(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("SortField should be a string, got %T", value)
	}

	var err error
	*i, err = SortFieldString(str)
	return err
}
