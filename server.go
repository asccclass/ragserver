package main

import (
   "os"
   "fmt"
   "github.com/asccclass/sherryserver"
)

func main() {
   port := os.Getenv("PORT")
   if port == "" {
      port = "80"
   }
   documentRoot := os.Getenv("DocumentRoot")
   if documentRoot == "" {
      documentRoot = "www"
   }
   templateRoot := os.Getenv("TemplateRoot")
   if templateRoot == "" {
      templateRoot = "www/html"
   }

   server, err := SherryServer.NewServer(":" + port, documentRoot, templateRoot)
   if err != nil {
      panic(err)
   }
   router := NewRouter(server, documentRoot)
   if router == nil {
      fmt.Println("router return nil")
      return
   }
   server.Server.Handler = router  // server.CheckCROS(router)  // 需要自行implement, overwrite 預設的
   server.Start()
}
