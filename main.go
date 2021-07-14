package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3032"
	}

	//minioClient, err := minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
	//	Creds:  credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), ""),
	//	Secure: false,
	//})
	//if err != nil {
	//	log.Println("Failed to connect to Minio", err)
	//}

	http.HandleFunc("/demo", Demo)
	http.HandleFunc("/iframe", IFrame)
	http.HandleFunc("/envs", Envs)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("SUCCESS!"))
		w.Write([]byte("\n"))
		for _, e := range os.Environ() {
			fmt.Fprintf(w, "%s\n", e)
		}
	})
	log.Println("Starting server on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func Envs(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	for _, e := range os.Environ() {
		w.Write([]byte(e + "\n"))
	}
}

func Demo(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	w.Write([]byte(`
<html>
<head>
<title>Demo</title>
</head>
<body>
<h1>IFrame Demo</h1>
<button data-message="increment">Increment</button>
<button data-message="decrement">Decrement</button>
<br>
<br>
<iframe src="/iframe"></iframe>

<script>
const iframe = document.querySelector("iframe");
console.log("IFRAME", iframe);
setInterval(() => {
	iframe.contentWindow.postMessage("ping", "*");
}, 5000);

window.addEventListener("message", e => {
  console.log("PARENT MESSAGE", e.origin, e.data);
}, false)

document.querySelectorAll("[data-message]").forEach(el => {
  el.addEventListener("click", () => {
	  iframe.contentWindow.postMessage(el.getAttribute('data-message'), "*");
  })
})
</script>

</body>
</html>
`))
}

func IFrame(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	w.Write([]byte(`
<html>
<head>
<title>IFrame</title>
</head>
<body>
<h1>This is the iframe!</h1>
<p id="count">0</p>
</body>

<script>
const countEl = document.querySelector("#count");
window.addEventListener("message", e => {
  console.log("IFRAME MESSAGE", e.origin, e.data);
  if (e.data === "increment") countEl.innerHTML = parseInt(countEl.innerHTML) + 1;
  if (e.data === "decrement") countEl.innerHTML = parseInt(countEl.innerHTML) - 1;
  if (e.data === "ping") e.source.postMessage("pong", e.origin);
}, false)
window.parent.window.postMessage("testing from child", "*");
</script>

</html>
`))
}

//func ListBuckets(minio *minio.Client) func(w http.ResponseWriter, r *http.Request) {
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Println(r.URL.String())
//		if r.URL.Path != "/" {
//			return
//		}
//		if minio == nil {
//			http.Error(w, "MinIO not configured", http.StatusInternalServerError)
//			return
//		}
//
//		buckets, err := minio.ListBuckets(r.Context())
//		if err != nil {
//			log.Println("Failed to list buckets", err)
//			http.Error(w, "Failed to list buckets", http.StatusInternalServerError)
//			return
//		}
//
//		w.Header().Set("content-type", "application/json")
//
//		enc := json.NewEncoder(w)
//		_ = enc.Encode(map[string]interface{}{
//			"buckets": buckets,
//		})
//	}
//}
