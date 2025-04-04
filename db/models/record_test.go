// db/models/record_test.go
package models_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/dukerupert/doxie-discs/db/models"
)

var (
	testDB        *sql.DB
	recordService *models.RecordService
	artistService *models.ArtistService
	genreService  *models.GenreService
	labelService  *models.LabelService
	testUserID    int
	testArtistID  int
	testGenreID   int
	testLabelID   int
)

func TestMain(m *testing.M) {
	// Load environment variables from .env.test file
	err := godotenv.Load("../../.env.test")
	if err != nil {
		log.Println("No .env.test file found, using environment variables")
	}

	// Set up test database connection
	testDBURL := os.Getenv("TEST_DB_URL")
	if testDBURL == "" {
		// Fallback to constructing URL from individual params
		testDBURL = os.Getenv("DB_URL")
		if testDBURL == "" {
			testDBURL = "postgres://postgres:postgres@localhost:5432/golang_app_test?sslmode=disable"
		}
	}

	testDB, err = sql.Open("postgres", testDBURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	// Create services
	recordService = models.NewRecordService(testDB)
	artistService = models.NewArtistService(testDB)
	genreService = models.NewGenreService(testDB)
	labelService = models.NewLabelService(testDB)

	// Clean up database before tests
	cleanDatabase()

	// Create test user
	testUserID = setupTestUser()

	// Create test data
	testArtistID = setupTestArtist()
	testGenreID = setupTestGenre()
	testLabelID = setupTestLabel()

	// Run tests
	exitCode := m.Run()

	// Clean up after tests
	cleanDatabase()

	os.Exit(exitCode)
}

func cleanDatabase() {
	// Order matters due to foreign key constraints
	testDB.Exec("DELETE FROM tracks")
	testDB.Exec("DELETE FROM artist_record")
	testDB.Exec("DELETE FROM genre_record")
	testDB.Exec("DELETE FROM records")
	testDB.Exec("DELETE FROM artists")
	testDB.Exec("DELETE FROM genres")
	testDB.Exec("DELETE FROM labels")
	testDB.Exec("DELETE FROM users")
}

func setupTestUser() int {
	var userID int
	err := testDB.QueryRow(`
        INSERT INTO users (email, password_hash, name)
        VALUES ('test@example.com', 'hashedpassword', 'Test User')
        RETURNING id
    `).Scan(&userID)

	if err != nil {
		log.Fatalf("Failed to create test user: %v", err)
	}
	return userID
}

func setupTestArtist() int {
	var artistID int
	err := testDB.QueryRow(`
        INSERT INTO artists (name, description, user_id)
        VALUES ('Test Artist', 'Test artist description', $1)
        RETURNING id
    `, testUserID).Scan(&artistID)

	if err != nil {
		log.Fatalf("Failed to create test artist: %v", err)
	}
	return artistID
}

func setupTestGenre() int {
	var genreID int
	err := testDB.QueryRow(`
        INSERT INTO genres (name, description, user_id)
        VALUES ('Test Genre', 'Test genre description', $1)
        RETURNING id
    `, testUserID).Scan(&genreID)

	if err != nil {
		log.Fatalf("Failed to create test genre: %v", err)
	}
	return genreID
}

func setupTestLabel() int {
	var labelID int
	err := testDB.QueryRow(`
        INSERT INTO labels (name, description, user_id)
        VALUES ('Test Label', 'Test label description', $1)
        RETURNING id
    `, testUserID).Scan(&labelID)

	if err != nil {
		log.Fatalf("Failed to create test label: %v", err)
	}
	return labelID
}

// Test creating a record
func TestCreateRecord(t *testing.T) {
	// Create a new record
	record := &models.Record{
		Title:           "Test Album",
		ReleaseYear:     2020,
		CatalogNumber:   "TEST-001",
		Condition:       "Mint",
		Notes:           "Test notes",
		StorageLocation: "Shelf A",
		UserID:          testUserID,
		LabelID:         testLabelID,
		Artists: []models.Artist{
			{ID: testArtistID, Role: "Primary Artist"},
		},
		Genres: []models.Genre{
			{ID: testGenreID},
		},
		Tracks: []models.Track{
			{
				Title:    "Track 1",
				Duration: "3:45",
				Position: "A1",
			},
			{
				Title:    "Track 2",
				Duration: "4:20",
				Position: "A2",
			},
		},
	}

	createdRecord, err := recordService.Create(record)
	if err != nil {
		t.Fatalf("Failed to create record: %v", err)
	}

	// Verify the record was created
	if createdRecord.ID <= 0 {
		t.Errorf("Expected record ID to be greater than 0, got %d", createdRecord.ID)
	}

	if createdRecord.Title != "Test Album" {
		t.Errorf("Expected record title to be 'Test Album', got '%s'", createdRecord.Title)
	}

	if createdRecord.ReleaseYear != 2020 {
		t.Errorf("Expected release year to be 2020, got %d", createdRecord.ReleaseYear)
	}

	if createdRecord.UserID != testUserID {
		t.Errorf("Expected user ID to be %d, got %d", testUserID, createdRecord.UserID)
	}

	// Verify tracks were created
	if len(createdRecord.Tracks) != 2 {
		t.Errorf("Expected 2 tracks, got %d", len(createdRecord.Tracks))
	}

	// Clean up
	testDB.Exec("DELETE FROM tracks WHERE record_id = $1", createdRecord.ID)
	testDB.Exec("DELETE FROM artist_record WHERE record_id = $1", createdRecord.ID)
	testDB.Exec("DELETE FROM genre_record WHERE record_id = $1", createdRecord.ID)
	testDB.Exec("DELETE FROM records WHERE id = $1", createdRecord.ID)
}

// Test retrieving a record by ID
func TestGetRecordByID(t *testing.T) {
	// First create a record to retrieve
	var recordID int
	err := testDB.QueryRow(`
        INSERT INTO records (title, release_year, catalog_number, user_id, label_id)
        VALUES ('Get Test Album', 2021, 'TEST-002', $1, $2)
        RETURNING id
    `, testUserID, testLabelID).Scan(&recordID)
	if err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}

	// Add artist relationship
	_, err = testDB.Exec(
		"INSERT INTO artist_record (artist_id, record_id, role) VALUES ($1, $2, $3)",
		testArtistID, recordID, "Primary Artist",
	)
	if err != nil {
		t.Fatalf("Failed to add artist to record: %v", err)
	}

	// Add genre relationship
	_, err = testDB.Exec(
		"INSERT INTO genre_record (genre_id, record_id) VALUES ($1, $2)",
		testGenreID, recordID,
	)
	if err != nil {
		t.Fatalf("Failed to add genre to record: %v", err)
	}

	// Add a track
	_, err = testDB.Exec(
		"INSERT INTO tracks (title, duration, position, record_id) VALUES ($1, $2, $3, $4)",
		"Test Track", "3:30", "A1", recordID,
	)
	if err != nil {
		t.Fatalf("Failed to add track to record: %v", err)
	}

	// Now retrieve the record
	record, err := recordService.GetByID(recordID)
	if err != nil {
		t.Fatalf("Failed to get record by ID: %v", err)
	}

	// Verify record details
	if record.ID != recordID {
		t.Errorf("Expected record ID %d, got %d", recordID, record.ID)
	}

	if record.Title != "Get Test Album" {
		t.Errorf("Expected title 'Get Test Album', got '%s'", record.Title)
	}

	if record.ReleaseYear != 2021 {
		t.Errorf("Expected release year 2021, got %d", record.ReleaseYear)
	}

	if record.CatalogNumber != "TEST-002" {
		t.Errorf("Expected catalog number 'TEST-002', got '%s'", record.CatalogNumber)
	}

	// Verify relationships
	if len(record.Artists) != 1 {
		t.Errorf("Expected 1 artist, got %d", len(record.Artists))
	} else if record.Artists[0].ID != testArtistID {
		t.Errorf("Expected artist ID %d, got %d", testArtistID, record.Artists[0].ID)
	}

	if len(record.Genres) != 1 {
		t.Errorf("Expected 1 genre, got %d", len(record.Genres))
	} else if record.Genres[0].ID != testGenreID {
		t.Errorf("Expected genre ID %d, got %d", testGenreID, record.Genres[0].ID)
	}

	if len(record.Tracks) != 1 {
		t.Errorf("Expected 1 track, got %d", len(record.Tracks))
	} else if record.Tracks[0].Title != "Test Track" {
		t.Errorf("Expected track title 'Test Track', got '%s'", record.Tracks[0].Title)
	}

	// Clean up
	testDB.Exec("DELETE FROM tracks WHERE record_id = $1", recordID)
	testDB.Exec("DELETE FROM artist_record WHERE record_id = $1", recordID)
	testDB.Exec("DELETE FROM genre_record WHERE record_id = $1", recordID)
	testDB.Exec("DELETE FROM records WHERE id = $1", recordID)
}

