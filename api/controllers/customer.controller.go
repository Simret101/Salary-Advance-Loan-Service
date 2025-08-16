package controllers

import (
	"SalaryAdvance/internal/usecases"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CustomerController struct {
	uc *usecases.CustomerUseCase
}

func NewCustomerController(uc *usecases.CustomerUseCase) *CustomerController {
	return &CustomerController{uc: uc}
}

func (ctrl *CustomerController) ImportCustomers(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get file from form"})
		return
	}
	defer file.Close()

	ctx := c.Request.Context()
	customers, logs, err := ctrl.uc.ImportCustomers(ctx, file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Customers imported",
		"data":    customers,
		"logs":    logs,
	})
}

func (ctrl *CustomerController) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()
	customer, err := ctrl.uc.GetCustomer(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (ctrl *CustomerController) GetAllCustomers(c *gin.Context) {
	ctx := c.Request.Context()
	customers, err := ctrl.uc.GetAllCustomers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}
