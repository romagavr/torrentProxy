package main

import (
    "fmt"
    _ "fmt"
    "io/ioutil"
    "math/rand"
    "net/http"
    "net/url"
    "strconv"
    "torrentProxy/src"
)

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
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    fmt.Printf("%s", body)
}