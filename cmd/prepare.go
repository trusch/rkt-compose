// Copyright © 2017 Tino Rusch
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
)

// prepareCmd represents the prepare command
var prepareCmd = &cobra.Command{
	Use:   "prepare",
	Short: "prepare prepares a pod for run",
	Long:  `prepare prepares a pod for run`,
	Run: func(cmd *cobra.Command, args []string) {
		prepare()
	},
}

func init() {
	RootCmd.AddCommand(prepareCmd)
}

func prepare() {
	if checkIfPrepareNeeded() {
		log.Print("prepare pod-manifest...")
		composeFile := getComposeFile()
		if err := composeFile.Prepare(viper.GetString("manifest")); err != nil {
			log.Fatal("error preparing pod-manifest: ", err)
		}
	} else {
		log.Print("manifest already up to date")
	}
}

func checkIfPrepareNeeded() bool {
	composePath := viper.GetString("file")
	manifestPath := viper.GetString("manifest")
	composeStat, err1 := os.Stat(composePath)
	manifestStat, err2 := os.Stat(manifestPath)
	if err1 != nil || err2 != nil || composeStat.ModTime().After(manifestStat.ModTime()) {
		return true
	}
	return false
}