// Test listing records by user ID
func TestListRecordsByUserID(t *testing.T) {
	// Create test records
	recordIDs := make([]int, 3)
	for i := 0; i < 3; i++ {
		var recordID int
		err := testDB.QueryRow(`
            INSERT INTO records (title, release_year, user_id, label_id)
            VALUES ($1, $2, $3, $4)
            RETURNING id
        `, "List Test "+string(rune(65+i)), 2000+i, testUserID, testLabelID).Scan(&recordID)
		if err != nil {
			t.Fatalf("Failed to create test record %d: %v", i, err)
		}
		recordIDs[i] = recordID
	}

	// Now list records for the user
	records, err := recordService.ListByUserID(testUserID)
	if err != nil {
		t.Fatalf("Failed to list records by user ID: %v", err)
	}

	// There should be at least 3 records (we may have leftover records from previous tests)
	if len(records) < 3 {
		t.Errorf("Expected at least 3 records, got %d", len(records))
	}

	// Clean up
	for _, id := range recordIDs {
		testDB.Exec("DELETE FROM records WHERE id = $1", id)
	}
}

// Test updating a record
func TestUpdateRecord(t *testing.T) {
	// First create a record to update
	var recordID int
	err := testDB.QueryRow(`
        INSERT INTO records (title, release_year, catalog_number, user_id, label_id)
        VALUES ('Update Test Album', 2022, 'TEST-003', $1, $2)
        RETURNING id
    `, testUserID, testLabelID).Scan(&recordID)
	if err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}

	// Add relationships for initial state
	_, err = testDB.Exec(
		"INSERT INTO artist_record (artist_id, record_id, role) VALUES ($1, $2, $3)",
		testArtistID, recordID, "Primary Artist",
	)
	if err != nil {
		t.Fatalf("Failed to add artist to record: %v", err)
	}

	_, err = testDB.Exec(
		"INSERT INTO genre_record (genre_id, record_id) VALUES ($1, $2)",
		testGenreID, recordID,
	)
	if err != nil {
		t.Fatalf("Failed to add genre to record: %v", err)
	}

	// Now update the record
	updatedRecord := &models.Record{
		ID:              recordID,
		Title:           "Updated Album Title",
		ReleaseYear:     2023,
		CatalogNumber:   "TEST-003-UPDATED",
		Condition:       "Very Good",
		Notes:           "Updated test notes",
		StorageLocation: "Shelf B",
		UserID:          testUserID,
		LabelID:         testLabelID,
		// Change relationships - keep same IDs but we'll update the artist role
		Artists: []models.Artist{
			{ID: testArtistID, Role: "Producer"},
		},
		Genres: []models.Genre{
			{ID: testGenreID},
		},
		// Add tracks
		Tracks: []models.Track{
			{
				Title:    "Updated Track 1",
				Duration: "5:15",
				Position: "B1",
			},
		},
	}

	result, err := recordService.Update(updatedRecord)
	if err != nil {
		t.Fatalf("Failed to update record: %v", err)
	}

	// Verify updated fields
	if result.Title != "Updated Album Title" {
		t.Errorf("Expected title 'Updated Album Title', got '%s'", result.Title)
	}

	if result.ReleaseYear != 2023 {
		t.Errorf("Expected release year 2023, got %d", result.ReleaseYear)
	}

	if result.CatalogNumber != "TEST-003-UPDATED" {
		t.Errorf("Expected catalog number 'TEST-003-UPDATED', got '%s'", result.CatalogNumber)
	}

	// Now retrieve the record to check relationships
	retrievedRecord, err := recordService.GetByID(recordID)
	if err != nil {
		t.Fatalf("Failed to get updated record: %v", err)
	}

	// Check artist role is updated
	if len(retrievedRecord.Artists) != 1 {
		t.Errorf("Expected 1 artist, got %d", len(retrievedRecord.Artists))
	} else if retrievedRecord.Artists[0].Role != "Producer" {
		t.Errorf("Expected artist role 'Producer', got '%s'", retrievedRecord.Artists[0].Role)
	}

	// Check tracks
	if len(retrievedRecord.Tracks) != 1 {
		t.Errorf("Expected 1 track, got %d", len(retrievedRecord.Tracks))
	} else if retrievedRecord.Tracks[0].Title != "Updated Track 1" {
		t.Errorf("Expected track title 'Updated Track 1', got '%s'", retrievedRecord.Tracks[0].Title)
	}

	// Clean up
	testDB.Exec("DELETE FROM tracks WHERE record_id = $1", recordID)
	testDB.Exec("DELETE FROM artist_record WHERE record_id = $1", recordID)
	testDB.Exec("DELETE FROM genre_record WHERE record_id = $1", recordID)
	testDB.Exec("DELETE FROM records WHERE id = $1", recordID)
}

