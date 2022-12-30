// Netrunner
// Copyright 2021-2023 DERO Foundation. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package main

import (
	"image/color"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/blang/semver/v4"
	"github.com/deroproject/derohe/config"
	"github.com/deroproject/derohe/globals"
	"github.com/docopt/docopt-go"
)

type App struct {
	app      fyne.App
	window   fyne.Window
	focus    bool
	explorer fyne.Window
}

type Status struct {
	// Daemon status
	active          int64
	integrator      string
	ip_daemon       string
	network         bool
	fastsync        bool
	block_time      float32
	difficulty      uint64
	height          int64
	last_height     int64
	stable_height   int64
	topo_height     int64
	peers           uint64
	peer_height     int64
	miners          int
	estimate_1hr    float64
	estimate_1d     float64
	estimate_7d     float64
	blocks_accepted int64
	blocks_rejected int64
	offset_p2p      string
	offset_ntp      string
	total_blocks    int64
	supply          uint64
	tx_pool         int
	reg_pool        int
	uptime          string
	version         string
}

const (
	MIN_WIDTH  = 1100
	MIN_HEIGHT = 600
)

var a App
var bw Blackwall
var colors Colors
var status Status
var version semver.Version

func main() {
	var err error

	// Turn off profiler
	runtime.MemProfileRate = 0

	// Initialize terminal arguments if any
	globals.Arguments, err = docopt.ParseArgs(command_line, nil, config.Version.String())
	if err != nil {
		globals.Logger.Error(err, "Error while parsing options err: %s\n")
	}

	globals.Initialize()

	version = semver.MustParse("0.1.0")
	a.app = app.New()
	t := &nTheme{}
	a.app.Settings().SetTheme(t)
	a.app.SetIcon(resourceIconPng)
	a.window = a.app.NewWindow("Netrunner")
	a.window.SetIcon(resourceIconPng)
	a.window.SetMaster()
	a.window.SetCloseIntercept(appClose)
	a.window.SetPadded(false)
	a.window.CenterOnScreen()
	a.window.Resize(fyne.NewSize(MIN_WIDTH, MIN_HEIGHT))
	a.window.SetFixedSize(true)

	a.explorer = a.app.NewWindow("Explorer")
	a.explorer.SetIcon(resourceIconPng)
	a.explorer.SetPadded(false)
	a.explorer.Resize(fyne.NewSize(MIN_WIDTH, MIN_HEIGHT))
	a.explorer.SetFixedSize(true)
	a.explorer.CenterOnScreen()
	a.explorer.SetContent(layoutExplorer())
	a.explorer.Hide()

	colors.gray = color.RGBA{55, 55, 55, 255}
	colors.darkmatter = color.RGBA{25, 25, 25, 55}
	colors.green = color.RGBA{5, 182, 5, 255}
	colors.red = color.RGBA{185, 71, 68, 255}
	colors.yellow = color.RGBA{201, 205, 85, 255}
	colors.white = color.RGBA{195, 195, 195, 255}

	a.window.SetContent(layoutLoad())

	go func() {
		time.Sleep(5 * time.Second)
		a.window.SetContent(layoutMain())
	}()

	a.window.ShowAndRun()
}
