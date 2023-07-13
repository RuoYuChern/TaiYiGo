package indicators

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

var (
	flZero = *big.NewFloat(0)

	// NaN == Not a Number
	NaN = NewDecimal(math.NaN())

	// ZERO == 0
	ZERO = NewFromString("0")

	// ONE == 1
	ONE = NewFromString("1")

	// TEN == 10
	TEN = NewFromString("10")

	// MarshalQuoted - can toggle this to true to marshal values as strings
	MarshalQuoted = false
)

type Decimal struct {
	fl *big.Float
}

func NewDecimal(val float64) Decimal {
	var fl *big.Float

	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case big.ErrNaN:
				fl = nil
			}
		}
	}()

	fl = big.NewFloat(val)

	return Decimal{
		fl: fl,
	}
}

// NewFromString creates a new Decimal type from a string value.
func NewFromString(str string) Decimal {
	bfl := big.NewFloat(0)

	if _, _, err := bfl.Parse(str, 10); err != nil {
		return NaN
	}

	return Decimal{bfl}
}

// NewFromInt creates a new Decimal type from an int value
func NewFromInt(dec int) Decimal {
	return Decimal{big.NewFloat(float64(dec))}
}

// Add adds a decimal instance to another Decimal instance.
func (d Decimal) Add(addend Decimal) Decimal {
	return nanGuard(func() Decimal {
		return Decimal{d.cpy().Add(d.fl, addend.fl)}
	}, d, addend)
}

// Sub subtracts another decimal instance from this Decimal instance.
func (d Decimal) Sub(subtrahend Decimal) Decimal {
	return nanGuard(func() Decimal {
		return Decimal{d.cpy().Sub(d.fl, subtrahend.fl)}
	}, d, subtrahend)
}

// Mul multiplies another decimal instance with this Decimal instance.
func (d Decimal) Mul(factor Decimal) Decimal {
	return nanGuard(func() Decimal {
		return Decimal{d.cpy().Mul(d.fl, factor.fl)}
	}, d, factor)
}

// Div divides this Decimal by the denominator passed.
func (d Decimal) Div(denominator Decimal) Decimal {
	return nanGuard(func() Decimal {
		return Decimal{d.cpy().Quo(d.fl, denominator.fl)}
	}, d, denominator)
}

// Frac returns another Decimal instance representing this Decimal multiplied by the
// provided float.
func (d Decimal) Frac(f float64) Decimal {
	fractionFactor := NewDecimal(f)

	return nanGuard(func() Decimal {
		return d.Mul(NewDecimal(f))
	}, d, fractionFactor)
}

// Neg returns this Decimal multiplied by -1.
func (d Decimal) Neg() Decimal {
	return nanGuard(func() Decimal {
		return d.Mul(NewDecimal(-1))
	}, d)
}

// Abs returns the absolute value of this Decimal
func (d Decimal) Abs() Decimal {
	if d.LT(ZERO) {
		return d.Mul(ONE.Neg())
	}

	return d
}

// Pow returns the decimal to the inputted power
func (d Decimal) Pow(exp int) Decimal {
	return nanGuard(func() Decimal {
		if exp == 0 {
			return ONE
		}

		x := Decimal{d.cpy()}

		for i := 1; i < exp; i++ {
			x = x.Mul(d)
		}

		return x
	}, d)
}

// Sqrt returns the decimal's square root
func (d Decimal) Sqrt() Decimal {
	return nanGuard(func() Decimal {
		return Decimal{d.cpy().Sqrt(d.cpy())}
	}, d)
}

// EQ returns true if this Decimal exactly equals the provided decimal.
func (d Decimal) EQ(other Decimal) bool {
	if anyNan(d, other) {
		return false
	}

	return d.Cmp(other) == 0
}

// LT returns true if this decimal is less than the provided decimal.
func (d Decimal) LT(other Decimal) bool {
	if anyNan(d, other) {
		return false
	}

	return d.Cmp(other) < 0
}

// LTE returns true if this decimal is less or equal to the provided decimal.
func (d Decimal) LTE(other Decimal) bool {
	if anyNan(d, other) {
		return false
	}

	return d.Cmp(other) <= 0
}

// GT returns true if this decimal is greater than the provided decimal.
func (d Decimal) GT(other Decimal) bool {
	if anyNan(d, other) {
		return false
	}

	return d.Cmp(other) > 0
}

// GTE returns true if this decimal is greater than or equal to the provided decimal.
func (d Decimal) GTE(other Decimal) bool {
	if anyNan(d, other) {
		return false
	}

	return d.Cmp(other) >= 0
}

// Cmp will return 1 if this decimal is greater than the provided, 0 if they are the same, and -1 if it is less.
func (d Decimal) Cmp(other Decimal) int {
	if anyNan(d, other) {
		return 0
	}

	return d.fl.Cmp(other.fl)
}

// Float will return this Decimal as a float value.
// Note that there may be some loss of precision in this operation.
func (d Decimal) Float() float64 {
	if d.NaN() {
		return math.NaN()
	}

	f, _ := d.fl.Float64()
	return f
}

func (d Decimal) FormatFloat(decimal int) float64 {
	if d.NaN() {
		return math.NaN()
	}
	f, _ := d.fl.Float64()
	dd := float64(1)
	if decimal > 0 {
		dd = math.Pow10(decimal)
	}
	res := strconv.FormatFloat(math.Trunc(f*dd)/dd, 'f', -1, 64)
	fv, _ := strconv.ParseFloat(res, 64)
	return fv
}

// Zero will return true if this Decimal is equal to 0.
// Deprecated: Use IsZero instead
func (d Decimal) Zero() bool {
	return d.IsZero()
}

// NaN returns true if the underlying is not a valid number
func (d Decimal) NaN() bool {
	return d.fl == nil
}

// IsZero will return true if this Decimal is equal to 0.
func (d Decimal) IsZero() bool {
	if d.NaN() {
		return false
	}

	return d.fl == nil || d.fl.Cmp(&flZero) == 0
}

func (d Decimal) String() string {
	if d.NaN() {
		return "NaN"
	}

	if d.fl == nil {
		d.fl = new(big.Float)
	}

	return d.fl.String()
}

// FormattedString returns the string value of the number to the requested precision
func (d Decimal) FormattedString(places int) string {
	if d.NaN() {
		return d.String()
	}

	format := "%." + fmt.Sprint(places) + "f"
	fl := d.Float()
	return fmt.Sprintf(format, fl)
}

func anyNan(decimals ...Decimal) bool {
	for _, decimal := range decimals {
		if decimal.NaN() {
			return true
		}
	}

	return false
}

func nanGuard(yeildFunc func() Decimal, decimals ...Decimal) Decimal {
	if anyNan(decimals...) {
		return NaN
	}

	return yeildFunc()
}

func (d Decimal) cpy() *big.Float {
	cpy := new(big.Float)
	return cpy.Copy(d.fl)
}
