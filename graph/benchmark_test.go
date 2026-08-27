package graph

import (
	cmd "downhill-api/cmd/api"
	"downhill-api/graph/model"
	"encoding/json"
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBenchmarkRunner(t *testing.T) {
	// Marker test for go test discovery
}

func BenchmarkPasswordHashingCost14(b *testing.B) {
	password := "SecretPass123!@"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cmd.HashPassword(password)
	}
}

func BenchmarkPasswordHashingCost10(b *testing.B) {
	password := "SecretPass123!@"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bcrypt.GenerateFromPassword([]byte(password), 10)
	}
}

func BenchmarkPasswordCheck(b *testing.B) {
	password := "StrongP@ssw0rd!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.CheckPassword(password)
	}
}

func BenchmarkEmailRegex(b *testing.B) {
	email := "user.mitblr2024@learner.manipal.edu"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cmd.CheckEmail(email)
	}
}

func BenchmarkJWTTokenCreation(b *testing.B) {
	os.Setenv("JWT_SECRET_KEY", "58d11e13d5a80beb9578d23de2173966cb059dfb7ae3627c1619bd6ab07b4511")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = createToken("12345")
	}
}

func BenchmarkJSONMarshalCompanySlice(b *testing.B) {
	companies := make([]*model.Company, 50)
	for i := 0; i < 50; i++ {
		companies[i] = &model.Company{
			ID:          "100",
			CompanyName: "Benchmark Technologies Inc.",
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(companies)
	}
}

func BenchmarkJSONUnmarshalCompanySlice(b *testing.B) {
	companies := make([]*model.Company, 50)
	for i := 0; i < 50; i++ {
		companies[i] = &model.Company{
			ID:          "100",
			CompanyName: "Benchmark Technologies Inc.",
		}
	}
	data, _ := json.Marshal(companies)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res []*model.Company
		_ = json.Unmarshal(data, &res)
	}
}
