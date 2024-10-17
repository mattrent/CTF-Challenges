package handlers

import (
	"bytes"
	"deployer/config"
	"deployer/internal/storage"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	ctfd "github.com/ctfer-io/go-ctfd/api"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Add challenge
	ch, err = postChallenges(client, &PostContainerChallengesParams{
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
		Identifier:     challenge.Id,
	})
	if err != nil {
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
					Name:    "handout",
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

// from library
func postChallenges(client *ctfd.Client, params *PostContainerChallengesParams) (*ctfd.Challenge, error) {
	chall := &ctfd.Challenge{}
	if err := post(client, "/challenges", params, &chall); err != nil {
		return nil, err
	}
	return chall, nil
}

func post(client *ctfd.Client, edp string, params any, dst any) error {
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}
	req, _ := http.NewRequest(http.MethodPost, edp, bytes.NewBuffer(body))

	return call(client, req, dst)
}

func call(client *ctfd.Client, req *http.Request, dst any) error {
	// Set API base URL
	newUrl, err := url.Parse("/api/v1" + req.URL.String())
	if err != nil {
		return err
	}
	req.URL = newUrl

	// Issue HTTP request
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Decode response
	resp := ctfd.Response{
		Data: dst,
	}
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return errors.Wrapf(err, "CTFd responded with invalid JSON for content")
	}

	// Handle errors if any
	if resp.Errors != nil {
		return fmt.Errorf("CTFd responded with errors: %v", resp.Errors)
	}
	if !resp.Success {
		// This case should not happen, as status code already serves this goal
		// and errors gives the reasons.
		if resp.Message != nil {
			return fmt.Errorf("CTFd responded with no success but no error, got message: %s", *resp.Message)
		}
		return errors.New("CTFd responded with no success but no error, and no message")
	}
	return nil
}

type PostContainerChallengesParams struct {
	Name           string             `json:"name"`
	Category       string             `json:"category"`
	Description    string             `json:"description"`
	ConnectionInfo *string            `json:"connection_info,omitempty"`
	Value          int                `json:"value"`
	Function       string             `json:"function"`
	Initial        *int               `json:"initial,omitempty"`
	Decay          *int               `json:"decay,omitempty"`
	Minimum        *int               `json:"minimum,omitempty"`
	MaxAttempts    *int               `json:"max_attempts,omitempty"`
	NextID         *int               `json:"next_id,omitempty"`
	Requirements   *ctfd.Requirements `json:"requirements,omitempty"`
	State          string             `json:"state"`
	Type           string             `json:"type"`
	Identifier     string             `json:"identifier"`
}

type ChallengeReponse struct {
	ID         int    `json:"id"`
	Identifier string `json:"identifier"`
}
