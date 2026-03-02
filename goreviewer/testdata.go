package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type UselessData struct {
	ID        int
	Name      string
	Email     string
	Age       int
	Address   string
	Phone     string
	Country   string
	City      string
	ZipCode   string
	CreatedAt time.Time
}

func (u *UselessData) String() string {
	return fmt.Sprintf("ID: %d, Name: %s, Email: %s, Age: %d", u.ID, u.Name, u.Email, u.Age)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func generateRandomEmail() string {
	domains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "example.com"}
	username := generateRandomString(8)
	domain := domains[rand.Intn(len(domains))]
	return fmt.Sprintf("%s@%s", username, domain)
}

func generateRandomAddress() string {
	streets := []string{"Main St", "Oak Ave", "Maple Dr", "Cedar Ln", "Pine Rd", "Elm St", "Washington Blvd", "Park Ave"}
	cities := []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego"}
	states := []string{"NY", "CA", "IL", "TX", "AZ", "PA", "TX", "CA"}
	zipCodes := []string{"10001", "90001", "60601", "77001", "85001", "19101", "78201", "92101"}

	street := streets[rand.Intn(len(streets))]
	number := rand.Intn(9999) + 1
	city := cities[rand.Intn(len(cities))]
	state := states[rand.Intn(len(states))]
	zip := zipCodes[rand.Intn(len(zipCodes))]

	return fmt.Sprintf("%d %s, %s, %s %s", number, street, city, state, zip)
}

func processUselessData(data []UselessData) []UselessData {
	result := make([]UselessData, 0)
	for _, d := range data {
		if d.Age > 0 && d.Age < 150 {
			result = append(result, d)
		}
	}
	return result
}

func calculateAverageAge(data []UselessData) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0
	for _, d := range data {
		sum += d.Age
	}
	return float64(sum) / float64(len(data))
}

func filterByCountry(data []UselessData, country string) []UselessData {
	result := make([]UselessData, 0)
	for _, d := range data {
		if d.Country == country {
			result = append(result, d)
		}
	}
	return result
}

func searchByName(data []UselessData, name string) []UselessData {
	result := make([]UselessData, 0)
	lowerName := strings.ToLower(name)
	for _, d := range data {
		if strings.Contains(strings.ToLower(d.Name), lowerName) {
			result = append(result, d)
		}
	}
	return result
}

func sortByAge(data []UselessData) []UselessData {
	sorted := make([]UselessData, len(data))
	copy(sorted, data)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Age > sorted[j].Age {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func groupByCity(data []UselessData) map[string][]UselessData {
	groups := make(map[string][]UselessData)
	for _, d := range data {
		groups[d.City] = append(groups[d.City], d)
	}
	return groups
}

func calculateStats(data []UselessData) map[string]interface{} {
	stats := make(map[string]interface{})

	ages := make([]int, 0)
	for _, d := range data {
		ages = append(ages, d.Age)
	}

	sum := 0
	minAge := 999
	maxAge := 0
	for _, age := range ages {
		sum += age
		if age < minAge {
			minAge = age
		}
		if age > maxAge {
			maxAge = age
		}
	}

	avgAge := 0.0
	if len(ages) > 0 {
		avgAge = float64(sum) / float64(len(ages))
	}

	variance := 0.0
	for _, age := range ages {
		diff := float64(age) - avgAge
		variance += diff * diff
	}
	if len(ages) > 0 {
		variance /= float64(len(ages))
	}
	stdDev := math.Sqrt(variance)

	stats["count"] = len(data)
	stats["average_age"] = avgAge
	stats["min_age"] = minAge
	stats["max_age"] = maxAge
	stats["std_dev"] = stdDev
	stats["total_age"] = sum

	return stats
}

func validateEmail(email string) bool {
	if !strings.Contains(email, "@") {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}
	if !strings.Contains(parts[1], ".") {
		return false
	}
	return true
}

func normalizePhone(phone string) string {
	result := ""
	for _, c := range phone {
		if c >= '0' && c <= '9' {
			result += string(c)
		}
	}
	return result
}

func generateSampleData(count int) []UselessData {
	firstNames := []string{"John", "Jane", "Bob", "Alice", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}
	countries := []string{"USA", "Canada", "UK", "Germany", "France", "Japan", "Australia", "Brazil", "India", "China"}

	data := make([]UselessData, count)
	for i := 0; i < count; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		fullName := fmt.Sprintf("%s %s", firstName, lastName)

		data[i] = UselessData{
			ID:        i + 1,
			Name:      fullName,
			Email:     generateRandomEmail(),
			Age:       rand.Intn(80) + 18,
			Address:   generateRandomAddress(),
			Phone:     fmt.Sprintf("+1%d", rand.Intn(10000000000)),
			Country:   countries[rand.Intn(len(countries))],
			City:      strings.Split(generateRandomAddress(), ",")[1],
			ZipCode:   fmt.Sprintf("%d", rand.Intn(90000)+10000),
			CreatedAt: time.Now().Add(-time.Duration(rand.Intn(365)) * 24 * time.Hour),
		}
	}

	return data
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RunUselessTests() {
	fmt.Println("Starting useless data generator...")

	data := generateSampleData(1000)

	fmt.Println("Generated sample data")

	processed := processUselessData(data)
	fmt.Printf("Processed %d records\n", len(processed))

	avgAge := calculateAverageAge(processed)
	fmt.Printf("Average age: %.2f\n", avgAge)

	usaData := filterByCountry(processed, "USA")
	fmt.Printf("USA users: %d\n", len(usaData))

	stats := calculateStats(processed)
	fmt.Printf("Stats: %v\n", stats)

	groups := groupByCity(processed)
	fmt.Printf("Cities: %d\n", len(groups))

	sorted := sortByAge(processed)
	fmt.Printf("Youngest: %s, Oldest: %s\n", sorted[0].Name, sorted[len(sorted)-1].Name)

	_ = validateEmail
	_ = normalizePhone
	_ = searchByName

	fmt.Println("Done!")
}
