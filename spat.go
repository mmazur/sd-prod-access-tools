package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type GHRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func getLatestGitHubRelease(orgrepo string) (string, []string, error) {
	// TODO: figure out how to handle this handles pre-releases/test releases and correct the code

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", orgrepo)

	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response into a Release struct
	var release GHRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", nil, err
	}

	// Extract the version tag and list of assets from the Release struct
	version := release.TagName
	assets := []string{}
	for _, asset := range release.Assets {
		assets = append(assets, asset.URL)
	}

	return version, assets, nil
}

type GLRelease struct {
	TagName string `json:"tag_name"`
	Assets  struct {
		Count int `json:"count"`
		Links []struct {
			DirectAssetURL string `json:"direct_asset_url"`
		} `json:"links"`
	} `json:"assets"`
}

func getLatestGitLabRelease(url string) (string, []string, error) {
	// TODO: figure out how to handle this handles pre-releases/test releases and correct the code

	// Send a GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	// Decode the JSON response into a slice of Release structs
	var releases []GLRelease
	err = json.NewDecoder(resp.Body).Decode(&releases)
	if err != nil {
		return "", nil, err
	}

	// Extract the version tag and list of assets from the latest Release struct
	latestRelease := releases[0]
	version := latestRelease.TagName
	assets := make([]string, len(latestRelease.Assets.Links))
	for i, asset := range latestRelease.Assets.Links {
		assets[i] = asset.DirectAssetURL
	}

	return version, assets, nil
}

func cmdCheck() {
	fmt.Println("Latest versions:")

	url := "https://gitlab.cee.redhat.com/api/v4/projects/33674/releases"
	version, assets, err := getLatestGitLabRelease(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("v%s, \t%d assets,\tservice/backplane-cli (gitlab)\n", version, len(assets))

	ghtools := []string{"openshift-online/ocm-cli", "openshift/osdctl", "openshift/rosa", "coreos/butane", "prometheus/prometheus"}
	for _, orgrepo := range ghtools {
		version, assets, err := getLatestGitHubRelease(orgrepo)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("%s, \t%d assets,\t%s (github)\n", version, len(assets), orgrepo)
		/*		fmt.Println("Assets:")
				for _, asset := range assets {
					fmt.Println("-", asset)
				}*/
	}
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "spat",
		Short: "SD Prod Access Tools Manager",
		Long:  "SD Prod Access Tools Manager",
	}

	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Check and report upstream versions of all managed tools",
		Run: func(cmd *cobra.Command, args []string) {
			cmdCheck()
		},
	}

	rootCmd.AddCommand(checkCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
