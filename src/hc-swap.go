package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/manifoldco/promptui"
	"golang.org/x/net/html"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func prompt(title string, items []string) string {
	prompt := promptui.Select{
		Label: title,
		Items: items,
	}
	_, result, err := prompt.Run()
	if err != nil {
		os.Exit(0)
	}
	return result
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()
	os.MkdirAll(dest, 0755)
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()
		path := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func toolDownload(tool string, toolbin string, versions []string) {
	hashi := "https://releases.hashicorp.com/"
	for _, vers := range versions {
		fmt.Printf("\nDownloading %v v%v...", tool, vers)
		tooldir := toolbin + "/" + vers + "/"
		err := os.MkdirAll(tooldir, 0755)
		check(err)
		os.Chdir(tooldir)
		toolURL := hashi + tool + "/" + vers + "/" + tool + "_" + vers + "_" + runtime.GOOS + "_" + runtime.GOARCH + ".zip"
		resp, err := http.Get(toolURL)
		check(err)
		defer resp.Body.Close()
		zipfile, err := os.Create(fmt.Sprintf("./%s.zip", vers))
		check(err)
		defer zipfile.Close()
		_, err = io.Copy(zipfile, resp.Body)
		check(err)
		fmt.Printf("\nDownload complete, unzipping binary...")
		err = unzip(fmt.Sprintf("./%v.zip", vers), tooldir)
		check(err)
	}
}

func getLatestVersions(tool string) []string {
	url := "https://releases.hashicorp.com/" + tool + "/"
	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)
	var latestVersions []string
	for len(latestVersions) < 10 {
		tt := z.Next()
		if tt == html.StartTagToken {
			t := z.Token()
			isAnchor := t.Data == "a"
			if isAnchor {
				for _, a := range t.Attr {
					if a.Key == "href" && a.Val != "../" && a.Val != "https://fastly.com/?utm_source=hashicorp" {
						s := strings.TrimPrefix(a.Val, fmt.Sprintf("/%v/", tool))
						s = strings.Trim(s, "/")
						latestVersions = append(latestVersions, s)
					}
				}
			}
		}
	}
	return latestVersions
}

func toolSymlink(tool string, toolbin string, homebin string, version string) {
	fmt.Printf("\nCreating Symlink for %v v%v...\n\n", tool, version)
	path := toolbin + version + "/"
	target := filepath.Join(path, tool)
	symlink := filepath.Join(homebin, tool)
	if _, err := os.Lstat(symlink); err == nil {
		if err := os.Remove(symlink); err != nil {
			panic(err)
		}
	} else if os.IsNotExist(err) {
		err := os.MkdirAll(homebin, 0755)
		check(err)
	}
	err := os.Symlink(target, symlink)
	check(err)
	cmd := exec.Command(fmt.Sprintf("%v", tool), "--version")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func toolUninstall(tool string, toolbin string, version string) {
	fmt.Printf("\nUninstalling %v v%v...", tool, version)
	path := toolbin + version
	err := os.RemoveAll(path)
	check(err)
	fmt.Printf("\nUninstall complete!\n\n")
	main()
}

func main() {
	usr, _ := user.Current()
	home := usr.HomeDir

	tool := prompt("Tool Select", []string{"Terraform", "Packer", "Vault"})
	tool = strings.ToLower(tool)
	toolbin := home + "/hc-swap/" + tool + "-versions/"
	homebin := "/usr/local/bin"

	files, err := ioutil.ReadDir(toolbin)
	if os.IsNotExist(err) {
		fmt.Printf("\nNo %v installations found...", tool)
		fmt.Printf("\nSetting up standard versioning directories at %v", toolbin)
		err := os.MkdirAll(toolbin, 0755)
		check(err)
		toolLatest := getLatestVersions(tool)
		version := prompt("Select version to install:", toolLatest)
		toolDownload(tool, toolbin, []string{version})
		toolSymlink(tool, toolbin, homebin, version)
	} else {
		versions := []string{}
		for _, f := range files {
			fname := f.Name()
			versions = append(versions, fname)
		}
		versions = append(versions, "Install New", "Uninstall", "Exit")
		version := prompt("\nSelect Version", versions)
		switch {
		case version == "Install New":
			latest := getLatestVersions(tool)
			version := prompt("\nSelect Version", latest)
			toolDownload(tool, toolbin, []string{version})
			toolSymlink(tool, toolbin, homebin, version)
		case version == "Uninstall":
			versions = remove(versions, "Install New")
			versions = remove(versions, "Uninstall")
			version := prompt("\nSelect Version", versions)
			toolUninstall(tool, toolbin, version)
		case version == "Exit":
			break
		default:
			toolSymlink(tool, toolbin, homebin, version)
		}
	}
}
