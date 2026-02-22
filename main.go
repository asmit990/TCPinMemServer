package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Item struct {
	value     string
	expiresAt int64
}


var (
	store = make(map[string]string)
	mu    sync.RWMutex 

)


func main() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer l.Close()

	fmt.Println("Server is listening on port 8080")

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		
		go handlerClient(conn)
	}
}

func handlerClient(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return 
		}

		m := strings.TrimSpace(string(buffer[:n]))
		p := strings.Fields(m)

		if len(p) == 0 {
			continue
		}

		
		switch strings.ToUpper(p[0]) {
		case "SET":
			if len(p) != 3 {
				conn.Write([]byte("error: SET requires key and value\n"))
				continue
			}

          
            
               item := Item{
                 value: p[2],
                     }
            if len(p) == 5 && strings.ToUpper(p[3]) == "EX" {
                seconds, err := strconv.Atoi(p[4])
                if err == nil {
                    item.expiresAt = time.Now().Unix() + int64(seconds)
                }
            }
			mu.Lock() 
			store[p[1]] = p[2]
			mu.Unlock()
			conn.Write([]byte("ok\n"))

		case "GET":
			if len(p) != 2 {
				conn.Write([]byte("error: GET requires a key\n"))
				continue
			}
			mu.RLock()
			v, ok := store[p[1]]
			mu.RUnlock()

			if !ok {
				conn.Write([]byte("nil\n"))
			} else {
				conn.Write([]byte(v + "\n"))
			}

        case "DEL":
            if len(p) != 2 {
				conn.Write([]byte("error: GET requires a key\n"))
				continue
			}
            mu.Lock()
            _, ok := store[p[1]]
            if ok {
                delete(store, p[1])
            }
            mu.Unlock()
            if ok {
            conn.Write([]byte("1\n")) 
           } else {
            conn.Write([]byte("0\n")) 
             }

             
        case "LEN":
            mu.RLock()
            l := len(store)
            mu.RUnlock()
            conn.Write([]byte(fmt.Sprintf("%d\n", l)))

        case "EXISTS":
           if len(p) != 2 {
          conn.Write([]byte("error: EXISTS requires a key\n"))
                       continue
                   }

                   mu.RLock()
                   _, ok := store[p[1]]
               mu.RUnlock()
               
              if ok {
                       conn.Write([]byte("1\n"))
                   } else {
              conn.Write([]byte("0\n"))
                   }
		default:
			conn.Write([]byte("UNKNOWN COMMAND\n"))
		}
	}
}