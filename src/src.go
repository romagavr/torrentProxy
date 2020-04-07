package src

import (
    "bufio"
    "bytes"
    "crypto/sha1"
    "encoding/binary"
    "fmt"
    "io"
    "io/ioutil"
    "math/rand"
    "net"
    "net/http"
    "net/url"
    "strconv"
    "time"
)


type BencodeTorrent struct {
    //Common fields
    Announce string
    Info Info

    //Optional fields
    AnnounceList AnnounceList
    CreationDate time.Time
    Comment string
    CreatedBy string
    Encoding string

    //related fields
    PeerId [20]byte
    InfoHash [20]uint8
}

type Info struct {
    Pieces [][]byte
    PieceLength int
    Name string

    //Optional fields
    Private int

    //SingleFileInfo
    Length int
    Md5Sum string

    //MultiFileInfo
    Files []File
}

type AnnounceList struct {
    Announces []string
}

type File struct {
    Length int
    Path []string
    Md5Sum string
}

func (an *AnnounceList) decodeList(r *bufio.Reader) []string {
    for peekChar(r) == "l" {
        discardChar(r)
        an.Announces = append(an.Announces,an.decodeList(r)...)
    }
    for peekChar(r) != "e" {
        an.Announces = append(an.Announces,decodeString(r))
    }
    discardChar(r)
    return an.Announces
}

func decodeFile(r *bufio.Reader) File {
    discardChar(r)
    f := File{}
    for peekChar(r) != "e" {
        switch decodeString(r) {
        case "length":
            f.Length = decodeInt(r)
        case "md5sum":
            f.Md5Sum = decodeString(r)
        case "path":
            discardChar(r) //discard l
            for peekChar(r) != "e" {
                f.Path = append(f.Path, decodeString(r))
            }
            discardChar(r) //discard e (l)
        }
    }
    discardChar(r) //discard e (d)
    return f
}

func decodeString(r *bufio.Reader) string {
    num, _ := r.ReadString(':')
    length, _ := strconv.Atoi(num[:len(num)-1])
    read, _ := r.Read(make([]byte, length))
    return string(read)
}

func decodeInt(r *bufio.Reader) int {
    num, _ := r.ReadString('e')
    n, _ := strconv.Atoi(num[1:len(num)-1])
    return n
}

func peekChar(r *bufio.Reader) string {
    var ch []byte
    var e error
    if ch, e = r.Peek(1); e != nil {
        panic(e)
    }
    return string(ch)
}

func readChar(r *bufio.Reader) string {
    var ch byte
    var e error
    if ch, e = r.ReadByte(); e != nil {
        panic(e)
    }
    return string(ch)
}

func discardChar(r *bufio.Reader) {
    if _, e := r.Discard(1); e != nil {
        panic(e)
    }
}

func (b *BencodeTorrent) decodeDictionary(r *bufio.Reader) {
    discardChar(r)
    for peekChar(r) != "e" {
        switch decodeString(r) {
            case "announce":
                b.Announce = decodeString(r)
            case "announce-list":
                b.AnnounceList.decodeList(r)
            case "comment":
                b.Comment = decodeString(r)
            case "creation date":
                b.CreationDate = time.Unix(int64(decodeInt(r)), 0)
            case "created by":
                b.CreatedBy = decodeString(r)
            case "encoding":
                b.Encoding = decodeString(r)
            case "info":
                b.decodeDictionary(r)
            case "pieces":
                str := decodeString(r)
                for k:=0; k < len(str)/20; k++ {
                    copy(b.Info.Pieces[k], str[k*20:k*20+20])
                }
            case "piece length":
                b.Info.PieceLength = decodeInt(r)
            case "private":
                b.Info.Private = decodeInt(r)
            case "length":
                b.Info.Length = decodeInt(r)
            case "name":
                b.Info.Name = decodeString(r)
            case "md5sum":
                b.Info.Name = decodeString(r)
            case "files":
                discardChar(r) //discard l
                for peekChar(r) == "d" {
                    b.Info.Files = append(b.Info.Files,decodeFile(r))
                }
                discardChar(r) //discard e (l)
        }
    }
    discardChar(r) //discard e (d)
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

func (bencode *BencodeTorrent) GetTorrentResponse() TorrentResponse {
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
