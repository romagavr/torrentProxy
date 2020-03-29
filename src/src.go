package src

import (
    "bytes"
    "crypto/sha1"
    "encoding/binary"
    "fmt"
    "io/ioutil"
    "math/rand"
    "net"
    "net/http"
    "net/url"
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
        key := d.decodeString()
        var f bool = key == "info"
        var t int
        if f {
             t = d.position
        }
        dic[key] = d.Decode()
        if f {
           dic["infoHash"] = sha1.Sum(d.source[t : d.position])
        }
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

type bencodeInfo struct {
    Pieces [][]byte
    PieceLength int
    Length int
    Name string
} 

type BencodeTorrent struct {
    Announce string
    AnnounceList []string
    InfoHash [20]uint8
    Info bencodeInfo
    PeerId [20]byte
}

func (bencode *BencodeTorrent) GetBencodeStruct(data []byte) {
    t := New(data).Decode()
    bencode.PeerId = CreatePeerId()
    for i, val := range t.(map[string]interface{}) {
        switch i {
            case "announce":
                bencode.Announce = val.(string)
            case "announce-list":
                for _, value := range val.([]interface{}) {
                   for _, vll := range value.([]interface{}) {
                    bencode.AnnounceList = append(bencode.AnnounceList, vll.(string))
                    }
                }
            case "infoHash":
                bencode.InfoHash = val.([20]uint8)
            case "info":
                for j, value := range val.(map[string]interface{}) {                  
                    switch j {
                        case "pieces":
                            byts := []byte(value.(string))
                            bencode.Info.Pieces = make([][]byte, len(byts)/20)
                            for k:=0; k < len(byts)/20; k++ {
                               copy(bencode.Info.Pieces[k], byts[k*20:k*20+20])
                            }
                        case "piece length":
                            bencode.Info.PieceLength = value.(int)
                        case "length":
                            bencode.Info.Length = value.(int)
                        case "name":
                            bencode.Info.Name = value.(string)
                    }
                }
        }
    }
}

func CreatePeerId() [20]byte {
    var peerId [20]byte
    letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    for i := range peerId {
        peerId[i] = letters[rand.Intn(len(letters))]
    }
    return peerId
}

type Peer struct {
    Ip net.IP
    Port uint16
}

type TorrentResponse struct {
    Interval int
    MinInterval int
    Peers []Peer
}

func (bencode *BencodeTorrent) GetTorrentRequest() TorrentResponse {
    base, err := url.Parse(bencode.Announce)
    if err != nil {
        fmt.Println("error")
    }
    params := url.Values{
        "info_hash":  []string{string(bencode.InfoHash[:])},
        "peer_id":    []string{string(bencode.PeerId[:])},
        "port":       []string{strconv.Itoa(6969)},
        "uploaded":   []string{"0"},
        "downloaded": []string{"0"},
        "compact":    []string{"1"},
        "left":       []string{strconv.Itoa(bencode.Info.Length)},
    }
    base.RawQuery = params.Encode()
    resp, err := http.Get(base.String())
    //defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    return parseTorrentResponse(New(body).Decode())
}

func parseTorrentResponse(data interface{}) TorrentResponse {
    response := TorrentResponse{}
    for i, val := range data.(map[string]interface{}) {
        switch i {
            case "interval":
                response.Interval = val.(int)
            case "min interval":
                response.MinInterval = val.(int)
            case "peers":
                byts := []byte(val.(string))
                response.Peers = make([]Peer, len(byts)/6)
                for k:=0; k < len(byts)/6; k++ {
                    response.Peers[k] = Peer{net.IP(byts[6*k:6*k+4]), binary.BigEndian.Uint16(byts[6*k+4 : 6*k+6])}
                }
        }
    }
    return response
}
