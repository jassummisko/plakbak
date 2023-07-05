package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type FileEntry struct {
	filePath string
	contents []byte
}

func cutStartingPath(startingPath string, path string) string {
	if !strings.HasSuffix(startingPath, "/") {
		startingPath = startingPath + "/"
	}
	path, _ = strings.CutPrefix(path, startingPath)
	return path
}

func walkFiles(startingPath string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(startingPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if len(path) > 0 {
				files = append(files, path)
			}
			return nil
		},
	)
	return files, err
}

func formatFileEntries(entries []FileEntry) string {
	buffer := ""
	for _, entry := range entries {
		subBuffer := ""
		subBuffer += "{\"" + entry.filePath + "\", "
		subBuffer += "{"
		for _, b := range entry.contents {
			subBuffer += fmt.Sprint(b) + ","
		}
		buffer += subBuffer + "}},\n"
		fmt.Println("Built " + entry.filePath)
	}
	return buffer
}

func getUserApiKey(devkey string, username string, password string) (string, error) {
	client := &http.Client{}
	var data = strings.NewReader(`api_dev_key=` + devkey + `&api_user_name=` + username + `&api_user_password=` + password)
	req, err := http.NewRequest("POST", "https://pastebin.com/api/api_login.php", data)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyText), nil
}

func submitPaste(devkey string, str string, userkey string) (string, error) {
	client := &http.Client{}
	var data = strings.NewReader(`api_dev_key=` + devkey + `&api_paste_code=` + str + `&api_option=paste&api_user_key=` + userkey + `&api_paste_expire_date=10M`)
	req, err := http.NewRequest("POST", "https://pastebin.com/api/api_post.php", data)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyText), nil
}

func dumpToFile(data string) error {
	file, err := os.Create("dump.file")
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			panic(err)
		}
	}()

	file.WriteString(data)
	return nil
}

func readSessionKey() (string, error) {
	contents, err := os.ReadFile(".usersessionkey")
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func requestAndSaveSessionKey(devkey string, username string, password string) (string, error) {
	key, err := getUserApiKey(devkey, username, password)
	if err != nil {
		return "", err
	}

	file, err := os.Create(".usersessionkey")
	if err != nil {
		return "", err
	}

	file.WriteString(key)
	return key, nil
}

func main() {
	var walkFolderFlag = flag.String("folder", "", "Folder to package")
	flag.StringVar(walkFolderFlag, "f", "", "Folder to package")
	flag.Parse()

	config, err := ReadConfig()
	if err != nil {
		panic(err)
	}

	walkFolder := *walkFolderFlag

	if walkFolder == "" {
		walkFolder = config.SourceFolder
	}

	stat, err := os.Stat(walkFolder)
	if err != nil {
		if err := os.Mkdir(walkFolder, os.ModePerm); err != nil {
			panic(err)
		}
	} else {
		if !stat.IsDir() {
			fmt.Println("SourceFolder marked is not a directory")
			os.Exit(1)
		}
	}

	if config.DevApiKey == "YOUR DEV API KEY HERE" {
		fmt.Println("Please add your dev API key in config.plakbak.\nVisit https://pastebin.com/doc_api for more information.")
		os.Exit(1)
	}

	devkey := config.DevApiKey
	username := config.Username
	password := config.Password

	sessionkey, err := readSessionKey()
	if err != nil {
		sessionkey, err = requestAndSaveSessionKey(devkey, username, password)
		fmt.Println("Pastebin user session key saved to .usersessionkey")
		if err != nil {
			panic(err)
		}
	}

	files, err := walkFiles(walkFolder)
	if err != nil {
		panic(err)
	}

	entries := []FileEntry{}

	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			panic(err)
		}

		if fi.IsDir() {
			continue
		}

		contents, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}

		entries = append(entries, FileEntry{cutStartingPath(walkFolder, file), contents})
	}

	str := formatFileEntries(entries)
	finalSubmission := fmt.Sprintf(pasteTemplate, str)

	fmt.Println("Getting pastebin code ...")
	thing, err := submitPaste(devkey, finalSubmission, sessionkey)
	if err != nil {
		panic(err)
	}

	code, found := strings.CutPrefix(thing, "https://pastebin.com/")
	fmt.Println(code)

	if found {
		fmt.Println("Run 'pastebin run " + code + "' in your desired folder on your CC computer.")
	} else {
		fmt.Println("Dumping to dump.file\nCopy this into your computer through your clipboard.")
		dumpToFile(finalSubmission)
	}
}
