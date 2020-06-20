

package ip

import (
  "net"
  "strings"
)

// prints the local ip address
func Find() string {

  // ifconfig | grep netmask

  if interfaces, err := net.Interfaces(); err == nil {
    for _, interfac := range interfaces {
      if interfac.HardwareAddr.String() != "" {
        if strings.Index(interfac.Name, "en") == 0 ||
          strings.Index(interfac.Name, "eth") == 0 {
          if addrs, err := interfac.Addrs(); err == nil {
            for _, addr := range addrs {
              if addr.Network() == "ip+net" {
                pr := strings.Split(addr.String(), "/")
                if len(pr) == 2 && len(strings.Split(pr[0], ".")) == 4 {
                  return pr[0]
                }
              }
            }
          }
        }
      }
    }
  }
  return "localhost"

}