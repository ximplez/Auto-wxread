package github

import (
	"errors"
	"fmt"

	"github.com/ximplez/wxread/utils/http"
	"github.com/ximplez/wxread/utils/json_tool"
)

func getGithubRepoPubKey(githubToken, repo string) (string, string, error) {
	if githubToken == "" || repo == "" {
		return "", "", nil
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/actions/secrets/public-key", repo)
	if _, bytes, err := http.Get(url, nil, buildGithubHeader(githubToken)); err != nil {
		return "", "", err
	} else {
		resp := json_tool.PhaseJson[map[string]string](bytes)
		if resp == nil {
			return "", "", errors.New(fmt.Sprintf("获取公钥失败(%s) resp nil", url))
		}
		r := *resp
		if r["key_id"] == "" || r["key"] == "" {
			return "", "", errors.New(fmt.Sprintf("获取公钥失败(%s): %s", url, json_tool.ToJson(r, false)))
		}
		return r["key_id"], r["key"], nil
	}
}

func buildGithubHeader(githubToken string) map[string]string {
	return map[string]string{
		"Accept":               "application/vnd.github+json",
		"Authorization":        fmt.Sprintf("Bearer %s", githubToken),
		"X-GitHub-Api-Version": "2022-11-28",
	}
}
