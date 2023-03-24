package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

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

func cmdCheck(cmd *cobra.Command, args []string) {
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

func createDirectories() error {
	// TODO: this is probably too linux-specific
	//       and needs error reporting
	fmt.Print("Creating directory structure…\t\t")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	dirs := []string{
		filepath.Join(homeDir, ".spat"),
		filepath.Join(homeDir, ".spat", "bin"),
		filepath.Join(homeDir, ".spat", "tmp"),
	}

	for _, dir := range dirs {
		err := os.Mkdir(dir, 0700)
		if err != nil && !os.IsExist(err) {
			fmt.Println("FAILED")
			return err
		}
	}

	fmt.Println("DONE (in `$HOME/.spat`)")

	return nil
}

func findAsset(assets []string, rematch string) string {
	r := regexp.MustCompile(rematch)
	for _, asset := range assets {
		if r.MatchString(asset) {
			return asset
		}
	}
	return ""
}

func downloadAsset(url string, path string) string {
	tmpfile := filepath.Join(path, "tmpfile") // TODO: just no

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return ""
	}
	defer resp.Body.Close()

	out, err := os.Create(tmpfile)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return ""
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return ""
	}

	return tmpfile
}

func installTool(tooltype string, toolid string, assetmatch string, binaryname string) {
	if tooltype != "github" {
		fmt.Printf("Installing tools of type %s isn't supported", tooltype)
		return
	}

	version, assets, err := getLatestGitHubRelease(toolid)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	asset := findAsset(assets, assetmatch)
	if asset == "" {
		fmt.Println("Couldn't find asset to download")
		return
	}

	homeDir, _ := os.UserHomeDir()
	bindir := filepath.Join(homeDir, ".spat", "bin")
	binfile := filepath.Join(bindir, binaryname)
	tmpdir := filepath.Join(homeDir, ".spat", "tmp")

	// TODO: support unzipping first

	tmpfile := downloadAsset(asset, tmpdir)
	err = os.Rename(tmpfile, binfile)
	if err != nil {
		fmt.Println("Error moving file:", err)
		return
	}
	os.Chmod(binfile, os.FileMode(0755))

	fmt.Printf("Installed %s %s as binary '%s'\n", toolid, version, binaryname)

}

func cmdInit(cmd *cobra.Command, args []string) {
	_ = createDirectories()
	installTool("github", "openshift-online/ocm-cli", "/ocm-linux-amd64$", "ocm")
}

func cmdUpgrade(cmd *cobra.Command, args []string) {
	fmt.Println("Upgrading tools…\t\t\tFAIL (not implemented yet)")
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "spat",
		Short: "SD Prod Access Tools Manager",
		Long:  "SD Prod Access Tools Manager",
	}

	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize spat and install prod access tools",
		Run:   cmdInit,
	}

	var checkCmd = &cobra.Command{
		Use:   "check",
		Short: "Check local and upstream versions of all managed tools",
		Run:   cmdCheck,
	}

	var upgradeCmd = &cobra.Command{
		Use:   "upgrade",
		Short: "Upgraded (chosen) managed tools to their latest versions",
		Run:   cmdUpgrade,
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(upgradeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
