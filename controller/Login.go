package controller

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	model "github.com/anakilang-ai/backend/models"
	"github.com/anakilang-ai/backend/utils"
	"github.com/badoux/checkmail"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/argon2"
)

// LogIn handles user login requests
func LogIn(db *mongo.Database, respw http.ResponseWriter, req *http.Request, privateKey string) {
	var user model.User

	// Decode the request body into the user struct
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		utils.ErrorResponse(respw, req, http.StatusBadRequest, "Bad Request", "Error parsing request body: "+err.Error())
		return
	}

	// Validate the input
	if user.Email == "" || user.Password == "" {
		utils.ErrorResponse(respw, req, http.StatusBadRequest, "Bad Request", "Email and password are required")
		return
	}

	if err := checkmail.ValidateFormat(user.Email); err != nil {
		utils.ErrorResponse(respw, req, http.StatusBadRequest, "Bad Request", "Invalid email format")
		return
	}

	// Retrieve user from database
	existsDoc, err := utils.GetUserFromEmail(user.Email, db)
	if err != nil {
		utils.ErrorResponse(respw, req, http.StatusInternalServerError, "Internal Server Error", "Error retrieving user: "+err.Error())
		return
	}

	// Decode the salt and hash the password
	salt, err := hex.DecodeString(existsDoc.Salt)
	if err != nil {
		utils.ErrorResponse(respw, req, http.StatusInternalServerError, "Internal Server Error", "Error decoding salt: "+err.Error())
		return
	}
	hash := argon2.IDKey([]byte(user.Password), salt, 1, 64*1024, 4, 32)

	// Compare the hashed password with the stored hash
	if hex.EncodeToString(hash) != existsDoc.Password {
		utils.ErrorResponse(respw, req, http.StatusUnauthorized, "Unauthorized", "Incorrect password")
		return
	}

	// Generate the authentication token
	tokenString, err := utils.Encode(user.ID, user.Email, privateKey)
	if err != nil {
		utils.ErrorResponse(respw, req, http.StatusInternalServerError, "Internal Server Error", "Error generating token: "+err.Error())
		return
	}

	// Respond with success and the token
	resp := map[string]string{
		"status":  "success",
		"message": "Login successful",
		"token":   tokenString,
	}
	utils.WriteJSON(respw, http.StatusOK, resp)
}
