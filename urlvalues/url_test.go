package urlvalues_test

import (
	"errors"
	"math"
	"net/url"
	"time"

	"github.com/gideaworx/go-encoding/urlvalues"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	BoolVal       bool       `url:"bo"`
	ByteVal       byte       `url:"by"`
	Complex64Val  complex64  `url:"c64"`
	Complex128Val complex128 `url:"c128"`
	Float32Val    float32    `url:"f32"`
	Float64Val    float64    `url:"f64"`
	IntVal        int        `url:"i"`
	Int8Val       int8       `url:"i8"`
	Int16Val      int16      `url:"i16"`
	Int32Val      int32      `url:"i32"`
	Int64Val      int64      `url:"i64"`
	RuneVal       rune       `url:"r"`
	StringVal     string     `url:"str"`
	UintVal       uint       `url:"ui"`
	Uint8Val      uint8      `url:"ui8"`
	Uint16Val     uint16     `url:"ui16"`
	Uint32Val     uint32     `url:"ui32"`
	Uint64Val     uint64     `url:"ui64"`
	TimeVal       time.Time  `url:"t"`
	ArrayVal      [3]int     `url:"a"`
	SliceVal      []string   `url:"s"`
	StringPtr     *string    `url:"strp"`
	SlicePtr      *[]int     `url:"sp"`
	SliceOfPtrs   []*string  `url:"sop"`
	ErrorVal      error      `url:"e"`
	OmitEmptyVal  int        `url:"o,omitempty"`
	SkipVal       bool       `url:"-"`
	NoTags        int
	unexported    string
}