// Test deleting a record
func TestDeleteRecord(t *testing.T) {
	// First create a record to delete
	var recordID int
	err := testDB.QueryRow(`
        INSERT INTO records (title, release_year, user_id)
        VALUES ('Delete Test Album', 2024, $1)
        RETURNING id
    `, testUserID).Scan(&recordID)
	if err != nil {
		t.Fatalf("Failed to create test record: %v", err)
	}

	// Add a track, artist, and genre for complete deletion test
	_, err = testDB.Exec(
		"INSERT INTO tracks (title, position, record_id) VALUES ($1, $2, $3)",
		"Delete Test Track", "A1", recordID,
	)
	if err != nil {
		t.Fatalf("Failed to add track to record: %v", err)
	}

	_, err = testDB.Exec(
		"INSERT INTO artist_record (artist_id, record_id) VALUES ($1, $2)",
		testArtistID, recordID,
	)
	if err != nil {
		t.Fatalf("Failed to add artist to record: %v", err)
	}

	_, err = testDB.Exec(
		"INSERT INTO genre_record (genre_id, record_id) VALUES ($1, $2)",
		testGenreID, recordID,
	)
	if err != nil {
		t.Fatalf("Failed to add genre to record: %v", err)
	}

	// Now delete the record
	err = recordService.Delete(recordID)
	if err != nil {
		t.Fatalf("Failed to delete record: %v", err)
	}

	// Verify the record is deleted
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM records WHERE id = $1", recordID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check record deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected record to be deleted, but it still exists")
	}

	// Verify related tracks are deleted (due to ON DELETE CASCADE)
	err = testDB.QueryRow("SELECT COUNT(*) FROM tracks WHERE record_id = $1", recordID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check tracks deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected tracks to be deleted, but they still exist")
	}

	// Verify artist_record relations are deleted
	err = testDB.QueryRow("SELECT COUNT(*) FROM artist_record WHERE record_id = $1", recordID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check artist_record deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected artist_record relations to be deleted, but they still exist")
	}

	// Verify genre_record relations are deleted
	err = testDB.QueryRow("SELECT COUNT(*) FROM genre_record WHERE record_id = $1", recordID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check genre_record deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected genre_record relations to be deleted, but they still exist")
	}
}

