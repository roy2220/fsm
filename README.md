# fsm

[![Build Status](https://travis-ci.com/roy2220/fsm.svg?branch=master)](https://travis-ci.com/roy2220/fsm) [![Coverage Status](https://codecov.io/gh/roy2220/fsm/branch/master/graph/badge.svg)](https://codecov.io/gh/roy2220/fsm)

File Space Management

## Example

```go
package main

import (
        "encoding/binary"
        "fmt"
        "os"

        "github.com/roy2220/fsm"
)

func main() {
        const fn = "./test_storage"
        defer os.Remove(fn)

        // store a list of words
        func() {
                fs := new(fsm.FileStorage).Init()

                if err := fs.Open(fn, true); err != nil {
                        panic(err)
                }

                defer fs.Close()
                // set the list head
                fs.SetPrimarySpace(-1)

                for _, word := range []string{"cat", "dog", "fish", "ant", "bird"} {
                        // allocate space for a list item
                        space, buf := fs.AllocateSpace(8 + 1 + len(word))
                        // set the next list item (8 bytes)
                        binary.BigEndian.PutUint64(buf[0:], uint64(fs.PrimarySpace()))
                        // set the word length (1 byte)
                        buf[8] = byte(len(word))
                        // set the word (n byte)
                        copy(buf[9:], word)
                        // update the list head
                        fs.SetPrimarySpace(space)
                }
        }()

        // load the list of words
        func() {
                fs := new(fsm.FileStorage).Init()

                if err := fs.Open(fn, true); err != nil {
                        panic(err)
                }

                defer fs.Close()
                // get the list head
                nextSpace := fs.PrimarySpace()

                for nextSpace >= 0 {
                        space := nextSpace
                        buf := fs.AccessSpace(space)
                        // get the next list item (8 bytes)
                        nextSpace = int64(binary.BigEndian.Uint64(buf[0:]))
                        // get the word length (1 byte)
                        wordLen := int(buf[8])
                        // get the word (n byte)
                        word := string(buf[9 : 9+wordLen])
                        fmt.Println(word)
                        // release the space of the list item
                        // (free space could be reused by the following AllocateSpace calls)
                        //fs.FreeSpace(space)
                }
        }()
}
```

## Documentation

See https://godoc.org/github.com/roy2220/fsm#example-package
