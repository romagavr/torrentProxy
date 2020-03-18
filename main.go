package main

import (
    "fmt"
    "github.com/romagavr/torrentProxy/bencodeParser"
)

func main()  {
    bencode := bencodeTorrent{}
    bencode = getBencodeStruct("home/roman/docker/golang/test3.torrent")
    fmt.Println(bencode)
}