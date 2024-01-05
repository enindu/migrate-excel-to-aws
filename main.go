package main

import (
	"flag"
	"html/template"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mail.v2"
	"gorm.io/gorm"
)

const (
	DB_USERNAME       = ""
	DB_PASSWORD       = ""
	DB_PROTOCOL       = "tcp"
	DB_HOST           = ""
	DB_NAME           = ""
	DB_CHARSET        = "utf8mb4"
	DB_LOCALE         = "Local"
	DB_PARSE_TIME     = "true"
	AWS_BUCKET        = ""
	AWS_REGION        = ""
	AWS_ACCESS_KEY    = ""
	AWS_SECRET_KEY    = ""
	SMTP_HOST         = ""
	SMTP_PORT         = 587
	SMTP_USERNAME     = ""
	SMTP_PASSWORD     = ""
	SMTP_FROM_ADDRESS = ""
)

type User struct {
	Id               uint64    `gorm:"column:id;primaryKey"`
	JobCategoryId    uint64    `gorm:"column:job_category_id;default:null"`
	JobSubcategoryId uint64    `gorm:"column:job_subcategory_id;default:null"`
	JobTypeId        uint64    `gorm:"column:job_type_id;default:null"`
	RememberToken    string    `gorm:"column:remember_token;default:null"`
	Name             string    `gorm:"column:name"`
	HospitalName     string    `gorm:"column:hospital_name;default:null"`
	Job              string    `gorm:"column:job;default:null"`
	Email            string    `gorm:"column:email"`
	Phone            string    `gorm:"column:phone;default:null"`
	Website          string    `gorm:"column:website;default:null"`
	Address          string    `gorm:"column:address;default:null"`
	Country          string    `gorm:"column:country;default:null"`
	City             string    `gorm:"column:city;default:null"`
	Age              string    `gorm:"column:age;default:null"`
	Gender           string    `gorm:"column:gender"`
	Languages        string    `gorm:"column:languages;default:null"`
	Bio              string    `gorm:"column:bio;default:null"`
	EducationLevel   string    `gorm:"column:education_level;default:null"`
	ExperienceLevel  string    `gorm:"column:experience_level;default:null"`
	ExpectedLocation string    `gorm:"column:expected_location;default:null"`
	ExpectedSalary   string    `gorm:"column:expected_salary;default:null"`
	Password         string    `gorm:"column:password"`
	Role             string    `gorm:"column:role"`
	Package          string    `gorm:"column:package;default:null"`
	Image            string    `gorm:"column:image;default:null"`
	PackageExpiresAt time.Time `gorm:"column:package_expires_at;default:null"`
	ApprovedAt       time.Time `gorm:"column:approved_at;default:null"`
	EmailVerifiedAt  time.Time `gorm:"column:email_verified_at;default:null"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;default:null"`
	DeletedAt        time.Time `gorm:"column:deleted_at;default:null"`
}

type CandidateCvs struct {
	Id        uint64    `gorm:"column:id;primaryKey"`
	UserId    uint64    `gorm:"column:user_id"`
	File      string    `gorm:"column:file"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:null"`
	DeletedAt time.Time `gorm:"column:deleted_at;default:null"`
}

func main() {
	// Create flags
	fileFlag := flag.String("F", "", "Enter file path. Valid file formats are XLAM, XLSM, XLSX, XLTM and XLTX.")
	sheetFlag := flag.String("S", "", "Enter sheet name.")

	flag.Parse()

	if *fileFlag == "" || *sheetFlag == "" {
		printFlagErrorMessageAndExit()
	}

	// Header message
	startTime := printHeaderMessage()

	// Create database connection
	database := createDatabaseConnection()

	// Create AWS S3 uploader
	awsS3Uploader := createAwsS3Uploader()

	// Get file rows
	fileRows := getFileRows(fileFlag, sheetFlag)

	// Process CVs
	for index, value := range fileRows {
		processCv(database, awsS3Uploader, index, value)
	}

	// Footer message
	printFooterMessage(startTime)
}

func processCv(d *gorm.DB, u *s3manager.Uploader, i int, v []string) {
	// Create downloadable URL of CV
	regex, exception := regexp.Compile(`/d/([^/]+)/`)
	handle(exception)

	regexMatches := regex.FindStringSubmatch(v[6])
	if len(regexMatches) < 2 {
		printRegexErrorMessageAndExit(i)
	}

	fileId := regexMatches[1]
	fileUrl := "https://docs.google.com/uc?id=" + fileId

	// Download CV on local computer
	response, exception := http.Get(fileUrl)
	handle(exception)

	defer response.Body.Close()

	// Upload CV on S3 bucket
	temporaryFileName := uuid.NewString() + ".pdf"

	awsS3UploadInput := &s3manager.UploadInput{
		Bucket:      aws.String(AWS_BUCKET),
		Key:         aws.String("cvs/" + temporaryFileName),
		Body:        response.Body,
		ContentType: aws.String("application/pdf"),
	}

	_, exception = u.Upload(awsS3UploadInput)
	handle(exception)

	// Update database
	id, exception := strconv.ParseUint(v[0], 10, 64)
	handle(exception)

	passwordBytes, exception := bcrypt.GenerateFromPassword([]byte(v[4]), 10)
	handle(exception)

	d.Create(&User{
		Id:         id,
		Name:       v[1],
		Email:      v[2],
		Gender:     v[3],
		Password:   string(passwordBytes),
		Role:       v[5],
		ApprovedAt: time.Now(),
		CreatedAt:  time.Now(),
	})

	d.Create(&CandidateCvs{
		UserId:    id,
		File:      temporaryFileName,
		CreatedAt: time.Now(),
	})

	// Send email
	smtpMessage := mail.NewMessage()
	smtpDialer := mail.NewDialer(SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD)

	smtpMessage.SetHeader("From", SMTP_FROM_ADDRESS)
	smtpMessage.SetHeader("To", v[2])
	smtpMessage.SetHeader("Subject", "Your Profile Added to HealthCareerMe")
	smtpMessage.SetBodyWriter("text/html", func(w io.Writer) error {
		emailTemplate, exception := template.ParseFiles("resources/email-template.html")
		handle(exception)

		return emailTemplate.Execute(w, "")
	})

	exception = smtpDialer.DialAndSend(smtpMessage)
	handle(exception)

	// Record added message
	printRecordAddedMessage(i)
}
