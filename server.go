package govision

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	vision "google.golang.org/api/vision/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"google.golang.org/appengine/user"
)

const (
	authEmail = "https://www.googleapis.com/auth/userinfo.email"
	authGCP   = "https://www.googleapis.com/auth/cloud-platform"
)

// Data ...
type Data struct {
	Base64Str string `json:"image"`
}

func init() {
	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/api/vision", HandleVision)
}

// HandleRoot ...
func HandleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")
	ctx := appengine.NewContext(r)
	u := user.Current(ctx)
	if u == nil {
		url, err := user.LoginURL(ctx, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return
	}

	http.ServeFile(w, r, "index.html")
}

// HandleVision ...
func HandleVision(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	log.Debugf(c, ">>> HandleVision")

	data, err := json2Data(r.Body)
	if err != nil {
		log.Errorf(c, "Error json2Data : %v", err)
		http.Error(w, "Invalid json format: "+err.Error(), http.StatusBadRequest)
		return
	}

	service, err := createService(c)
	if err != nil {
		log.Errorf(c, "Error, initialize service account: %v", err)
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := service.Images.Annotate(createRequests(data)).Do()
	if err != nil {
		log.Errorf(c, "Error, Annotate: %v", err)
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		log.Errorf(c, "Error, json.Marshal: %v", err)
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", jsonData)
}

func createService(c context.Context) (*vision.Service, error) {
	client := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(c, authEmail, authGCP),
			Base: &urlfetch.Transport{
				Context: c,
			},
		},
	}
	return vision.New(client)
}

func createRequests(data *Data) *vision.BatchAnnotateImagesRequest {
	req := &vision.AnnotateImageRequest{
		Image: &vision.Image{
			Content: data.Base64Str,
		},
		Features: []*vision.Feature{
			{
				MaxResults: 5,
				Type:       "FACE_DETECTION",
			},
			{
				MaxResults: 5,
				Type:       "LABEL_DETECTION",
			},
			{
				MaxResults: 5,
				Type:       "LANDMARK_DETECTION",
			},
			{
				MaxResults: 5,
				Type:       "LOGO_DETECTION",
			},
			{
				MaxResults: 1,
				Type:       "TEXT_DETECTION",
			},
		},
	}

	return &vision.BatchAnnotateImagesRequest{
		Requests: []*vision.AnnotateImageRequest{req},
	}
}

func json2Data(rc io.ReadCloser) (*Data, error) {
	defer rc.Close()
	var data Data
	err := json.NewDecoder(rc).Decode(&data)
	return &data, err
}
