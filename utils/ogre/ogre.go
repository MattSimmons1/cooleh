
package ogre

import (
  "fmt"
  "time"
)

type Ogre struct {
  message string
}

func New(message string) Ogre {
  o := Ogre {message}
  return o
}

func (o Ogre) Growl() {

  letter := "o"

  for i := 0; i < 10; i++ {
    time.Sleep(1e7)
    o.message += letter
    fmt.Printf("\rcoo%vleh", o.message)
  }
  fmt.Printf("\n")
}