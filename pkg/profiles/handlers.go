package profiles

import (
	"errors"
	"net/http"

	"github.com/Angak0k/pimpmypack/pkg/helper"
	"github.com/gin-gonic/gin"
)

// GetPublicProfile returns the public profile and shared packs for a username
// @Summary Get public user profile
// @Description Returns public profile information and shared packs for a user.
// @Description Returns 404 for non-existent users or users without a public profile (anti-enumeration).
// @Tags Profiles
// @Produce json
// @Param username path string true "Username"
// @Success 200 {object} PublicProfile
// @Failure 404 {object} apitypes.ErrorResponse "Profile not found"
// @Failure 500 {object} apitypes.ErrorResponse "Internal Server Error"
// @Router /user/{username} [get]
func GetPublicProfile(c *gin.Context) {
	username := c.Param("username")

	profile, err := returnPublicProfileByUsername(c.Request.Context(), username)
	if err != nil {
		if errors.Is(err, ErrProfileNotFound) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
			return
		}
		helper.LogAndSanitize(err, "get public profile: fetch profile failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	sharedPacks, err := returnSharedPacksByUsername(c.Request.Context(), username)
	if err != nil {
		helper.LogAndSanitize(err, "get public profile: fetch shared packs failed")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": helper.ErrMsgInternalServer})
		return
	}

	profile.SharedPacks = sharedPacks

	c.IndentedJSON(http.StatusOK, profile)
}
