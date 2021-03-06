package main

import (
	"bytes"
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	// local import
	"application"
)

var app application.App

func TestMain(m *testing.M) {
	app = application.App{}
	app.Initialize(
		os.Getenv("COCKROACH_USER"),
		os.Getenv("COCKROACH_DB"))
	ensureTablesExist()
	code := m.Run()
	clearTables()
	os.Exit(code)
}

func TestEmptyTables(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/recipes", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestGetNonExistentRecipe(t *testing.T) {
	clearTables()

	req, err := http.NewRequest("GET", "/v1/recipes/11", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Recipe not found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Recipe not found'. Got '%s'", m["error"])
	}
}

func TestCreateRecipe(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test recipe" {
		t.Errorf("Expected recipe name to be 'test recipe'. Got '%v'", m["name"])
	}

	if m["preptime"] != 0.1 {
		t.Errorf("Expected recipe price to be '0.1'. Got '%v'", m["preptime"])
	}

	// difficulty is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["difficulty"] != 2.0 {
		t.Errorf("Expected recipe difficulty to be '2'. Got '%v'", m["difficulty"])
	}

	if m["vegetarian"] != true {
		t.Errorf("Expected recipe vegetarian to be 'true'. Got '%v'", m["vegetarian"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected recipe ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetRecipe(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest: %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestUpdatePutRecipe(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - updated","preptime":11.11,"difficulty":3,"vegetarian":false}`)

	req, err = http.NewRequest("PUT", "/v1/recipes/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PUT): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalRecipe["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalRecipe["id"], m["id"])
	}
	if m["name"] == originalRecipe["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalRecipe["name"], m["name"], m["name"])
	}
	if m["preptime"] == originalRecipe["preptime"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalRecipe["preptime"], m["preptime"], m["preptime"])
	}
	if m["difficulty"] == originalRecipe["difficulty"] {
		t.Errorf("Expected the difficulty to change from '%v' to '%v'. Got '%v'", originalRecipe["difficulty"], m["difficulty"], m["difficulty"])
	}
	if m["vegetarian"] == originalRecipe["vegetarian"] {
		t.Errorf("Expected the vegetarian to change from '%v' to '%v'. Got '%v'", originalRecipe["vegetarian"], m["vegetarian"], m["vegetarian"])
	}
}

func TestUpdatePatchRecipe(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	var originalRecipe map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalRecipe)

	payload := []byte(`{"name":"test recipe - updated","preptime":11.11,"difficulty":3,"vegetarian":false}`)

	req, err = http.NewRequest("PATCH", "/v1/recipes/1", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (PATCH): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalRecipe["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalRecipe["id"], m["id"])
	}
	if m["name"] == originalRecipe["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalRecipe["name"], m["name"], m["name"])
	}
	if m["preptime"] == originalRecipe["preptime"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalRecipe["preptime"], m["preptime"], m["preptime"])
	}
	if m["difficulty"] == originalRecipe["difficulty"] {
		t.Errorf("Expected the difficulty to change from '%v' to '%v'. Got '%v'", originalRecipe["difficulty"], m["difficulty"], m["difficulty"])
	}
	if m["vegetarian"] == originalRecipe["vegetarian"] {
		t.Errorf("Expected the vegetarian to change from '%v' to '%v'. Got '%v'", originalRecipe["vegetarian"], m["vegetarian"], m["vegetarian"])
	}
}

func TestDeleteRecipe(t *testing.T) {
	clearTables()
	addRecipes(1)

	req, err := http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (GET): %s", err)
	}
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("DELETE", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (DELETE): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, err = http.NewRequest("GET", "/v1/recipes/1", nil)
	if err != nil {
		t.Errorf("Error on http.NewRequest (Second GET): %s", err)
	}
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)
	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTablesExist() {
	if _, err := app.DB.Exec(recipesTableCreationQuery); err != nil {
		log.Fatal(err)
	}
	if _, err := app.DB.Exec(ratingsTableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTables() {
	app.DB.Exec("DELETE FROM recipes")
	app.DB.Exec("ALTER SEQUENCE recipes_id_seq RESTART WITH 1")
	app.DB.Exec("DELETE FROM recipe_ratings")
	app.DB.Exec("ALTER SEQUENCE recipe_ratings_rating_id_seq RESTART WITH 1")
}

func TestAddRating(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (1st POST): %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	payload = []byte(`{"rating":3}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (2nd POST): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["rating_id"] != 1.0 {
		t.Errorf("Expected rating ID to be '1'. Got '%v'", m["rating_id"])
	}
}

func TestSearch(t *testing.T) {
	clearTables()

	payload := []byte(`{"name":"test recipe","preptime":0.1,"difficulty":2,"vegetarian":true}`)

	req, err := http.NewRequest("POST", "/v1/recipes", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (1st POST): %s", err)
	}
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	payload = []byte(`{"rating":3}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (2nd POST): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["rating_id"] != 1.0 {
		t.Errorf("Expected rating ID to be '1'. Got '%v'", m["rating_id"])
	}

	payload = []byte(`{"rating":2}`)

	req, err = http.NewRequest("POST", "/v1/recipes/1/rating", bytes.NewBuffer(payload))
	if err != nil {
		t.Errorf("Error on http.NewRequest (3rd POST): %s", err)
	}
	response = executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	json.Unmarshal(response.Body.Bytes(), &m)

	// the id is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["rating_id"] != 2.0 {
		t.Errorf("Expected rating ID to be '2'. Got '%v'", m["rating_id"])
	}

	var bb bytes.Buffer
	mw := multipart.NewWriter(&bb)
	mw.WriteField("count", "1")
	mw.WriteField("start", "0")
	mw.WriteField("preptime", "50.0")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/recipes/search", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (4th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var mm []map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &mm)
	// only want the first one
	m = mm[0]

	if m["name"] != "test recipe" {
		t.Errorf("Expected recipe name to be 'test recipe'. Got '%v'", m["name"])
	}

	if m["preptime"] != 0.1 {
		t.Errorf("Expected recipe price to be '0.1'. Got '%v'", m["preptime"])
	}

	// difficulty is compared to 2.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["difficulty"] != 2.0 {
		t.Errorf("Expected recipe difficulty to be '2'. Got '%v'", m["difficulty"])
	}

	if m["vegetarian"] != true {
		t.Errorf("Expected recipe vegetarian to be 'true'. Got '%v'", m["vegetarian"])
	}

	// the avg_rating is compared to 2.5 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["avg_rating"] != 2.5 {
		t.Errorf("Expected average recipe rating to be '2.5'. Got '%v'", m["id"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	//     floats (float64), when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected recipe ID to be '1'. Got '%v'", m["id"])
	}

	addRecipes(12)

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "10")
	mw.WriteField("start", "1")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/recipes/search", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (5th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 10 {
		t.Errorf("Expected '10' recipes. Got '%v'", len(mm))
	}

	mw = multipart.NewWriter(&bb)
	mw.WriteField("count", "10")
	mw.WriteField("start", "1")
	mw.WriteField("preptime", "30.0")
	mw.Close()

	req, err = http.NewRequest("POST", "/v1/recipes/search", &bb)
	if err != nil {
		t.Errorf("Error on http.NewRequest (6th POST): %s", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	json.Unmarshal(response.Body.Bytes(), &mm)

	// Search page limit
	if len(mm) != 2 {
		t.Errorf("Expected '2' recipes. Got '%v'", len(mm))
	}
}

func addRecipes(count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		app.DB.Exec("INSERT INTO recipes(name, preptime, difficulty, vegetarian) VALUES($1, $2, $3, $4)",
			"Recipe "+strconv.Itoa(i), (i+1.0)*10, i%3+1, true)
	}
}

func addRecipeRatings(recipe int, count int) {
	if count < 1 {
		count = 1
	}
	for i := 0; i < count; i++ {
		addRecipeRating(recipe, i%5+1)
	}
}

func addRecipeRating(recipe int, rating int) {
	app.DB.Exec("INSERT INTO recipe_ratings(recipe_id, rating) VALUES($1, $2)", recipe, rating)
}

const recipesTableCreationQuery = `CREATE TABLE IF NOT EXISTS recipes
(
	id BIGSERIAL,
	name TEXT NOT NULL,
	preptime FLOAT(4) NOT NULL DEFAULT 0.0,
	difficulty NUMERIC(1) NOT NULL CHECK (difficulty > 0) CHECK (difficulty < 4) DEFAULT 0,
	vegetarian BOOLEAN NOT NULL DEFAULT false,
	CONSTRAINT recipes_pkey PRIMARY KEY (id)
)`

const ratingsTableCreationQuery = `CREATE TABLE IF NOT EXISTS recipe_ratings
(
	recipe_id BIGINT REFERENCES recipes(id) ON DELETE CASCADE,
	rating_id BIGSERIAL,
	rating SMALLINT NOT NULL CHECK (rating > 0) CHECK (rating < 6) DEFAULT 0,
	PRIMARY KEY (recipe_id, rating_id)
)`
