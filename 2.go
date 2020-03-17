package main

import (
    "fmt"
    "io/ioutil"
    "bytes"
    "strconv"
)

func check(e error) {
    if e != nil {
        fmt.Print("Error")
        panic(e)
    }
}

type Decoder struct {
    source   []byte
    position int
    char     byte
}

func (d *Decoder) readChar() {
    if d.position >= len(d.source) {
        d.char = 0
        return
    }
    d.char = d.source[d.position]
}

func New(source []byte) *Decoder {
    d := &Decoder{source: source}
    d.readChar()
    return d
}

func (d *Decoder) Decode() interface{} {
    switch d.char {
    case 'd':
        return d.decodeDictionary()
    case 'l':
        return d.decodeList()
    case 'i':
        return d.decodeInt()
    default:
        if isDigit(d.char) {
            return d.decodeString()
        }
        return nil
    }
}

func (d *Decoder) decodeInt() int {
    var i int
    e := d.position + bytes.IndexByte(d.source[d.position:], byte('e'))
    d.position++
    s := d.position
    d.position += e - s
    i, _ = strconv.Atoi(string(d.source[s:d.position]))
    d.incrementPosition(1)
    return i
}

func (d *Decoder) decodeList() []interface{} {
    l := []interface{}{}
    d.incrementPosition(1)
    for d.char != 'e' {
        l = append(l,d.Decode())
    }
    d.incrementPosition(1)
    return l
}

func (d *Decoder) decodeDictionary() map[string]interface{} {
    dic := map[string]interface{}{}
    d.incrementPosition(1)
    for d.char != 'e' {
        dic[d.decodeString()] = d.Decode()
    }
    d.incrementPosition(1)
    return dic
}

func (d *Decoder) decodeString() string {
    colon := d.position + bytes.IndexByte(d.source[d.position:], byte(':'))
    length, _ := strconv.Atoi(string(d.source[d.position:colon]))
    d.incrementPosition(colon - d.position + length + 1)
    return string(d.source[colon+1 : d.position])
}

func isDigit(b byte) bool {
    return b>='0' && b <= '9'
}

func (d *Decoder) incrementPosition(pos int) {
    d.position += pos
    d.readChar()
}

func main() {
    dat, err := ioutil.ReadFile("/home/roman/docker/golang/test3.torrent")
    check(err)
    //fmt.Printf("File contents: %s", dat)
    d := New([]byte(dat))
    t := d.Decode()
        var slc []string
    for _, j := range t {
        switch v := j.(type) {
        case string:
            slc = append(slc, v)
        case fmt.Stringer:
            slc = append(slc, v.String())
        default:
            slc = append(slc, fmt.Sprintf("%v", v))
        }
    }

    fmt.Println(slc)
    //fmt.Printf("%s", t["info"].(map[string]interface{}))
}
