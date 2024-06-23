package helperHT

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/dghubble/oauth1"
	"github.com/gin-gonic/gin"
	"gitlab.com/uchile1/helper/helperCommon"
	"gitlab.com/uchile1/helper/helperLog"
)

func IsValidWeeks(weeks int) bool {
	validWeeks := []int{3, 5, 10, 15, 20}
	for _, w := range validWeeks {
		if w == weeks {
			return true
		}
	}
	return false
}

func IsValidStadium(stadium int) bool {
	valid := []int{1, 2}
	for _, w := range valid {
		if w == stadium {
			return true
		}
	}
	return false
}

// helper para obtener y validar un booleano de la consulta
func GetQueryBool(c *gin.Context, key string, defaultValue bool) (bool, error) {
	v, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, nil
	}
	return strconv.ParseBool(v)
}

// helper para obtener y validar una matriz de cadenas de la consulta
func GetQueryStringArray(c *gin.Context, key string, defaultValue []string) ([]string, error) {
	v, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, nil
	}
	arr := strings.Split(v, ",")
	for _, item := range arr {
		if _, err := strconv.Atoi(item); err != nil {
			return nil, fmt.Errorf("invalid value for %s parameter: %v", key, err)
		}
	}
	return arr, nil
}

// helper para obtener y validar un entero de la consulta
func GetQueryInt(c *gin.Context, key string, defaultValue int, validator func(int) bool) (int, error) {
	v, exists := c.GetQuery(key)
	if !exists {
		return defaultValue, nil
	}
	intValue, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %v", key, err)
	}
	if validator != nil && !validator(intValue) {
		return 0, fmt.Errorf("invalid value for %s parameter: %d", key, intValue)
	}
	return intValue, nil
}

func GetResultsFromHattrick(pathHattrick string, v any) error {
	config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET"))
	httpClient := config.Client(oauth1.NoContext, oauth1.NewToken(os.Getenv("OAUTH1_TOKEN"), os.Getenv("OAUTH1_TOKEN_SECRET")))
	path := fmt.Sprintf("%s%s", os.Getenv("BASE_RESOURCE_URL"), pathHattrick)
	resp, err := httpClient.Get(path)
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
