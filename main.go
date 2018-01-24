package main

import (
	"encoding/json"
	"fmt"
	"github.com/boyter/golang-license-checker/parsers"
	vectorspace "github.com/boyter/golangvectorspace"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var dirPath = "/home/bboyter/Projects/hyperfine/"
var pathBlacklist = ".git,.hg,.svn"
var extentionBlacklist = "woff,eot,cur,dm,xpm,emz,db,scc,idx,mpp,dot,pspimage,stl,dml,wmf,rvm,resources,tlb,docx,doc,xls,xlsx,ppt,pptx,msg,vsd,chm,fm,book,dgn,blines,cab,lib,obj,jar,pdb,dll,bin,out,elf,so,msi,nupkg,pyc,ttf,woff2,jpg,jpeg,png,gif,bmp,psd,tif,tiff,yuv,ico,xls,xlsx,pdb,pdf,apk,com,exe,bz2,7z,tgz,rar,gz,zip,zipx,tar,rpm,bin,dmg,iso,vcd,mp3,flac,wma,wav,mid,m4a,3gp,flv,mov,mp4,mpg,rm,wmv,avi,m4v,sqlite,class,rlib,ncb,suo,opt,o,os,pch,pbm,pnm,ppm,pyd,pyo,raw,uyv,uyvy,xlsm,swf"

func readFile(filepath string) string {
	// TODO only read as deep into the file as we need
	b, err := ioutil.ReadFile(filepath)

	if err != nil {
		fmt.Print(err)
	}
	content := string(b)

	return content
}

func loadDatabase(filepath string) []parsers.License {
	jsonFile, err := os.Open(filepath)

	if err != nil {
		fmt.Println(err)
		return []parsers.License{}
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var database []parsers.License
	err = json.Unmarshal(byteValue, &database)

	if err != nil {
		fmt.Println(err)
		return []parsers.License{}
	}

	for i, v := range database {
		database[i].Concordance = vectorspace.BuildConcordance(strings.ToLower(v.Text))
	}

	return database
}

func walkDirectory(directory string, rootLicenses []parsers.LicenseMatch) {
	all, _ := ioutil.ReadDir(directory)

	directories := []string{}
	files := []string{}

	// Work out which directories and files we want to investigate
	for _, f := range all {
		if f.IsDir() {
			add := true

			for _, black := range strings.Split(pathBlacklist, ",") {
				if f.Name() == black {
					add = false
				}
			}

			if add == true {
				directories = append(directories, f.Name())
			}
		} else {
			files = append(files, f.Name())
		}
	}

	// Determine any possible licence files which would classify everything else
	possibleLicenses := parsers.FindPossibleLicenseFiles(files)
	for _, possibleLicense := range possibleLicenses {

		guessLicenses := parsers.GuessLicense(readFile(filepath.Join(directory, possibleLicense)), true, loadDatabase("database_keywords.json"))

		if len(guessLicenses) != 0 {
			rootLicenses = append(rootLicenses, guessLicenses[0])
		}
	}

	for _, file := range files {
		process := true

		for _, possibleLicenses := range possibleLicenses {
			if file == possibleLicenses {
				process = false
			}
		}

		for _, ext := range strings.Split(extentionBlacklist, ",") {
			if strings.HasSuffix(file, ext) {
				// Needs to be smarter we should skip reading the contents but it should still be under the license in the root folders
				process = false
			}
		}

		if process == true {
			licenseGuesses := parsers.GuessLicense(readFile(filepath.Join(directory, file)), true, loadDatabase("database_keywords.json"))

			licenseString := ""
			for _, v := range licenseGuesses {
				licenseString += fmt.Sprintf(" %s (%.1f%%)", v.Shortname, (v.Percentage * 100))
			}

			fmt.Println(directory, file, licenseString, rootLicenses)
		}
	}

	for _, newdirectory := range directories {
		walkDirectory(filepath.Join(directory, newdirectory), rootLicenses)
	}
}

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "golang-license-checker"
	app.Version = "1.0"
	app.Usage = "Check directory for licenses and list what license(s) a file is under"
	app.Action = func(c *cli.Context) error {
		return nil
	}

	app.Run(os.Args)

	// Everything after here needs to be refactored out to a subpackage
	walkDirectory(dirPath, []parsers.LicenseMatch{})
}
