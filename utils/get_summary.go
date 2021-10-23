package main

import (
  "io/ioutil"
  "net/http"
  "fmt"
)

func main() {
  resp, err := http.Post("https://api-3moji.herokuapp.com/api/v1/summary/", "", nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Printf("%s", body)
}