// Test searching for records
func TestSearchRecords(t *testing.T) {
	// Create records with specific attributes for search testing
	recordIDs := make([]int, 3)
	
	// Record 1: Specific title and location
	var recordID1 int
	err := testDB.QueryRow(`
        INSERT INTO records (title, release_year, storage_location, user_id, label_id)
        VALUES ('SEARCHABLE Jazz Collection', 2010, 'Top Shelf', $1, $2)
        RETURNING id
    `, testUserID, testLabelID).Scan(&recordID1)
	if err != nil {
		t.Fatalf("Failed to create search test record 1: %v", err)
	}
	recordIDs[0] = recordID1
	
	// Add a specific genre to record 1
	var jazzGenreID int
	err = testDB.QueryRow(`
        INSERT INTO genres (name, user_id)
        VALUES ('SEARCHABLE Jazz', $1)
        RETURNING id
    `, testUserID).Scan(&jazzGenreID)
	if err != nil {
		t.Fatalf("Failed to create jazz genre: %v", err)
	}
	
	_, err = testDB.Exec(
		"INSERT INTO genre_record (genre_id, record_id) VALUES ($1, $2)",
		jazzGenreID, recordID1,
	)
	if err != nil {
		t.Fatalf("Failed to add genre to record: %v", err)
	}
	
	// Record 2: Different title but same artist
	var recordID2 int
	err = testDB.QueryRow(`
        INSERT INTO records (title, release_year, storage_location, user_id)
        VALUES ('Rock Anthology', 2015, 'Bottom Shelf', $1)
        RETURNING id
    `, testUserID).Scan(&recordID2)
	if err != nil {
		t.Fatalf("Failed to create search test record 2: %v", err)
	}
	recordIDs[1] = recordID2
	
	// Create a specific artist for search
	var searchArtistID int
	err = testDB.QueryRow(`
        INSERT INTO artists (name, user_id)
        VALUES ('SEARCHABLE Artist', $1)
        RETURNING id
    `, testUserID).Scan(&searchArtistID)
	if err != nil {
		t.Fatalf("Failed to create search artist: %v", err)
	}
	
	// Add this artist to both records
	_, err = testDB.Exec(
		"INSERT INTO artist_record (artist_id, record_id, role) VALUES ($1, $2, $3)",
		searchArtistID, recordID1, "Featured",
	)
	if err != nil {
		t.Fatalf("Failed to add artist to record 1: %v", err)
	}
	
	_, err = testDB.Exec(
		"INSERT INTO artist_record (artist_id, record_id, role) VALUES ($1, $2, $3)",
		searchArtistID, recordID2, "Primary",
	)
	if err != nil {
		t.Fatalf("Failed to add artist to record 2: %v", err)
	}
	
	// Record 3: With specific label
	var recordID3 int
	var searchLabelID int
	
	err = testDB.QueryRow(`
        INSERT INTO labels (name, user_id)
        VALUES ('SEARCHABLE Records', $1)
        RETURNING id
    `, testUserID).Scan(&searchLabelID)
	if err != nil {
		t.Fatalf("Failed to create search label: %v", err)
	}
	
	err = testDB.QueryRow(`
        INSERT INTO records (title, release_year, user_id, label_id)
        VALUES ('Classical Compilation', 2020, $1, $2)
        RETURNING id
    `, testUserID, searchLabelID).Scan(&recordID3)
	if err != nil {
		t.Fatalf("Failed to create search test record 3: %v", err)
	}
	recordIDs[2] = recordID3
	
	// Test 1: Search by keyword in title
	records, err := recordService.Search(testUserID, "SEARCHABLE", "", "", "", "")
	if err != nil {
		t.Fatalf("Failed to search records by title: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record when searching for 'SEARCHABLE' in title, got %d", len(records))
	}
	
	// Test 2: Search by artist
	records, err = recordService.Search(testUserID, "", "SEARCHABLE", "", "", "")
	if err != nil {
		t.Fatalf("Failed to search records by artist: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("Expected 2 records when searching for 'SEARCHABLE' in artist, got %d", len(records))
	}
	
	// Test 3: Search by genre
	records, err = recordService.Search(testUserID, "", "", "Jazz", "", "")
	if err != nil {
		t.Fatalf("Failed to search records by genre: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record when searching for 'Jazz' in genre, got %d", len(records))
	}
	
	// Test 4: Search by label
	records, err = recordService.Search(testUserID, "", "", "", "SEARCHABLE", "")
	if err != nil {
		t.Fatalf("Failed to search records by label: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record when searching for 'SEARCHABLE' in label, got %d", len(records))
	}
	
	// Test 5: Search by storage location
	records, err = recordService.Search(testUserID, "", "", "", "", "Top")
	if err != nil {
		t.Fatalf("Failed to search records by location: %v", err)
	}
	if len(records) != 1 {
		t.Errorf("Expected 1 record when searching for 'Top' in location, got %d", len(records))
	}
	
	// Clean up
	for _, id := range recordIDs {
		testDB.Exec("DELETE FROM tracks WHERE record_id = $1", id)
		testDB.Exec("DELETE FROM artist_record WHERE record_id = $1", id)
		testDB.Exec("DELETE FROM genre_record WHERE record_id = $1", id)
		testDB.Exec("DELETE FROM records WHERE id = $1", id)
	}
	testDB.Exec("DELETE FROM artists WHERE name = 'SEARCHABLE Artist'")
	testDB.Exec("DELETE FROM genres WHERE name = 'SEARCHABLE Jazz'")
	testDB.Exec("DELETE FROM labels WHERE name = 'SEARCHABLE Records'")
}