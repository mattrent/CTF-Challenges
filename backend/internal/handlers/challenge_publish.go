package handlers

import (
	"deployer/config"
	"deployer/internal/storage"
	"log"
	"net/http"
	"os"
	"path/filepath"

	ctfd "github.com/ctfer-io/go-ctfd/api"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type ChallengeCtfd struct {
	Name           string        `json:"name"`
	Author         string        `json:"author"`
	Category       string        `json:"category"`
	Description    string        `json:"description"`
	Value          int           `json:"value"`
	Type           string        `json:"type"`
	Extra          Extra         `json:"extra"`
	Image          interface{}   `json:"image"`
	Protocol       interface{}   `json:"protocol"`
	Host           interface{}   `json:"host"`
	ConnectionInfo string        `json:"connection_info"`
	Healthcheck    string        `json:"healthcheck"`
	Attempts       int           `json:"attempts"`
	Flags          []string      `json:"flags"`
	Topics         []string      `json:"topics"`
	Tags           []string      `json:"tags"`
	Files          []string      `json:"files"`
	Hints          []HintElement `json:"hints"`
	Requirements   []string      `json:"requirements"`
	State          string        `json:"state"`
	Version        string        `json:"version"`
}

type Extra struct {
	Initial  int    `json:"initial"`
	Decay    int    `json:"decay"`
	Minimum  int    `json:"minimum"`
	Function string `json:"function"`
}

type FlagClass struct {
	Type    string  `json:"type"`
	Content string  `json:"content"`
	Data    *string `json:"data,omitempty"`
}

type HintClass struct {
	Content string `json:"content"`
	Cost    int    `json:"cost"`
}

type HintElement struct {
	HintClass *HintClass
	String    *string
}

func PublishChallenge(c *gin.Context) {
	challengeId := c.Param("id")
	userId := c.GetString(userIdValue)
	challenge, err := storage.GetChallenge(challengeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if challenge.UserId != userId {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	// Publish to CTFd
	nonce, session, err := ctfd.GetNonceAndSession(config.Values.CTFDURL)
	if err != nil {
		log.Println("Could not connect to CTFd: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := ctfd.NewClient(config.Values.CTFDURL, nonce, session, "")
	client.SetAPIKey(config.Values.CTFDAPIToken)

	dst := filepath.Join(config.Values.UploadPath, challenge.Id, "challenge.yml")
	yamlFile, err := os.ReadFile(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	conf := &ChallengeCtfd{}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var ch *ctfd.Challenge

	if challenge.CtfdId.Valid {
		err = client.DeleteChallenge(int(challenge.CtfdId.Int64))
		if err != nil {
			log.Println("Could not delete from CTFd: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Add challenge
	ch, err = client.PostChallenges(&ctfd.PostChallengesParams{
		Name:           conf.Name,
		Category:       conf.Category,
		Description:    conf.Description,
		ConnectionInfo: &conf.ConnectionInfo,
		Value:          conf.Value,
		MaxAttempts:    &conf.Attempts,
		Function:       conf.Extra.Function,
		Initial:        &conf.Extra.Initial,
		Decay:          &conf.Extra.Decay,
		Minimum:        &conf.Extra.Minimum,
		State:          conf.State,
		Type:           conf.Type,
	})
	if err != nil {
		log.Println("Could not add challenge to CTFd: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Upload handout
	file := filepath.Join(config.Values.UploadPath, challenge.Id, "handout.zip")
	data, err := os.ReadFile(file)
	if err == nil {
		_, err = client.PostFiles(&ctfd.PostFilesParams{
			Files: []*ctfd.InputFile{
				{
					Name:    "handout.zip",
					Content: data,
				},
			},
			Challenge: &ch.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	_, err = client.PostFlags(&ctfd.PostFlagsParams{
		Challenge: ch.ID,
		Content:   conf.Flags[0],
		Type:      "static",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Change challenge status
	err = storage.PublishChallengeWithReference(challenge.Id, ch.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Println("Challenge status changed to published")

	c.IndentedJSON(http.StatusCreated, gin.H{})
}
