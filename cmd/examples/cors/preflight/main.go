package main

import (
	"flag"
	"log"
	"net/http"
)

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
</head>
<body>
    <h1>Course Registration - Preflight CORS</h1>
    <h2> Student has been added to database </h2>
    <div id="output"></div>
    <script>
         document.addEventListener('DOMContentLoaded', function() {
         fetch("http://localhost:4000/v1/students", {
           method: "POST",
           headers: {
                    'Content-Type': 'application/json'
                    },
           body: JSON.stringify({
                    Fname:'Catherine',
                    Lname:'Fuller', 
                    gender: 'Female', 
                    age: 23,
                    district_id: 1,
                    program: 13,
                    gpa: 3.21
                 })
           }).then( function(response) {
               response.text().then(function (text) {
                  document.getElementById("output").innerHTML = text;
               });
             },
             function(err) {
               document.getElementById("output").innerHTML = err;
             }
          );
       });
  </script>
</body>
</html>`

func main() {
	addr := flag.String("addr", ":9000", "Server address")
	flag.Parse()

	log.Printf("starting server on %s", *addr)

	err := http.ListenAndServe(*addr,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(html))
		}))
	log.Fatal(err)
}
