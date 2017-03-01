package semaphoreci

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	cc "github.com/commonchat/commonchat-go"
	"github.com/grokify/gotilla/strings/stringsutil"
	"github.com/grokify/webhook-proxy-go/src/adapters"
	"github.com/grokify/webhook-proxy-go/src/config"
	"github.com/grokify/webhook-proxy-go/src/util"
	"github.com/valyala/fasthttp"
)

const (
	DisplayName = "Semaphore"
	HandlerKey  = "semaphoreci"
	IconURLX    = "https://d2rbro28ib85bu.cloudfront.net/images/integrations/128/semaphore.png"
	IconURL     = "https://a.slack-edge.com/ae7f/plugins/semaphore/assets/service_512.png"
	ICON_URL_2  = "https://s3.amazonaws.com/semaphore-media/logos/png/gear/semaphore-gear-large.png"
)

// FastHttp request handler for Semaphore CI outbound webhook
type SemaphoreciOutToGlipHandler struct {
	Config  config.Configuration
	Adapter adapters.Adapter
}

// FastHttp request handler constructor for Semaphore CI outbound webhook
func NewSemaphoreciOutToGlipHandler(cfg config.Configuration, adapter adapters.Adapter) SemaphoreciOutToGlipHandler {
	return SemaphoreciOutToGlipHandler{Config: cfg, Adapter: adapter}
}

// HandleFastHTTP is the method to respond to a fasthttp request.
func (h *SemaphoreciOutToGlipHandler) HandleFastHTTP(ctx *fasthttp.RequestCtx) {
	ccMsg, err := Normalize(ctx.PostBody())

	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusNotAcceptable)
		log.WithFields(log.Fields{
			"type":   "http.response",
			"status": fasthttp.StatusNotAcceptable,
		}).Info(fmt.Sprintf("%v request is not acceptable.", DisplayName))
		return
	}

	util.SendWebhook(ctx, h.Adapter, ccMsg)
}

//func NormalizeBytes(bytes []byte) (glipwebhook.GlipWebhookMessage, error) {
func Normalize(bytes []byte) (cc.Message, error) {
	message := cc.NewMessage()
	message.IconURL = IconURL

	baseMsg, err := SemaphoreciBaseOutMessageFromBytes(bytes)
	if err != nil {
		return message, err
	}

	switch baseMsg.Event {
	case "build":
		srcMsg, err := SemaphoreciBuildOutMessageFromBytes(bytes)
		if err != nil {
			return message, err
		}
		return NormalizeSemaphoreciBuildOutMessage(srcMsg), nil
	case "deploy":
		srcMsg, err := SemaphoreciDeployOutMessageFromBytes(bytes)
		if err != nil {
			return message, err
		}
		return NormalizeSemaphoreciDeployOutMessage(srcMsg), nil
	}
	return cc.Message{IconURL: IconURL}, errors.New("EventNotFound")
}

