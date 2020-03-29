package main

import (
    "encoding/binary"
    "fmt"
    _ "fmt"
    "io/ioutil"
    "math/rand"
    "net"
    "net/http"
    _ "net/http"
    "net/url"
    "strconv"
    "torrentProxy/src"
)

type Peer struct {
    Ip net.IP
    Port uint16
}

type TorrentResponse struct {
    Interval int
    MinInterval int
    Peers []Peer
}

func main()  {
    bencode := src.BencodeTorrent{}
    bencode = src.GetBencodeStruct("/home/roman/golang/testfiles/2.torrent")
    //fmt.Println(bencode.InfoHash)
    base, err := url.Parse(bencode.Announce)
    if err != nil {
       fmt.Println("error")
       return
    }
    peerId := make([]byte, 20)
    letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    for i := range peerId {
        peerId[i] = letters[rand.Intn(len(letters))]
    }
    params := url.Values{
        "info_hash":  []string{string(bencode.InfoHash[:])},
        "peer_id":    []string{string(peerId[:])},
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
    ioutil.WriteFile("/home/roman/golang/testfiles/4.torrent", body, 0644)
    serverResp, err := ioutil.ReadFile("/home/roman/golang/testfiles/4.torrent")
    d := src.New(serverResp)
    t := d.Decode()
    response := TorrentResponse{}
    for i, val := range t.(map[string]interface{}) {
        switch i {
        case "interval":
            response.Interval = val.(int)
        case "min interval":
            response.MinInterval = val.(int)
        case "peers":
            byteStr := []byte(val.(string))
            peersCount := len(byteStr)/6
            for k:=0;k < peersCount ; k++  {
               offset := k * 6
               response.Peers = append(
                   response.Peers,
                   Peer{net.IP(byteStr[offset : offset + 4]), binary.BigEndian.Uint16(byteStr[offset+4 : offset+6])})
            }

        }
    }
    fmt.Printf("%s", response.Peers[0].Port)
}