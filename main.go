package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

type Invoice struct {
	InvoiceId   int    `json:"invoiceId" binding:"gte=0"`
	CustomerId  int    `json:"customerId" binding:"required,gte=0"`
	Price       int    `json:"price" binding:"required,gte=0"`
	Description string `json:"description" binding:"required"`
}
type PrintJob struct {
	JobId     int    `json:"jobId"`
	InvoiceId int    `json:"invoiceId"`
	Format    string `json:"format"`
}

func createPrintJob(invoiceId int) bool {
	client := resty.New()
	var p PrintJob
	// Call PrinterService via RESTful interface
	resp, err := client.R().
		SetBody(PrintJob{Format: "A4", InvoiceId: invoiceId}).
		SetResult(&p).
		Post("http://localhost:60001/v2/print-invoice")

	if err != nil {
		log.Println("InvoiceGenerator: unable to connect PrinterService")
		return false
	}

	if !resp.IsSuccess() {
		log.Printf("InvoiceGenerator: PrinterService has sent a response with status: %v", resp.StatusCode())
		return false
	}

	log.Printf("InvoiceGenerator: created print job #%v via PrinterService", p.JobId)
	return true
}
func main() {
	router := gin.Default()
	router.POST("/invoices", func(c *gin.Context) {
		var iv Invoice
		if err := c.ShouldBindJSON(&iv); err != nil {
			c.JSON(400, gin.H{"error": "Invalid input!"})
			return
		}
		log.Println("InvoiceGenerator: creating new invoice...")
		rand.Seed(time.Now().UnixNano())
		iv.InvoiceId = rand.Intn(1000)
		log.Printf("InvoiceGenerator: created invoice #%v", iv.InvoiceId)

		if createPrintJob(iv.InvoiceId) {
			c.JSON(200, iv)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Sending request to PrintService error!"})
		}
	})
	router.Run(":6000")
}
