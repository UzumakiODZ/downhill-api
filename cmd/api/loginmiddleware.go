package cmd

import (
	"fmt"
	"math/rand"
	"regexp"
	"net/http"
	"time"
	"golang.org/x/crypto/bcrypt"
)

var adjectives = []string{
	"Aggravating", "Happy", "Zealous", "Brave", "Calm", "Eager", "Fancy", "Jolly", "Odd", "Silly",
	"Clever", "Wild", "Sleepy", "Radiant", "Frosty", "Fierce", "Proud", "Zany", "Lucky", "Swift",
	"Cosmic", "Lunar", "Stellar", "Electric", "Neon", "Crimson", "Azure", "Golden", "Silver", "Jade",
}

var nouns = []string{
	"Penguin", "Tie", "Ad", "Banana", "Guitar", "Potato", "Tiger", "Kangaroo", "Pancake", "Waffle",
	"Dragon", "Kitten", "Robot", "Wizard", "Ninja", "Cactus", "Mango", "Unicorn", "Pixel", "Comet",
	"Falcon", "Panther", "Octopus", "Walrus", "Rhino", "Grizzly", "Hawk", "Wolf", "Bear", "Shark",
}

var(
	manipalEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+mitblr\d{4}@learner\.manipal\.edu$`)
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func CreateUsername(email, regID string) string {

	adj := adjectives[rand.Intn(len(adjectives))]
	noun := nouns[rand.Intn(len(nouns))]

	num := rand.Intn(9000) + 1000

	return fmt.Sprintf("%s_%s_%d", adj, noun, num)
}

func CheckEmail(email string) bool {
	return manipalEmailRegex.MatchString(email)
}

func CheckPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#~$%^&*()_+|<>?:{}]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSpecial
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), 
		HttpOnly: true,                 
		Secure:   true,                 
		SameSite: http.SameSiteStrictMode,
		Path:     "/",                   
	}

	http.SetCookie(w, cookie)
}

func ClearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Unix(0, 0), 
		HttpOnly: true,                 
		Secure:   true,                 
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	http.SetCookie(w, cookie)
}