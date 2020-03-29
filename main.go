package main

import (
    "bytes"
    "fmt"
    _ "fmt"
    "io"
    "io/ioutil"
    "net"
    _ "net/http"
    "strconv"
    "time"
    "torrentProxy/src"
)

func main()  {
    // Parse bencodeed File
    data, err := ioutil.ReadFile("/home/roman/golang/testfiles/9.torrent")
    bencode := src.BencodeTorrent{}
    bencode.GetBencodeStruct(data)

    response := bencode.GetTorrentRequest()

    //Format Bittorrent handshake
    var handShake []byte
    handShake = make([]byte, 68)
    handShake[0] = byte(len("BitTorrent protocol"))
    count := 1
    count += copy(handShake[count:], "BitTorrent protocol")
    count += copy(handShake[count:], make([]byte, 8))
    count += copy(handShake[count:], bencode.InfoHash[:])
    count += copy(handShake[count:], bencode.PeerId[:])
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