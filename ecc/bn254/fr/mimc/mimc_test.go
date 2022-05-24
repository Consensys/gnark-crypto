package mimc

// import (
// 	"testing"
// )

// func TestMimc(t *testing.T) {

// // Expected result from ethereum
// var data [3]fr.Element
// data[0].SetString("10909369219534740878285360918369814291778422174980871969149168794639722256599")
// data[1].SetString("3811523387212735178398974960485340561880938762308498768570292593755555588442")
// data[2].SetString("21761276089180230617904476026690048826689721630933485969915548849196498965166")

// h := NewMiMC()
// h.Write(data[0].Marshal())
// h.Write(data[1].Marshal())
// h.Write(data[2].Marshal())

// r := h.Sum(nil)

// var b big.Int
// b.SetBytes(r)
// fmt.Printf("%s\n", b.String())

//-------

// h := NewMiMC("mimc")
// var a [3]fr.Element
// a[0].SetRandom()
// a[1].SetRandom()
// a[2].SetRandom()
// fmt.Printf("%s\n", a[0].String())
// fmt.Printf("%s\n", a[1].String())
// fmt.Printf("%s\n", a[2].String())
// fmt.Println("")
// h.Write(a[0].Marshal())
// h.Write(a[1].Marshal())
// h.Write(a[2].Marshal())
// var a fr.Element
// a.SetUint64(2323)
// h.Write(a.Marshal())
// r := h.Sum(nil)
// var br big.Int
// br.SetBytes(r)
// fmt.Printf("%s\n", br.String())
//_h := h.(*digest)

// //var h1, h2, h3 fr.Element
// var h1, h2 fr.Element
// h1.SetString("948723")
// h2.SetString("236878")
// // h3.SetString("283")
// _h.data = append(_h.data, h1.Marshal()...)
// _h.data = append(_h.data, h2.Marshal()...)
// // _h.data = append(_h.data, h3.Marshal()...)

// _h.checksum()
// fmt.Printf("%s\n", _h.h.String())
// }
