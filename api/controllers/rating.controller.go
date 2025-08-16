package controllers

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RatingController struct {
	ratingUseCase domain.RatingUseCase
}

func NewRatingController(ruc domain.RatingUseCase) *RatingController {
	return &RatingController{ratingUseCase: ruc}
}


func (rc *RatingController) GetRating(c *gin.Context) {
	customerID := c.Param("customerID")
	rating, err := rc.ratingUseCase.CalculateRating(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rating})
}


func (rc *RatingController) GetAllRatings(c *gin.Context) {
	ratings, err := rc.ratingUseCase.GetAllRatings(c.Request.Context())
	if err != nil {
		c.JSON(config.GetStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ratings})
}