var _ = Describe("URL Values Marshaling and Unmarshaling", func() {
	expectedEverythingVals := url.Values{}
	expectedEverythingVals.Set("bo", "false")
	expectedEverythingVals.Set("by", "97")
	expectedEverythingVals.Set("c64", "(0+2i)")
	expectedEverythingVals.Set("c128", "(3+598i)")
	expectedEverythingVals.Set("f32", "3.5")
	expectedEverythingVals.Set("f64", "8589934592")
	expectedEverythingVals.Set("i", "1")
	expectedEverythingVals.Set("i8", "8")
	expectedEverythingVals.Set("i16", "16")
	expectedEverythingVals.Set("i32", "32")
	expectedEverythingVals.Set("i64", "64")
	expectedEverythingVals.Set("r", "128556")
	expectedEverythingVals.Set("str", "string")
	expectedEverythingVals.Set("ui", "2")
	expectedEverythingVals.Set("ui8", "9")
	expectedEverythingVals.Set("ui16", "17")
	expectedEverythingVals.Set("ui32", "33")
	expectedEverythingVals.Set("ui64", "65")
	expectedEverythingVals.Set("t", "2021-02-18T09:56:00Z")
	expectedEverythingVals.Add("a", "1")
	expectedEverythingVals.Add("a", "2")
	expectedEverythingVals.Add("a", "3")
	expectedEverythingVals.Add("s", "x")
	expectedEverythingVals.Add("s", "y")
	expectedEverythingVals.Add("s", "z")
	expectedEverythingVals.Add("sp", "-3")
	expectedEverythingVals.Add("sp", "-2")
	expectedEverythingVals.Add("sp", "-1")
	expectedEverythingVals.Add("sop", "ptr")
	expectedEverythingVals.Add("sop", "ptr")
	expectedEverythingVals.Add("NoTags", "0")
	expectedEverythingVals.Set("e", "some error")

	strPtr := "ptr"
	slicePtr := []int{-3, -2, -1}

	var a aBitOfEverythingValid
	BeforeEach(func() {
		a = aBitOfEverythingValid{
			unexported:    "hi",
			ByteVal:       'a',
			Complex64Val:  2i,
			Complex128Val: 3 + 598i,
			Float32Val:    3.5,
			Float64Val:    math.Pow(2, 33),
			IntVal:        1,
			Int8Val:       8,
			Int16Val:      16,
			Int32Val:      32,
			Int64Val:      64,
			RuneVal:       '😬',
			StringVal:     "string",
			UintVal:       2,
			Uint8Val:      9,
			Uint16Val:     17,
			Uint32Val:     33,
			Uint64Val:     65,
			TimeVal:       time.Unix(staticTimestamp, 0).UTC(),
			ArrayVal:      [3]int{1, 2, 3},
			SliceVal:      []string{"x", "y", "z"},
			SlicePtr:      &slicePtr,
			SliceOfPtrs:   []*string{&strPtr, &strPtr},
			ErrorVal:      errors.New("some error"),
			SkipVal:       true,
		}
	})

	Describe("Marshaling", func() {
		Describe("Default", func() {
			It("marshals a struct correctly", func() {
				vals, err := urlvalues.MarshalURLValues(a)
				Expect(err).NotTo(HaveOccurred())
				Expect(vals.Encode()).To(Equal(expectedEverythingVals.Encode()))
			})

			It("marshals a struct pointer correctly", func() {
				vals, err := urlvalues.MarshalURLValues(&a)
				Expect(err).NotTo(HaveOccurred())
				Expect(vals.Encode()).To(Equal(expectedEverythingVals.Encode()))
			})

			It("marshals a map correctly", func() {
				m := map[string]any{
					"a": 1,
					"b": "test",
					"c": []bool{false, true, false},
				}

				v := url.Values{}
				v.Set("a", "1")
				v.Set("b", "test")
				v.Add("c", "false")
				v.Add("c", "true")
				v.Add("c", "false")

				vals, err := urlvalues.MarshalURLValues(m)
				Expect(err).NotTo(HaveOccurred())
				Expect(vals.Encode()).To(Equal(v.Encode()))
			})
		})

		Describe("URLValuesMarshaller", func() {
			It("marshals a URLValuesMarshaler correctly", func() {
				x := url.Values{}
				x.Add("a", "1")

				c := custom{v: x}

				vals, err := urlvalues.MarshalURLValues(c)
				Expect(err).NotTo(HaveOccurred())
				Expect(vals.Encode()).To(Equal(x.Encode()))
			})
		})

		Describe("Error Conditions", func() {
			It("fails on nil", func() {
				_, err := urlvalues.MarshalURLValues(nil)
				Expect(err).To(HaveOccurred())
			})

			It("fails on bool", func() {
				_, err := urlvalues.MarshalURLValues(false)
				Expect(err).To(HaveOccurred())
			})

			It("fails on num", func() {
				_, err := urlvalues.MarshalURLValues(3)
				Expect(err).To(HaveOccurred())
			})

			It("fails on string", func() {
				_, err := urlvalues.MarshalURLValues("")
				Expect(err).To(HaveOccurred())
			})

			It("fails on slice", func() {
				_, err := urlvalues.MarshalURLValues([]int{1, 2, 3})
				Expect(err).To(HaveOccurred())
			})

			It("fails on array", func() {
				_, err := urlvalues.MarshalURLValues([2]bool{true, true})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Unmarshaling", func() {
		Describe("Default", func() {
			var b aBitOfEverythingValid

			BeforeEach(func() {
				b = aBitOfEverythingValid{}
			})

			It("Unmarshals a struct", func() {
				a.unexported = ""
				a.SkipVal = false

				err := urlvalues.UnmarshalURLValues(expectedEverythingVals, &b)
				Expect(err).NotTo(HaveOccurred())

				Expect(b).To(BeEquivalentTo(a))
			})

			It("Unmarshals a map", func() {
				v := url.Values{}
				v.Set("bool", "true")
				v.Set("real", "32")
				v.Set("str", "foo")
				v.Set("complex", "(1.2+3.78i)")
				v.Set("time", "2021-02-18T09:56:00Z")
				v.Add("numslice", "1.2")
				v.Add("numslice", "-7")
				v.Add("numslice", "44.3e7")

				m := make(map[string]any)
				err := urlvalues.UnmarshalURLValues(v, &m)
				Expect(err).NotTo(HaveOccurred())
				Expect(m).To(HaveLen(6))
				Expect(m["bool"]).To(Equal(true))
				Expect(m["real"]).To(Equal(float64(32)))
				Expect(m["str"]).To(Equal("foo"))
				Expect(m["complex"]).To(Equal(complex128(1.2 + 3.78i)))
				Expect(m["time"]).To(Equal(time.Unix(staticTimestamp, 0).UTC()))
				Expect(m["numslice"]).To(BeEquivalentTo([]any{float64(1.2), float64(-7), float64(44.3e7)}))
			})
		})

		Describe("Custom", func() {
			It("unmarshals a URLValuesUnmarshaller correctly", func() {
				var c custom

				vals := url.Values{}
				vals.Set("a", "b")

				err := urlvalues.UnmarshalURLValues(vals, &c)
				Expect(err).NotTo(HaveOccurred())
				Expect(c.v).To(Equal(vals))
			})
		})

		Describe("Error conditions", func() {
			vals := url.Values{}
			vals.Set("a", "b")

			It("fails on nil", func() {
				Expect(urlvalues.UnmarshalURLValues(vals, nil)).To(HaveOccurred())
			})

			It("fails on a non-pointer struct", func() {
				Expect(urlvalues.UnmarshalURLValues(vals, custom{})).To(HaveOccurred())
			})

			It("fails on a non-pointer map", func() {
				Expect(urlvalues.UnmarshalURLValues(vals, map[string]any{})).To(HaveOccurred())
			})

			It("fails on a non-struct pointer", func() {
				var s string
				Expect(urlvalues.UnmarshalURLValues(vals, &s)).To(HaveOccurred())
			})
		})
	})
})
