package main

import (
    "bytes"
    "fmt"
    "encoding/gob"
    "log"
    "os"
)

type Address struct {
    Type, city, country string
}

type Vcard struct {
    FirstName, LastName string
    Adresses []*Address
    Remark string
}

func main() {
    const GOB_FILE = "vc.gob"

    pa := &Address{"private", "beijing", "China"}
    wa := &Address{"work", "shanghai", "China"}
    vc := Vcard{"funky", "gao", []*Address{pa, wa}, "none"}

    file, _  := os.OpenFile(GOB_FILE, os.O_CREATE|os.O_WRONLY, 0600)

    enc := gob.NewEncoder(file)
    if err := enc.Encode(vc); err != nil {
        log.Fatal(err)
    }
    //file.Close()

    f, _ := os.Open(GOB_FILE)
    dec := gob.NewDecoder(f)
    x := new(Vcard)
    fmt.Printf("%v %#v\n", dec.Decode(x), x)

    demoStr()

}

func demoStr() {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    dec := gob.NewDecoder(&buf)

    x := []string{"a", "cdef", "mq"}
    buf.Reset()
    enc.Encode(x)
    result := buf.String()
    fmt.Println(result)

    xx := make([]string, 0)
    buf.Reset()
    buf.WriteString(result)
    dec.Decode(&xx)
    fmt.Printf("%#v\n", xx)
}
