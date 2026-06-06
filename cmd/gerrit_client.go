package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	HostEnv       = "GERRIT_HOST"
	UsernameEnv   = "GERRIT_USER"
	SkipVerifyTLS = "GERRIT_SKIP_VERIFY_TLS"
)

type GerritClient struct {
	host       string
	username   string
	password   string
	httpClient *http.Client
}

func (c *GerritClient) decodeResponse(resp *http.Response, v any) error {
	reader := bufio.NewReader(resp.Body)
	reader.ReadString('\n')
	return json.NewDecoder(reader).Decode(v)
}

func NewGerritClient() *GerritClient {
	host := os.Getenv(HostEnv)
	if host == "" {
		fmt.Printf("Environment variable %s cannot be empty\n", HostEnv)
		os.Exit(1)
	}

	username := os.Getenv(UsernameEnv)
	if username == "" {
		fmt.Printf("Environment variable %s cannot be empty\n", UsernameEnv)
		os.Exit(1)
	}

	password, _ := GetPassword(host, username) // Ignore error.  It's okay if password is empty

	return &GerritClient{
		host,
		username,
		password,
		&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: os.Getenv(SkipVerifyTLS) == "1",
				},
			},
		},
	}
}

type AccountInfo struct {
	AccountID int    `json:"_account_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
}

type ChangeInfo struct {
	ID                     string                              `json:"id"`
	TripletID              string                              `json:"triplet_id"`
	Project                string                              `json:"project"`
	Branch                 string                              `json:"branch"`
	FullBranch             string                              `json:"full_branch"`
	Topic                  string                              `json:"topic"`
	AttentionSet           map[string]json.RawMessage          `json:"attention_set"`
	RemovedAttentionSet    map[string]json.RawMessage          `json:"removed_from_attention_set"`
	Hashtags               []string                            `json:"hashtags"`
	CustomKeyedValues      map[string]string                   `json:"custom_keyed_values"`
	ChangeID               string                              `json:"change_id"`
	Subject                string                              `json:"subject"`
	Status                 string                              `json:"status"`
	Created                string                              `json:"created"`
	Updated                string                              `json:"updated"`
	Submitted              string                              `json:"submitted"`
	Submitter              *AccountInfo                        `json:"submitter"`
	Starred                bool                                `json:"starred"`
	Reviewed               bool                                `json:"reviewed"`
	SubmitType             string                              `json:"submit_type"`
	Mergeable              *bool                               `json:"mergeable"`
	Submittable            *bool                               `json:"submittable"`
	Insertions             *int                                `json:"insertions"`
	Deletions              *int                                `json:"deletions"`
	TotalCommentCount      int                                 `json:"total_comment_count"`
	UnresolvedCommentCount int                                 `json:"unresolved_comment_count"`
	Number                 int                                 `json:"_number"`
	VirtualIDNumber        int                                 `json:"virtual_id_number"`
	Owner                  AccountInfo                         `json:"owner"`
	Actions                map[string]json.RawMessage          `json:"actions"`
	SubmitRecords          []json.RawMessage                   `json:"submit_records"`
	Requirements           []json.RawMessage                   `json:"requirements"`
	SubmitRequirements     []json.RawMessage                   `json:"submit_requirements"`
	Labels                 map[string]json.RawMessage          `json:"labels"`
	PermittedLabels        map[string][]string                 `json:"permitted_labels"`
	RemovableLabels        map[string]map[string][]AccountInfo `json:"removable_labels"`
	RemovableReviewers     []AccountInfo                       `json:"removable_reviewers"`
	Reviewers              map[string][]AccountInfo            `json:"reviewers"`
	PendingReviewers       map[string][]AccountInfo            `json:"pending_reviewers"`
	ReviewerUpdates        []json.RawMessage                   `json:"reviewer_updates"`
	Messages               []json.RawMessage                   `json:"messages"`
	CurrentRevisionNumber  int                                 `json:"current_revision_number"`
	CurrentRevision        string                              `json:"current_revision"`
	Revisions              map[string]json.RawMessage          `json:"revisions"`
	MetaRevID              string                              `json:"meta_rev_id"`
	TrackingIDs            []json.RawMessage                   `json:"tracking_ids"`
	MoreChanges            bool                                `json:"_more_changes"`
	Problems               []json.RawMessage                   `json:"problems"`
	IsPrivate              bool                                `json:"is_private"`
	WorkInProgress         bool                                `json:"work_in_progress"`
	HasReviewStarted       bool                                `json:"has_review_started"`
	RevertOf               int                                 `json:"revert_of"`
	SubmissionID           string                              `json:"submission_id"`
	CherryPickOfChange     int                                 `json:"cherry_pick_of_change"`
	CherryPickOfPatchSet   int                                 `json:"cherry_pick_of_patch_set"`
	ContainsGitConflicts   bool                                `json:"contains_git_conflicts"`
}

func (c *GerritClient) GetAccountInfo() (*AccountInfo, error) {
	url := "https://" + c.host + "/a/accounts/self"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(responseBody))
	}

	var accountInfo AccountInfo

	err = c.decodeResponse(resp, &accountInfo)
	if err != nil {
		return nil, err
	}

	return &accountInfo, nil
}

func (c *GerritClient) GetChanges() ([]ChangeInfo, error) {
	url := "https://" + c.host + "/a/changes/?o=DETAILED_ACCOUNTS"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(responseBody))
	}

	var changeInfo []ChangeInfo

	err = c.decodeResponse(resp, &changeInfo)
	if err != nil {
		return nil, err
	}

	return changeInfo, nil
}
