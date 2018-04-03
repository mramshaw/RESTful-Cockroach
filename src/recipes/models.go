package recipes

import "database/sql"

// The Recipe entity is used to marshall/unmarshall JSON.
type Recipe struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	PrepTime   float32 `json:"preptime"`
	Difficulty int     `json:"difficulty"`
	Vegetarian bool    `json:"vegetarian"`
}

// The RecipeRated entity is used to marshall/unmarshall JSON.
type RecipeRated struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	PrepTime   float32 `json:"preptime"`
	Difficulty int     `json:"difficulty"`
	Vegetarian bool    `json:"vegetarian"`
	AvgRating  float32 `json:"avg_rating"`
}

// The RecipeRating entity is used to marshall/unmarshall JSON.
type RecipeRating struct {
	ID       int `json:"rating_id"`
	RecipeID int `json:"recipe_id"`
	Rating   int `json:"rating"`
}

// GetRecipe returns a single specified recipe.
func (r *Recipe) GetRecipe(db *sql.DB) error {
	return db.QueryRow("SELECT name, preptime, difficulty, vegetarian FROM recipes WHERE id=$1",
		r.ID).Scan(&r.Name, &r.PrepTime, &r.Difficulty, &r.Vegetarian)
}

// UpdateRecipe is used to modify a specific recipe.
func (r *Recipe) UpdateRecipe(db *sql.DB) (res sql.Result, err error) {
	res, err = db.Exec("UPDATE recipes SET name=$1, preptime=$2, difficulty=$3, vegetarian=$4 WHERE id=$5",
		r.Name, r.PrepTime, r.Difficulty, r.Vegetarian, r.ID)
	return res, err
}

// DeleteRecipe is used to delete a specific recipe.
func (r *Recipe) DeleteRecipe(db *sql.DB) (res sql.Result, err error) {
	res, err = db.Exec("DELETE FROM recipes WHERE id=$1", r.ID)
	return res, err
}

// CreateRecipe is used to create a single recipe.
func (r *Recipe) CreateRecipe(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO recipes(name, preptime, difficulty, vegetarian) VALUES($1, $2, $3, $4) RETURNING id",
		r.Name, r.PrepTime, r.Difficulty, r.Vegetarian).Scan(&r.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetRecipes returns a collection of known recipes.
func GetRecipes(db *sql.DB, start int, count int) ([]Recipe, error) {
	rows, err := db.Query(
		"SELECT id, name, preptime, difficulty, vegetarian FROM recipes LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	recipes := []Recipe{}
	for rows.Next() {
		var r Recipe
		if err := rows.Scan(&r.ID, &r.Name, &r.PrepTime, &r.Difficulty, &r.Vegetarian); err != nil {
			return nil, err
		}
		recipes = append(recipes, r)
	}

	return recipes, nil
}

// GetRecipesRated returns a collection of rated recipes.
func GetRecipesRated(db *sql.DB, start int, count int, preptime float32) ([]RecipeRated, error) {
	rows, err := db.Query(
		"SELECT id, name, preptime, difficulty, vegetarian, "+
			"(SELECT COALESCE(AVG(rating),0) AS avg_rating FROM recipe_ratings WHERE recipe_id = id)"+
			" FROM recipes WHERE preptime < $1 LIMIT $2 OFFSET $3",
		preptime, count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	recipesRated := []RecipeRated{}
	for rows.Next() {
		var rr RecipeRated
		if err := rows.Scan(&rr.ID, &rr.Name, &rr.PrepTime, &rr.Difficulty, &rr.Vegetarian, &rr.AvgRating); err != nil {
			return nil, err
		}
		recipesRated = append(recipesRated, rr)
	}

	return recipesRated, nil
}

// AddRecipeRating adds a rating for a specific recipe.
// There can be many ratings for any specific recipe
// and the ratings are never overwritten.
func (rr *RecipeRating) AddRecipeRating(db *sql.DB) error {
	err := db.QueryRow(
		"INSERT INTO recipe_ratings(recipe_id, rating) VALUES($1, $2) RETURNING rating_id",
		rr.RecipeID, rr.Rating).Scan(&rr.ID)

	if err != nil {
		return err
	}

	return nil
}
