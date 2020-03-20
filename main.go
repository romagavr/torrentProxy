package main

import (
    "fmt"
    "src"
)

func main()  {
    bencode := bencodeTorrent{}
    bencode = getBencodeStruct("home/roman/docker/golang/test3.torrent")
    fmt.Println(bencode)
}