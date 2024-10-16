// router.go
package main

import(
   "net/http"
   "github.com/asccclass/sherryserver"
   "github.com/asccclass/ragserver/libs/rag"
   "github.com/asccclass/sherryserver/libs/oauth"
)

func NewRouter(srv *SherryServer.Server, documentRoot string)(*http.ServeMux) {
   router := http.NewServeMux()

   // Static File server
   staticfileserver := SherryServer.StaticFileServer{documentRoot, "index.html"}
   staticfileserver.AddRouter(router)
   rag, err := SherryRAG.NewRAGServer(srv)
   if err == nil {
      rag.AddRouter(router)
   }
/*
   // Oauth
   oauth, err := Oauth.NewOauth(srv)
   if err == nil {
      oauth.AddRouter(router)
   }
   // App router
   router.Handle("/homepage", oauth.Protect(http.HandlerFunc(Home)))
*/
   router.Handle("/logout", http.HandlerFunc(oauth.Logout))
   return router
}
