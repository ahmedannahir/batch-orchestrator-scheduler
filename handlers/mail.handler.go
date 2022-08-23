package handlers

import (
	"fmt"
	"gestion-batches/entities"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
)

func SendExecInfosMail(batch entities.Batch, execution entities.Execution, db *gorm.DB) error {
	var profile entities.Profile
	var user entities.User
	err := db.First(&profile, batch.ProfileID).Error
	if err != nil {
		log.Println("Error retrieving profile data : ", err)
		return err
	}
	err = db.First(&user, profile.UserID).Error
	if err != nil {
		log.Println("Error retrieving user data : ", err)
		return err
	}

	switch execution.Status {
	case "1":
		execution.Status = "RUNNING"
	case "2":
		execution.Status = "FAILED"
	case "3":
		execution.Status = "COMPLETED"
	case "4":
		execution.Status = "ABORTED"
	}

	godotenv.Load()
	host := os.Getenv("MAIL_HOST")
	port, _ := strconv.Atoi(os.Getenv("MAIL_PORT"))
	fromEmail := os.Getenv("MAIL_ADDRESS")
	fromPassword := os.Getenv("MAIL_PASSWORD")

	toEmail := user.Email
	subject := fmt.Sprintf("Batch : %s execution update", strings.ToUpper(batch.Name))
	body := fmt.Sprintf("The batch %s is done running.\n Execution details :\n ID : %d\n Status : %s\n Exit code : %s\n Start time : %s\n End time : %s",
		batch.Name, execution.ID, execution.Status, execution.ExitCode, execution.StartTime.Format("2006-01-02 15:04:05"), execution.EndTime.Format("2006-01-02 15:04:05"))

	mail := gomail.NewMessage()
	mail.SetHeader("From", fromEmail)
	mail.SetHeader("To", toEmail)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/plain", body)

	dialer := gomail.NewDialer(host, port, fromEmail, fromPassword)

	log.Println("Sending email about batch ID : ", batch.ID, " to : ", user.Email)

	err = dialer.DialAndSend(mail)
	if err != nil {
		log.Println("Error sending email : ", err)
		return err
	}

	return nil
}
