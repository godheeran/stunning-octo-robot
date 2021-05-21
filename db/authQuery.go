package db

import (
	"app/model"
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

var ErrINVALIDPASSWORD = errors.New("Invalid password")
var ErrINVALIDEMAIL = errors.New("Already account associated")

// UserSignIn is compare email(id) and password when login page
func UserSignIn(email, password string) (model.User, error) {
	if UserClient == nil {
		dbConnect()
	}
	// collection := Client.Database("testApp").Collection("user")
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	// Check email
	var dbUser model.User
	err := UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&dbUser)
	if err != nil {
		return dbUser, err
	}

	if !checkPassword(dbUser.Password, password) {
		return dbUser, ErrINVALIDPASSWORD
	}

	dbUser.Password = ""

	return dbUser, nil
}

//UserSignUp is add new user info
func UserSignUp(user model.User) error {
	if UserClient == nil {
		getDBConfig()
	}
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	// Check email(unique)
	var dbUser model.User
	err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err == nil { // exist
		return ErrINVALIDEMAIL
	}

	err = hashPassword(&user.Password)
	if err != nil {
		return err
	}

	_, err = UserCollection.InsertOne(ctx, bson.D{
		{Key: "firstname", Value: user.FirstName},
		{Key: "lastname", Value: user.LastName},
		{Key: "email", Value: user.Email},
		{Key: "password", Value: user.Password},
	})
	if err != nil {
		return err
	}

	log.Println("Inserted New User into User Collection!")
	return nil

}

// Check the password is correct or not.
// This method will return an error if the hash does not match the provided password string.
func checkPassword(existingHash, incomingPass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(existingHash), []byte(incomingPass)) == nil
}

// Get the hash value of a password.
func hashPassword(s *string) error {
	if s == nil {
		return errors.New("Reference provided for hashing password is nil")
	}
	sBytes := []byte(*s)                                                        // Convert password string to byte slice.
	hashedBytes, err := bcrypt.GenerateFromPassword(sBytes, bcrypt.DefaultCost) // Obtain hashed password.
	if err != nil {
		return err
	}
	*s = string(hashedBytes[:]) // Update password string with the hashed version.
	return nil
}
