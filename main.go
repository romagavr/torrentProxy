package main

import (
    "fmt"
    _ "fmt"
    "torrentProxy/src"
)

func main()  {
    bencode := src.BencodeTorrent{}
    bencode = src.GetBencodeStruct("/home/roman/golang/testfiles/2.torrent")
    fmt.Println(bencode.InfoHash)
    //base, err := url.Parse(bencode.Announce)
    //if err != nil {
    //    fmt.Println("error");
    //    return
    //}
    //fmt.Println(base)
}