func NormalizeSemaphoreciBuildOutMessage(src SemaphoreciBuildOutMessage) cc.Message {
	message := cc.NewMessage()
	message.IconURL = IconURL

	message.Activity = fmt.Sprintf("%v %v %v", src.ProjectName, src.Event, src.Result)

	/*
		if strings.ToLower(strings.TrimSpace(src.Event)) == "build" {
			// Joe Cool's build #15 passed
			//message.Activity = fmt.Sprintf("%v %v #%v %v%v", src.ProjectName, src.Event, src.BuildNumber, src.Result, adapters.IntegrationActivitySuffix(DisplayName))
			message.Activity = fmt.Sprintf("%v %v", stringsutil.ToUpperFirst(src.Event), src.Result)
		} else {
			message.Activity = fmt.Sprintf("%v %v %v%v", src.ProjectName, src.Event, src.Result, adapters.IntegrationActivitySuffix(DisplayName))
		}
	*/
	message.Title = fmt.Sprintf("[%v #%v](%v) for **%v/%v** %v ([%v](%v))",
		stringsutil.ToUpperFirst(src.Event),
		src.BuildNumber,
		src.BuildURL,
		src.ProjectName,
		src.BranchName,
		src.Result,
		src.Commit.Id[:7],
		src.Commit.URL)

	attachment := cc.NewAttachment()

	if len(src.Commit.Message) > 0 {
		attachment.AddField(cc.Field{
			Title: "Message",
			Value: src.Commit.Message,
			Short: true})
	}
	if 1 == 0 {
		if len(src.ProjectName) > 0 {
			attachment.AddField(cc.Field{
				Title: "Project",
				Value: src.ProjectName,
				Short: true})
		}
		if len(src.BranchName) > 0 {
			attachment.AddField(cc.Field{
				Title: "Branch",
				Value: src.BranchName,
				Short: true})
		}
		if len(src.Event) > 0 {
			attachment.AddField(cc.Field{
				Title: "Event",
				Value: src.Event,
				Short: true})
		}
	}
	if len(src.Commit.AuthorName) > 0 {
		attachment.AddField(cc.Field{
			Title: "Committer",
			Value: src.Commit.AuthorName,
			Short: true})
	}

	message.AddAttachment(attachment)
	return message
}

func NormalizeSemaphoreciDeployOutMessage(src SemaphoreciDeployOutMessage) cc.Message {
	message := cc.NewMessage()
	message.IconURL = IconURL

	message.Activity = fmt.Sprintf("%v %v %v", src.ProjectName, src.Event, src.Result)

	message.Title = fmt.Sprintf("[%v #%v](%v) for **%v/%v** %v ([%v](%v))",
		stringsutil.ToUpperFirst(src.Event),
		src.Number, src.HtmlURL,
		src.ProjectName,
		src.BranchName,
		src.Result,
		src.Commit.Id[:7],
		src.Commit.URL)

	/*
				if strings.ToLower(strings.TrimSpace(src.Event)) == "build" {
					message.Activity = fmt.Sprintf("%v's %v #%v %v%v",
						src.Commit.AuthorName, src.Event, src.BuildNumber, src.Result, adapters.IntegrationActivitySuffix(DisplayName))
				} else {
					message.Activity = fmt.Sprintf("%v's %v %v%v",
						src.Commit.AuthorName, src.Event, src.Result, adapters.IntegrationActivitySuffix(DisplayName))
				}

				{
		    "project_name":"heroku-deploy-test",
		    "project_hash_id":"123-aga-471-6a8",
		    "result":"passed",
		    "event":"deploy",
		    "server_name":"server-heroku-master-automatic-2",
		    "number":2,
		    "created_at":"2013-07-30T13:52:33Z",
		    "updated_at":"2013-07-30T13:53:21Z",
		    "started_at":"2013-07-30T13:52:38Z",
		    "finished_at":"2013-07-30T13:53:21Z",
		    "html_url":"https://semaphoreci.com/projects/2420/servers/81/deploys/2",
		    "build_number":10,
		    "branch_name":"master",
		    "branch_html_url":"https://semaphoreci.com/projects/2420/branches/58394",
		    "build_html_url":"https://semaphoreci.com/projects/2420/branches/58394/builds/7",
		    "commit":{
		        "author_email":"rastasheep3@gmail.com",
		        "author_name":"Aleksandar Diklic",
		        "id":"43ddb7516ecc743f0563abd7418f0bd3617348c4",
		        "message":"One more time",
		        "timestamp":"2013-07-19T12:56:25Z",
		        "url":"https://github.com/rastasheep/heroku-deploy-test/commit/43ddb7516ecc743f0563abd7418f0bd3617348c4"
		    }
		}
	*/

	attachment := cc.NewAttachment()

	if len(src.Commit.Message) > 0 {
		attachment.AddField(cc.Field{
			Title: "Message",
			Value: src.Commit.Message})
	}
	if 1 == 0 {
		if len(src.ProjectName) > 0 {
			attachment.AddField(cc.Field{
				Title: "Project",
				Value: src.ProjectName,
				Short: true})
		}
		if len(src.BranchName) > 0 {
			attachment.AddField(cc.Field{
				Title: "Branch",
				Value: src.BranchName,
				Short: true})
		}
		if len(src.Event) > 0 {
			attachment.AddField(cc.Field{
				Title: "Event",
				Value: src.Event,
				Short: true})
		}
	}
	if len(src.Commit.AuthorName) > 0 {
		attachment.AddField(cc.Field{
			Title: "Committer",
			Value: src.Commit.AuthorName,
			Short: true})
	}

	message.AddAttachment(attachment)
	return message
}

