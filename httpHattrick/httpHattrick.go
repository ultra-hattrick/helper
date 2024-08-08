package httphattrick

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/dghubble/oauth1"
	"github.com/ultra-hattrick/helper/utilsConstants"
	"gitlab.com/uchile1/helper/helperCommon"
	"gitlab.com/uchile1/helper/helperLog"
	"gorm.io/gorm"
)

type UserRegister struct {
	ID             uint
	Role           *string
	AccessToken    *string
	AccessSecret   *string
	IsAccessFinish bool
}

type HattrickCHPP struct {
	db             *gorm.DB
	userRegisterID *uint
}

// NewHattrickUtils constructor para crear una nueva instancia de HattrickUtils
func NewHattrickCHPP(debe *gorm.DB, userRegisterID *uint) *HattrickCHPP {
	return &HattrickCHPP{db: debe, userRegisterID: userRegisterID}
}

func (h *HattrickCHPP) getClientByTokens(isPermitedTokenAdmin bool) (*http.Client, error) {
	config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET"))
	var userRegister *UserRegister
	if h.userRegisterID != nil {
		err := h.db.
			Where(&UserRegister{ID: *h.userRegisterID, IsAccessFinish: true}).
			Where("role in (?)", []string{utilsConstants.ROLE_ADMIN, utilsConstants.ROLE_USER_TRIAL, utilsConstants.ROLE_USER_PREMIUM}).
			First(&userRegister).Error
		if err != nil {
			return nil, err
		}
	} else {
		if !isPermitedTokenAdmin {
			return nil, fmt.Errorf("resource protected, not found userRegister to get admin token")
		}
		role := utilsConstants.ROLE_ADMIN
		err := h.db.Where(&UserRegister{Role: &role}).First(&userRegister).Error
		if err != nil {
			return nil, err
		}
	}

	if userRegister.AccessToken == nil || userRegister.AccessSecret == nil {
		return nil, fmt.Errorf("tokens access not found or inactive for userRegisterID: %d to invoke APIs CHPP of Hattrick", h.userRegisterID)
	}
	client := config.Client(oauth1.NoContext, oauth1.NewToken(*userRegister.AccessToken, *userRegister.AccessSecret))
	return client, nil
}

func (h *HattrickCHPP) GetResultsFromHattrick(pathHattrick string, v any) error {
	path := fmt.Sprintf("%s%s", os.Getenv("BASE_RESOURCE_URL"), pathHattrick)
	client, err := h.getClientByTokens(isHattrickResourceThatAdminCanGet(pathHattrick))
	if err != nil {
		return err
	}
	resp, err := client.Get(path)
	helperLog.Logger.Warn().Str(
		"function", helperCommon.GetFrame(2).Function,
	).Msgf("Se ocupa API Hattrick url: %s", path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Imprimir el status code
	helperLog.Logger.Debug().Msgf("HTTP Status Code: %d para la url: %s", resp.StatusCode, path)
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP Status Code: %d para la url: %s", resp.StatusCode, path)
	}

	// Leer el contenido del body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	helperLog.Logger.Debug().Msgf("Response Body: %s", string(body))

	// Reiniciar el cuerpo de la respuesta para que pueda ser decodificado
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	err = xml.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return err
	}
	// helperLog.Logger.Debug().Msgf("--->Arena: %v", hattrickData.Match.DetailsMatch.Arena)
	return nil
}

func isHattrickResourceThatAdminCanGet(path string) bool {
	switch {
	case strings.Contains(path, "file=teamdetails"):
		return true
	case strings.Contains(path, "file=matches"):
		return true
	case strings.Contains(path, "file=matchdetails"):
		return true
	case strings.Contains(path, "file=matchesarchive"):
		return true
	case strings.Contains(path, "file=matchlineup"):
		return true
	default:
		return false
	}
}
