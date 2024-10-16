package SherryRAG

import (
   "fmt"
   "mime"
   "net/http"
   "encoding/json"
)

func (app *RAGServer) ReadRequestJSON(req *http.Request, target any)(error) {
   contentType := req.Header.Get("Content-Type")
   mediaType, _, err := mime.ParseMediaType(contentType)
   if err != nil {
      return err
   }
   if mediaType != "application/json" {
      return fmt.Errorf("expect application/json Content-Type, got %s", mediaType)
   }

   dec := json.NewDecoder(req.Body)
   dec.DisallowUnknownFields()
   return dec.Decode(target)
}

// renderJSON renders 'v' as JSON and writes it as a response into w.
func (app *RAGServer) RenderJSON(w http.ResponseWriter, v any) {
   js, err := json.Marshal(v)
   if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
   w.Header().Set("Content-Type", "application/json")
   w.Write(js)
}