type SemaphoreciBaseOutMessage struct {
	Event string `json:"event,omitempty"`
}

func SemaphoreciBaseOutMessageFromBytes(bytes []byte) (SemaphoreciBaseOutMessage, error) {
	msg := SemaphoreciBaseOutMessage{}
	err := json.Unmarshal(bytes, &msg)
	return msg, err
}

type SemaphoreciBuildOutMessage struct {
	BranchName    string            `json:"branch_name,omitempty"`
	BranchURL     string            `json:"branch_url,omitempty"`
	ProjectName   string            `json:"project_name,omitempty"`
	ProjectHashId string            `json:"project_hash_id,omitempty"`
	BuildURL      string            `json:"build_url,omitempty"`
	BuildNumber   int64             `json:"build_number,omitempty"`
	Result        string            `json:"result,omitempty"`
	Event         string            `json:"event,omitempty"`
	StartedAt     string            `json:"started_at,omitempty"`
	FinishedAt    string            `json:"finished_at,omitempty"`
	Commit        SemaphoreciCommit `json:"commit,omitempty"`
}

type SemaphoreciCommit struct {
	Id          string `json:"id,omitempty"`
	URL         string `json:"url,omitempty"`
	AuthorName  string `json:"author_name,omitempty"`
	AuthorEmail string `json:"author_email,omitempty"`
	Message     string `json:"message,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

func SemaphoreciBuildOutMessageFromBytes(bytes []byte) (SemaphoreciBuildOutMessage, error) {
	msg := SemaphoreciBuildOutMessage{}
	err := json.Unmarshal(bytes, &msg)
	if err == nil {
		msg.Commit.Message = strings.ToLower(strings.TrimSpace(msg.Commit.Message))
	}
	return msg, err
}

type SemaphoreciDeployOutMessage struct {
	ProjectName   string            `json:"project_name,omitempty"`
	ProjectHashId string            `json:"project_hash_id,omitempty"`
	Result        string            `json:"result,omitempty"`
	Event         string            `json:"event,omitempty"`
	ServerName    string            `json:"server_name,omitempty"`
	Number        int64             `json:"number,omitempty"`
	CreatedAt     string            `json:"created_at,omitempty"`
	UpdatedAt     string            `json:"updated_at,omitempty"`
	StartedAt     string            `json:"started_at,omitempty"`
	FinishedAt    string            `json:"finished_at,omitempty"`
	HtmlURL       string            `json:"html_url,omitempty"`
	BuildNumber   int64             `json:"build_number,omitempty"`
	BranchName    string            `json:"branch_name,omitempty"`
	BranchHtmlURL string            `json:"branch_html_url,omitempty"`
	BuildHtmlURL  string            `json:"bulid_html_url,omitempty"`
	Commit        SemaphoreciCommit `json:"commit,omitempty"`
}

func SemaphoreciDeployOutMessageFromBytes(bytes []byte) (SemaphoreciDeployOutMessage, error) {
	msg := SemaphoreciDeployOutMessage{}
	err := json.Unmarshal(bytes, &msg)
	if err == nil {
		msg.Commit.Message = strings.ToLower(strings.TrimSpace(msg.Commit.Message))
	}
	return msg, err
}
