package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FlexFloat allows unmarshaling JSON numbers or strings with commas/dots into float64
type FlexFloat float64

func (f *FlexFloat) UnmarshalJSON(b []byte) error {
	str := string(b)
	str = strings.Trim(str, `"`)

	if str == "" {
		*f = 0
		return nil
	}

	// Replace comma with dot for Indonesian locale parsing
	str = strings.ReplaceAll(str, ",", ".")

	// If there are multiple dots, remove all but the last one (assuming thousand separators were dots)
	if strings.Count(str, ".") > 1 {
		lastDot := strings.LastIndex(str, ".")
		str = strings.ReplaceAll(str[:lastDot], ".", "") + str[lastDot:]
	}

	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return fmt.Errorf("invalid float value: %s", string(b))
	}

	*f = FlexFloat(val)
	return nil
}

func (f FlexFloat) MarshalJSON() ([]byte, error) {
	// Let's marshal it as a standard float, so the frontend can format it natively
	return json.Marshal(float64(f))
}

func (f FlexFloat) Float64() float64 {
	return float64(f)
}
