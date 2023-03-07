package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

type AzureBody struct {
	Status   int            `json:"status"`
	Comments []AzureComment `json:"comments"`
}

type AzureComment struct {
	ParentCommentID int    `json:"parentCommentId"`
	Content         string `json:"content"`
	CommentType     int    `json:"commentType"`
}

type AzureConf struct {
	URL, Auth string
}

func makeComment(coverage float64, badTests []string) string {
	coverageComment := "Total coverage is " + sf("%.2f", coverage) + "%"
	testComment := "All tests are successful. 💪\n\n"
	if len(badTests) != 0 {
		testComment = "Test failed. 🙅 \n\n Failed tests:\n\n"
		testComment += "|Test name|\n|--------|\n"
		for _, t := range badTests {
			testComment += "|" + t + "|\n"
		}
		testComment += "\n"
	}
	return testComment + coverageComment
}

func (az AzureConf) sendAzureComment(coverage float64, failedTests []string) error {
	if az.URL == "" {
		return nil
	}
	dat, err := json.Marshal(AzureBody{
		Status: 2,
		Comments: []AzureComment{{
			ParentCommentID: 0,
			CommentType:     1,
			Content:         makeComment(coverage, failedTests),
		}},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, az.URL, bytes.NewReader(dat))
	if err != nil {
		return err
	}

	fmt.Println("will do request")
	req.Header.Add("Authorization", "Bearer "+az.Auth)
	req.Header.Set("Content-Type", "application/json")

	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("Request: %s\n", dump)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error: %+v", err)
		return err
	}

	defer resp.Body.Close()

	dump, _ = httputil.DumpResponse(resp, true)
	fmt.Printf("Response: %s\n", dump)
	return nil
}
