package urlvalues_test

import (
	"math"
	"net/url"
	"time"
)

type custom struct {
	v url.Values
}

func (c custom) MarshalURLValues() (url.Values, error) {
	return c.v, nil
}

func (c *custom) UnmarshalURLValues(v url.Values) error {
	c.v = v
	return nil
}

const staticTimestamp = 1613642160

type aBitOfEverythingValid struct {
	BoolVal        bool       `url:"bo"`
	ByteVal        byte       `url:"by"`
	Complex64Val   complex64  `url:"c64"`
	Complex128Val  complex128 `url:"c128"`
	Float32Val     float32    `url:"f32"`
	Float64Val     float64    `url:"f64"`
	IntVal         int        `url:"i"`
	Int8Val        int8       `url:"i8"`
	Int16Val       int16      `url:"i16"`
	Int32Val       int32      `url:"i32"`
	Int64Val       int64      `url:"i64"`
	RuneVal        rune       `url:"r"`
	StringVal      string     `url:"str"`
	UintVal        uint       `url:"ui"`
	Uint8Val       uint8      `url:"ui8"`
	Uint16Val      uint16     `url:"ui16"`
	Uint32Val      uint32     `url:"ui32"`
	Uint64Val      uint64     `url:"ui64"`
	TimeVal        time.Time  `url:"t"`
	ArrayVal       [3]int     `url:"a"`
	SliceVal       []string   `url:"s"`
	StringPtr      *string    `url:"strp"`
	SlicePtr       *[]int     `url:"sp"`
	SliceOfPtrs    []*string  `url:"sop"`
	JoinedSlice    []string   `url:"j,join=', '"`
	JoinNoComma    []int      `url:"jnc,join='X'"`
	JoinMultiComma []float64  `url:"jmc,join='a,b,c'"`
	JoinEmptyStr   []string   `url:"jes,join=''"`
	ErrorVal       error      `url:"e"`
	OmitEmptyVal   int        `url:"o,omitempty"`
	SkipVal        bool       `url:"-"`
	NoTags         int
	unexported     string
}

type boolFormat struct {
	FalseShort      bool `urlformat:"short"`
	FalseShortLower bool `urlformat:"shortlower"`
	FalseUpper      bool `urlformat:"upper"`
	FalseCamel      bool `urlformat:"camel"`
	FalseLower      bool `urlformat:"lower"`
	FalseInt        bool `urlformat:"int"`
	FalseDefault    bool
	TrueShort       bool `urlformat:"short"`
	TrueShortLower  bool `urlformat:"shortlower"`
	TrueUpper       bool `urlformat:"upper"`
	TrueCamel       bool `urlformat:"camel"`
	TrueLower       bool `urlformat:"lower"`
	TrueInt         bool `urlformat:"int"`
	TrueDefault     bool
}

type durationFormat struct {
	Default     time.Duration `url:"d"`
	Nanosecond  time.Duration `url:"ns" urlformat:"int,ns"`
	Microsecond time.Duration `url:"us" urlformat:"int,us"`
	Millisecond time.Duration `url:"ms" urlformat:"int,ms"`
	Second      time.Duration `url:"s" urlformat:"int,s"`
	Minute      time.Duration `url:"m" urlformat:"int,m"`
	Hour        time.Duration `url:"h" urlformat:"int,h"`
}

type fcFormat struct {
	F32Maxe  float32    `urlformat:"e"`
	F32MaxE  float32    `urlformat:"E"`
	F32Maxf  float32    `urlformat:"f"`
	F32Maxg  float32    `urlformat:"g"`
	F32MaxG  float32    `urlformat:"G"`
	F32Mine  float32    `urlformat:"e"`
	F32MinE  float32    `urlformat:"E"`
	F32Minf  float32    `urlformat:"f"`
	F32Ming  float32    `urlformat:"g"`
	F32MinG  float32    `urlformat:"G"`
	F64Maxe  float64    `urlformat:"e"`
	F64MaxE  float64    `urlformat:"E"`
	F64Maxf  float64    `urlformat:"f"`
	F64Maxg  float64    `urlformat:"g"`
	F64MaxG  float64    `urlformat:"G"`
	F64Mine  float64    `urlformat:"e"`
	F64MinE  float64    `urlformat:"E"`
	F64Minf  float64    `urlformat:"f"`
	F64Ming  float64    `urlformat:"g"`
	F64MinG  float64    `urlformat:"G"`
	C64Maxe  complex64  `urlformat:"e"`
	C64MaxE  complex64  `urlformat:"E"`
	C64Maxf  complex64  `urlformat:"f"`
	C64Maxg  complex64  `urlformat:"g"`
	C64MaxG  complex64  `urlformat:"G"`
	C64Mine  complex64  `urlformat:"e"`
	C64MinE  complex64  `urlformat:"E"`
	C64Minf  complex64  `urlformat:"f"`
	C64Ming  complex64  `urlformat:"g"`
	C64MinG  complex64  `urlformat:"G"`
	C128Maxe complex128 `urlformat:"e"`
	C128MaxE complex128 `urlformat:"E"`
	C128Maxf complex128 `urlformat:"f"`
	C128Maxg complex128 `urlformat:"g"`
	C128MaxG complex128 `urlformat:"G"`
	C128Mine complex128 `urlformat:"e"`
	C128MinE complex128 `urlformat:"E"`
	C128Minf complex128 `urlformat:"f"`
	C128Ming complex128 `urlformat:"g"`
	C128MinG complex128 `urlformat:"G"`
}

const testDur time.Duration = time.Minute

const xf32 = math.MaxFloat32
const nf32 = math.SmallestNonzeroFloat32
const xf64 = math.MaxFloat64
const nf64 = math.SmallestNonzeroFloat64
const xc64 = complex(math.MaxFloat32, math.MaxFloat32)
const nc64 = complex(math.SmallestNonzeroFloat32, math.SmallestNonzeroFloat32)
const xc128 = complex(math.MaxFloat64, math.MaxFloat64)
const nc128 = complex(math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64)
