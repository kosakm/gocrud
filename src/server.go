package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	handlers "github.com/gorilla/handlers"
	mux "github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	emp "models"
)

var c *mgo.Collection

func bootstrapDB() {
	fmt.Println("bootstrapping the database")
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	//defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	c = session.DB("test").C("employee")

	// Clear current records
	c.RemoveAll(nil)

	employees := make([]interface{}, 5)

	employees[0] = emp.Employee{"123", "Ale", "Smith", "801 555-5555"}
	employees[1] = emp.Employee{"345", "Matt", "Williams", "801 555-5554"}
	employees[2] = emp.Employee{"782", "Ange", "Cooks", "801 555-5553"}
	employees[3] = emp.Employee{"998", "Paul", "Escobar", "801 555-5552"}
	employees[4] = emp.Employee{"672", "Tim", "McDonough", "801 555-5551"}
	err = c.Insert(employees...)
	if err != nil {
		log.Fatal(err)
	}

	numOfEmps, errr := c.Count()
	if errr != nil {
		log.Fatal(err)
	} else {
		fmt.Println("# of Employees:", numOfEmps)
	}
}

func getAllEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	var people []emp.Employee
	err := c.Find(nil).All(&people)

	if err != nil {
		log.Fatal("Error looking up people", err)
	}

	peopleJSON, err := json.Marshal(people)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(peopleJSON)
}

func getEmployeeHandler(w http.ResponseWriter, r *http.Request) {
	var people []emp.Employee
	err := c.Find(bson.M{"id": mux.Vars(r)["empId"]}).All(&people)

	if err != nil {
		log.Fatal("Error looking up people", err)
	}

	peopleJSON, err := json.Marshal(people)
	if err != nil {
		fmt.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(peopleJSON)
}

func putEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
	w.Write([]byte("PUT"))
}

func postEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	employee := r.Context().Value(PostDataValueKey)
	emp, ok := employee.(emp.Employee)
	if ok {
		fmt.Println("decoded emp", emp.FirstName)
		c.Insert(emp)
		w.Write([]byte("POST"))
	} else {
		w.Write([]byte("Error casting"))
	}
}

type PostDataValueKey struct{}

func jsonBodyMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var employee emp.Employee
		decoder := json.NewDecoder(r.Body)
		error := decoder.Decode(&employee)

		if error != nil {
			fmt.Println(error)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Error parsing json"))
		}
		fmt.Println("after decoding", employee.FirstName)
		ctx := context.WithValue(r.Context(), PostDataValueKey, employee)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	env := os.Getenv("ENV")
	fmt.Printf("Running in %v mode \n", env)

	if strings.ToLower(env) == "dev" {
		bootstrapDB()
	}

	router := mux.NewRouter()

	// Employee Routes
	router.HandleFunc("/employees", getAllEmployeesHandler).Methods("GET")
	router.HandleFunc("/employee/{empId}", getEmployeeHandler).Methods("GET")
	router.HandleFunc("/employee/{empId}", putEmployeesHandler).Methods("PUT")
	router.Handle("/employee", jsonBodyMiddleware(http.HandlerFunc(postEmployeesHandler))).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS()(handlers.LoggingHandler(os.Stdout, router))))
}
