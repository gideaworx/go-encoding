package urlvalues_test

import (
	"errors"
	"math"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.gideaworx.io/go-encoding/urlvalues"
)

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
	expectedEverythingVals.Add("j", "hello, world")
	expectedEverythingVals.Add("jnc", "1X2X3")
	expectedEverythingVals.Add("jmc", "1.1a,b,c2.2a,b,c3.3")
	expectedEverythingVals.Add("jes", "a")
	expectedEverythingVals.Add("jes", "b")
	expectedEverythingVals.Add("NoTags", "0")
	expectedEverythingVals.Set("e", "some error")

	strPtr := "ptr"
	slicePtr := []int{-3, -2, -1}

	boolTest := boolFormat{
		false, false, false, false, false, false, false,
		true, true, true, true, true, true, true,
	}

	durationTest := durationFormat{
		testDur, testDur, testDur, testDur, testDur, testDur, testDur,
	}

	numberTest := fcFormat{
		xf32, xf32, xf32, xf32, xf32,
		nf32, nf32, nf32, nf32, nf32,
		xf64, xf64, xf64, xf64, xf64,
		nf64, nf64, nf64, nf64, nf64,
		xc64, xc64, xc64, xc64, xc64,
		nc64, nc64, nc64, nc64, nc64,
		xc128, xc128, xc128, xc128, xc128,
		nc128, nc128, nc128, nc128, nc128,
	}

	encodedBool :=
		"FalseCamel=False&FalseDefault=false&FalseInt=0&FalseLower=false&FalseShort=F&FalseShortLower=f&FalseUpper=FALSE&" +
			"TrueCamel=True&TrueDefault=true&TrueInt=1&TrueLower=true&TrueShort=T&TrueShortLower=t&TrueUpper=TRUE"

	encodedDur := "d=1m0s&h=0&m=1&ms=60000&ns=60000000000&s=60&us=60000000"

	encodedNums := "C128MaxE=%281.7976931348623157E%2B308%2B1.7976931348623157E%2B308i%29&C128MaxG=%281.7976931348623157E%2B308%2B1.7976931348623157E%2B308i%29&C128Maxe=%281.7976931348623157e%2B308%2B1.7976931348623157e%2B308i%29&C128Maxf=%28179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000%2B179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000i%29&C128Maxg=%281.7976931348623157e%2B308%2B1.7976931348623157e%2B308i%29&C128MinE=%285E-324%2B5E-324i%29&C128MinG=%285E-324%2B5E-324i%29&C128Mine=%285e-324%2B5e-324i%29&C128Minf=%280.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005%2B0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005i%29&C128Ming=%285e-324%2B5e-324i%29&C64MaxE=%283.4028235E%2B38%2B3.4028235E%2B38i%29&C64MaxG=%283.4028235E%2B38%2B3.4028235E%2B38i%29&C64Maxe=%283.4028235e%2B38%2B3.4028235e%2B38i%29&C64Maxf=%28340282350000000000000000000000000000000%2B340282350000000000000000000000000000000i%29&C64Maxg=%283.4028235e%2B38%2B3.4028235e%2B38i%29&C64MinE=%281E-45%2B1E-45i%29&C64MinG=%281E-45%2B1E-45i%29&C64Mine=%281e-45%2B1e-45i%29&C64Minf=%280.000000000000000000000000000000000000000000001%2B0.000000000000000000000000000000000000000000001i%29&C64Ming=%281e-45%2B1e-45i%29&F32MaxE=3.4028235E%2B38&F32MaxG=3.4028235E%2B38&F32Maxe=3.4028235e%2B38&F32Maxf=340282350000000000000000000000000000000&F32Maxg=3.4028235e%2B38&F32MinE=1E-45&F32MinG=1E-45&F32Mine=1e-45&F32Minf=0.000000000000000000000000000000000000000000001&F32Ming=1e-45&F64MaxE=1.7976931348623157E%2B308&F64MaxG=1.7976931348623157E%2B308&F64Maxe=1.7976931348623157e%2B308&F64Maxf=179769313486231570000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000&F64Maxg=1.7976931348623157e%2B308&F64MinE=5E-324&F64MinG=5E-324&F64Mine=5e-324&F64Minf=0.000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005&F64Ming=5e-324"

	var a aBitOfEverythingValid
	BeforeEach(func() {
		a = aBitOfEverythingValid{
			unexported:     "hi",
			ByteVal:        'a',
			Complex64Val:   2i,
			Complex128Val:  3 + 598i,
			Float32Val:     3.5,
			Float64Val:     math.Pow(2, 33),
			IntVal:         1,
			Int8Val:        8,
			Int16Val:       16,
			Int32Val:       32,
			Int64Val:       64,
			RuneVal:        '😬',
			StringVal:      "string",
			UintVal:        2,
			Uint8Val:       9,
			Uint16Val:      17,
			Uint32Val:      33,
			Uint64Val:      65,
			TimeVal:        time.Unix(staticTimestamp, 0).UTC(),
			ArrayVal:       [3]int{1, 2, 3},
			SliceVal:       []string{"x", "y", "z"},
			SlicePtr:       &slicePtr,
			SliceOfPtrs:    []*string{&strPtr, &strPtr},
			ErrorVal:       errors.New("some error"),
			JoinedSlice:    []string{"hello", "world"},
			JoinNoComma:    []int{1, 2, 3},
			JoinMultiComma: []float64{1.1, 2.2, 3.3},
			JoinEmptyStr:   []string{"a", "b"},
			SkipVal:        true,
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

		Describe("Formats", func() {
			It("Marshals Bools properly", func() {
				vals, err := urlvalues.MarshalURLValues(boolTest)

				Expect(err).NotTo(HaveOccurred())
				Expect(vals.Encode()).To(Equal(encodedBool))
			})

			It("Marshals Durations properly", func() {
				vals, err := urlvalues.MarshalURLValues(durationTest)

				Expect(err).NotTo(HaveOccurred())
				Expect(vals.Encode()).To(Equal(encodedDur))
			})

			It("Marshals Floats and Complexes properly", func() {
				vals, err := urlvalues.MarshalURLValues(numberTest)

				Expect(err).NotTo(HaveOccurred())
				Expect(vals.Encode()).To(Equal(encodedNums))
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

		Describe("Formatting", func() {
			It("Unmarshals Bools properly", func() {
				vals, err := url.ParseQuery(encodedBool)
				Expect(err).NotTo(HaveOccurred())

				var b boolFormat
				err = urlvalues.UnmarshalURLValues(vals, &b)

				Expect(err).NotTo(HaveOccurred())
				Expect(b).To(BeEquivalentTo(boolTest))
			})

			It("Unmarshals Durations properly", func() {
				vals, err := url.ParseQuery(encodedDur)
				Expect(err).NotTo(HaveOccurred())

				var d durationFormat
				err = urlvalues.UnmarshalURLValues(vals, &d)

				// reset it here since it marshals to 0, since it was set to 1 minute
				durationTest.Hour = 0

				Expect(err).NotTo(HaveOccurred())
				Expect(d).To(BeEquivalentTo(durationTest))
			})

			It("Unmarshals floats and complexes as best as it can", func() {
				s := struct {
					D  float64    `url:"d" urlformat:"G"`
					F  float32    `url:"f" urlformat:"f"`
					BC complex128 `url:"bc" urlformat:"g"`
					SC complex64  `url:"sc" urlformat:"f"`
				}{
					xf64, 38_000_000_004.32, complex(nf64, nf64), (3.4 + 2.8e-7i),
				}

				x := struct {
					D  float64    `url:"d" urlformat:"E"`
					F  float32    `url:"f" urlformat:"f"`
					BC complex128 `url:"bc" urlformat:"g"`
					SC complex64  `url:"sc" urlformat:"f"`
				}{}

				v := url.Values{}
				v.Add("d", "1.7976931348623157E+308")
				v.Add("f", "38000000004.32")
				v.Add("bc", "(5e-324+5e-324i)")
				v.Add("sc", "(3.4+0.00000028i)")

				err := urlvalues.UnmarshalURLValues(v, &x)
				Expect(err).NotTo(HaveOccurred())
				Expect(x).To(BeEquivalentTo(s))
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
