package feature_test

import (
	"fmt"
	"strings"
)

// example from https://github.com/abice/go-enum

type Color int32

const (
	// ColorBlack is a Color of type Black.
	ColorBlack Color = iota
	// ColorWhite is a Color of type White.
	ColorWhite
	// ColorRed is a Color of type Red.
	ColorRed
	// ColorGreen is a Color of type Green.
	// Green starts with 33
	ColorGreen Color = iota + 30
	// ColorBlue is a Color of type Blue.
	ColorBlue
	// ColorGrey is a Color of type Grey.
	ColorGrey
	// ColorYellow is a Color of type Yellow.
	ColorYellow
	// ColorBlueGreen is a Color of type Blue-Green.
	ColorBlueGreen
	// ColorRedOrange is a Color of type Red-Orange.
	ColorRedOrange
	// ColorYellowGreen is a Color of type Yellow_green.
	ColorYellowGreen
	// ColorRedOrangeBlue is a Color of type Red-Orange-Blue.
	ColorRedOrangeBlue
)

const _ColorName = "BlackWhiteRedGreenBluegreyyellowblue-greenred-orangeyellow_greenred-orange-blue"

var _ColorMap = map[Color]string{
	ColorBlack:         _ColorName[0:5],
	ColorWhite:         _ColorName[5:10],
	ColorRed:           _ColorName[10:13],
	ColorGreen:         _ColorName[13:18],
	ColorBlue:          _ColorName[18:22],
	ColorGrey:          _ColorName[22:26],
	ColorYellow:        _ColorName[26:32],
	ColorBlueGreen:     _ColorName[32:42],
	ColorRedOrange:     _ColorName[42:52],
	ColorYellowGreen:   _ColorName[52:64],
	ColorRedOrangeBlue: _ColorName[64:79],
}

// String implements the Stringer interface.
func (x Color) String() string {
	if str, ok := _ColorMap[x]; ok {
		return str
	}
	return fmt.Sprintf("Color(%d)", x)
}

var _ColorValue = map[string]Color{
	_ColorName[0:5]:                    ColorBlack,
	strings.ToLower(_ColorName[0:5]):   ColorBlack,
	_ColorName[5:10]:                   ColorWhite,
	strings.ToLower(_ColorName[5:10]):  ColorWhite,
	_ColorName[10:13]:                  ColorRed,
	strings.ToLower(_ColorName[10:13]): ColorRed,
	_ColorName[13:18]:                  ColorGreen,
	strings.ToLower(_ColorName[13:18]): ColorGreen,
	_ColorName[18:22]:                  ColorBlue,
	strings.ToLower(_ColorName[18:22]): ColorBlue,
	_ColorName[22:26]:                  ColorGrey,
	strings.ToLower(_ColorName[22:26]): ColorGrey,
	_ColorName[26:32]:                  ColorYellow,
	strings.ToLower(_ColorName[26:32]): ColorYellow,
	_ColorName[32:42]:                  ColorBlueGreen,
	strings.ToLower(_ColorName[32:42]): ColorBlueGreen,
	_ColorName[42:52]:                  ColorRedOrange,
	strings.ToLower(_ColorName[42:52]): ColorRedOrange,
	_ColorName[52:64]:                  ColorYellowGreen,
	strings.ToLower(_ColorName[52:64]): ColorYellowGreen,
	_ColorName[64:79]:                  ColorRedOrangeBlue,
	strings.ToLower(_ColorName[64:79]): ColorRedOrangeBlue,
}

// ParseColor attempts to convert a string to a Color
func ParseColor(name string) (Color, error) {
	if x, ok := _ColorValue[name]; ok {
		return x, nil
	}
	return Color(0), fmt.Errorf("%s is not a valid Color", name)
}

func (x Color) Ptr() *Color {
	return &x
}

// MarshalText implements the text marshaller method
func (x Color) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method
func (x *Color) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseColor(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
