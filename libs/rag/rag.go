package SherryRAG

import(
   "os"
   "fmt"
   "context"
   "net/http"
   "google.golang.org/api/option"
   "github.com/asccclass/sherryserver"
   "github.com/google/generative-ai-go/genai"
   "github.com/weaviate/weaviate-go-client/v4/weaviate"
   "github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
)

type RAGServer {
   Srv		*SherryServer.Server
   Ctx		context.Context
   WvClient	*weaviate.Client
   AIModel	*genai.GenerativeModel
   EmbModel	*genai.EmbeddingModel
}

func(app *RAGServer) Close() {
}

func(app *RAGServer) DecodeGetResults(result *models.GraphQLResponse) ([]string, error) {
   data, ok := result.Data["Get"]
   if !ok {
      return nil, fmt.Errorf("Get key not found in result")
   }
   doc, ok := data.(map[string]any)
   if !ok {
      return nil, fmt.Errorf("Get key unexpected type")
   }
   slc, ok := doc["Document"].([]any)
   if !ok {
      return nil, fmt.Errorf("Document is not a list of results")
   }

   var out []string
   for _, s := range slc {
      smap, ok := s.(map[string]any)
      if !ok {
         return nil, fmt.Errorf("invalid element in list of documents")
      }
      s, ok := smap["text"].(string)
      if !ok {
         return nil, fmt.Errorf("expected string in list of documents")
      }
      out = append(out, s)
   }
   return out, nil
}

// Query
func(app *RAGServer) QueryDocuments(w http.ResponseWriter, req *http.Request) {
   type queryRequest struct {
      Content string
   }
   qr := &queryRequest{}
   if err := app.ReadRequestJSON(req, qr); err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
   }

   // Embed the query contents.
   rsp, err := app.EmbModel.EmbedContent(app.Ctx, genai.Text(qr.Content))
   if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }

   gql := app.WvClient.GraphQL()
   result, err := gql.Get().
      WithNearVector(
         gql.NearVectorArgBuilder().WithVector(rsp.Embedding.Values)).
      WithClassName("Document").
      WithFields(graphql.Field{Name: "text"}).
      WithLimit(3).
      Do(app.Ctx)
   if werr := combinedWeaviateError(result, err); werr != nil {
      http.Error(w, werr.Error(), http.StatusInternalServerError)
      return
   }

   contents, err := app.DecodeGetResults(result)
   if err != nil {
      http.Error(w, fmt.Errorf("reading weaviate response: %w", err).Error(), http.StatusInternalServerError)
      return
   }

   // Create a RAG query for the LLM with the most relevant documents as
   // context.
   ragQuery := fmt.Sprintf(ragTemplateStr, qr.Content, strings.Join(contents, "\n"))
   resp, err := app.GenModel.GenerateContent(app.Ctx, genai.Text(ragQuery))
   if err != nil {
      http.Error(w, "generative model error", http.StatusInternalServerError)
      return
   }

   if len(resp.Candidates) != 1 {
      http.Error(w, "generative model error", http.StatusInternalServerError)
      return
   }

   var respTexts []string
   for _, part := range resp.Candidates[0].Content.Parts {
      if pt, ok := part.(genai.Text); ok {
         respTexts = append(respTexts, string(pt))
      } else {
         log.Printf("bad type of part: %v", pt)
         http.Error(w, "generative model error", http.StatusInternalServerError)
         return
      }
   }

   app.RenderJSON(w, strings.Join(respTexts, "\n"))
}

// Add Documents
func (app *RAGServer) AddDocuments(w http.ResponseWriter, req *http.Request) {
   // Parse HTTP request from JSON.
   type document struct {
      Text string
   }
   type addRequest struct {
      Documents []document
   }
   ar := &addRequest{}

   err := app.ReadRequestJSON(req, ar)
   if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
   }

   // Use the batch embedding API to embed all documents at once.
   batch := app.embModel.NewBatch()
   for _, doc := range ar.Documents {
      batch.AddContent(genai.Text(doc.Text))
   }
   rsp, err := app.embModel.BatchEmbedContents(rs.ctx, batch)
   if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
   if len(rsp.Embeddings) != len(ar.Documents) {
      http.Error(w, "embedded batch size mismatch", http.StatusInternalServerError)
      return
   }

   objects := make([]*models.Object, len(ar.Documents))
   for i, doc := range ar.Documents {
      objects[i] = &models.Object{
         Class: "Document",
         Properties: map[string]any{
            "text": doc.Text,
         },
         Vector: rsp.Embeddings[i].Values,
      }
   }

   // Store documents with embeddings in the Weaviate DB.
   _, err = app.wvClient.Batch().ObjectsBatcher().WithObjects(objects...).Do(app.Ctx)
   if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

// Router
func(app *RAGServer) AddRouter(router *http.ServeMux) {
   router.HandleFunc("POST /add/", app.AddDocuments)
   router.HandleFunc("POST /query/", app.QueryDocuments)
}

func NewRAGServer(srv *SherryServer.Server)(*RAGServer, error) {
   aiName := os.Getenv("AI")
   if aiName == "" {
      return nil,fmt.Errorf("envfile has no AI")
   }
   aiKey := os.Getenv(aiName + "_API_KEY")
   if aiKey == "" {
      return nil,fmt.Errorf("envfile has no AI Key")
   }
   ctx := context.Background()
   wvClient, err := initWeaviate(ctx)
   if err != nil {
      return nil, err
   }
   genaiClient, err := genai.NewClient(ctx, option.WithAPIKey(aiKey))
   if err != nil {
      return nil, err
   }
   aiModel := os.Getenv(aiName + "_ModelName")
   embModel := os.Getenv(aiName + "_EmbeddingModel")
   if aiKey == "" or embModel == "" {
      return nil,fmt.Errorf("envfile has no AI Model Name or embeding model name")
   }

   return &RAGServer {
      Srv: srv,
      Ctx: ctx,
      WvClient: wvClient,
      AIModel:  genaiClient.GenerativeModel(aiModel),
      EmbModel: genaiClient.EmbeddingModel(embModel),
   }, nil
}
