package main

import (
    "bytes"
    "encoding/binary"
    "fmt"
    _ "fmt"
    "io"
    "io/ioutil"
    "math/rand"
    "net"
    "net/http"
    _ "net/http"
    "net/url"
    "strconv"
    "time"
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
    // Parse bencodeed File
    bencode := src.BencodeTorrent{}
    bencode = src.GetBencodeStruct("/home/roman/golang/testfiles/9.torrent")
    // Create random PeerId
    peerId := make([]byte, 20)
    letters := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
    for i := range peerId {
        peerId[i] = letters[rand.Intn(len(letters))]
    }
    // Send request to server
    base, err := url.Parse(bencode.Announce)
    if err != nil {
       fmt.Println("error")
       return
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
    //fmt.Printf("%s", response.Peers[0].Port)

    //Format Bittorrent handshake
    var handShake []byte
    handShake = make([]byte, 68)
    handShake[0] = byte(len("BitTorrent protocol"))
    count := 1
    count += copy(handShake[count:], "BitTorrent protocol")
    count += copy(handShake[count:], make([]byte, 8))
    count += copy(handShake[count:], bencode.InfoHash[:])
    count += copy(handShake[count:], peerId[:])
    //fmt.Println(handShake)

    //Connection to Peer
    fmt.Printf("%s", response.Peers[6])
    dkf  := response.Peers[3].Ip.String() + ":" + strconv.Itoa(int(response.Peers[3].Port))
    //fmt.Println(dkf)
    conn, err := net.DialTimeout("tcp", dkf, 10 * time.Second)
    if err != nil {
        //fmt.Println("sdkdsdf")
       fmt.Println(err)
    }

    fmt.Println("Written:" + string(handShake))
    _, err = conn.Write(handShake)
    fmt.Println("Written:" + string(handShake))
    if err != nil {
       fmt.Println("Peer writting error")
    }
    fmt.Println("2323")
    hhh := make([]byte, 1)
        _, err = io.ReadFull(conn, hhh)
        if err != nil {
            fmt.Println(err)
        }
        fmt.Println(hhh)
    handshakeBuf := make([]byte, 48+int(hhh[0]))
    _, err = io.ReadFull(conn, handshakeBuf)
    fmt.Println(string(handshakeBuf))
    hash := handshakeBuf[27:47]
    fmt.Printf("%s", hash)
    if bytes.Compare(hash, bencode.InfoHash[:]) == 0 {
        fmt.Println("OK!!")
    }
}