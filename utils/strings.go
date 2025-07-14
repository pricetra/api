package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-jet/jet/v2/postgres"
)

func BuildFullTableName(col postgres.Column) string {
	return fmt.Sprintf(`%s.%s`, col.TableName(), col.Name())
}

func BuildFullTableNameHyphen(col postgres.Column) string {
	return fmt.Sprintf(`%s_%s`, col.TableName(), col.Name())
}

func ToPostgresArray[T comparable](arr []T) string {
	str_arr := make([]string, len(arr))
	for i, val := range arr {
		str_arr[i] = fmt.Sprintf("%v", val)
	}
	return fmt.Sprintf("{%s}", strings.Join(str_arr, ","))
}

func PostgresArrayToStrArray(p_array string) []string {
	return strings.Split(strings.Trim(p_array, "{}"), ",")
}

func PostgresArrayToIntArray(p_array string) []int {
	var err error
	str_array := strings.Split(strings.Trim(p_array, "{}"), ",")
	int_array := make([]int, len(str_array))
	for i, v := range str_array {
		int_array[i], err = strconv.Atoi(v)
		if err != nil {
			int_array[i] = 0
		}
	}
	return int_array
}

func IsValidBase64Image(base64 string) bool {
	if len(base64) > 50 {
		base64 = base64[:50]
	}
	return strings.HasPrefix(base64, "data:image/") && strings.Contains(base64, ";base64,")
}

func buildWeightRegex(unitMap map[string]string) *regexp.Regexp {
	// Extract all keys from the normalization map
	units := make([]string, 0, len(unitMap))
	for key := range unitMap {
		// Escape special regex chars (e.g., "fl oz")
		escaped := regexp.QuoteMeta(key)
		units = append(units, escaped)
	}

	// Sort by descending length to prioritize multi-word units like "fluid ounce"
	sort.Slice(units, func(i, j int) bool {
		return len(units[i]) > len(units[j])
	})

	// Build the unit alternation pattern
	unitPattern := strings.Join(units, "|")

	// Final regex (numeric part + optional space + unit)
	regexPattern := fmt.Sprintf(`(?i)\b([\d.]+)\s*(%s)\b`, unitPattern)

	return regexp.MustCompile(regexPattern)
}

var weightRegex = buildWeightRegex(unitNormalization)

var unitNormalization = map[string]string{
	// Weight
	"oz": "oz", "ounce": "oz", "ounces": "oz",
	"lb": "lb", "lbs": "lb", "pound": "lb", "pounds": "lb",
	"g": "g", "gram": "g", "grams": "g",
	"kg": "kg", "kilogram": "kg", "kilograms": "kg",
	"mg": "mg", "milligram": "mg", "milligrams": "mg",

	// Volume
	"fl oz": "fl oz", "floz": "fl oz", "fluid ounce": "fl oz", "fluid ounces": "fl oz",
	"pt": "pt", "pint": "pt", "pints": "pt",
	"qt": "qt", "quart": "qt", "quarts": "qt",
	"gal": "gal", "gallon": "gal", "gallons": "gal",
	"ml": "ml", "milliliter": "ml", "milliliters": "ml",
	"l": "l", "liter": "l", "liters": "l", "litre": "l", "litres": "l",
	"cup": "cup", "cups": "cup",
	"tbsp": "tbsp", "tablespoon": "tbsp", "tablespoons": "tbsp",
	"tsp": "tsp", "teaspoon": "tsp", "teaspoons": "tsp",

	// Other
	"ct": "ct", "count": "ct", "pieces": "ct", "pcs": "ct",
	"pack": "pack", "packs": "pack", "pkg": "pack", "pkgs": "pack",
	"serving": "serving", "servings": "serving",
}
func ParseWeight(raw_weight string) *string {
	raw_weight = strings.ToLower(raw_weight)
	matches := weightRegex.FindStringSubmatch(raw_weight)
	if len(matches) < 3 {
		return nil
	}

	valueStr := matches[1]
	unit := matches[2]

	// Normalize unit
	if nunit, ok := unitNormalization[unit]; ok {
		unit = nunit
	}

	// Validate numeric value
	if _, err := strconv.ParseFloat(valueStr, 64); err != nil {
		return nil
	}

	full_weight := fmt.Sprintf("%s %s", valueStr, unit)
	return &full_weight
}
