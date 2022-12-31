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
	"fmt"
	"image/color"
	"runtime"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/p2p"
)

type Colors struct {
	gray       color.RGBA
	darkmatter color.RGBA
	green      color.RGBA
	red        color.RGBA
	yellow     color.RGBA
	white      color.RGBA
}

func layoutLoad() fyne.CanvasObject {
	loadResources()

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(MIN_WIDTH, MIN_HEIGHT))

	c := container.NewMax(
		rect,
		res.load,
	)

	layout := container.NewMax(
		res.background,
		c,
	)

	return layout
}

func layoutMain() fyne.CanvasObject {
	loadResources()

	m.Threads = runtime.GOMAXPROCS(0) / 2

	status.fastsync = true
	globals.Arguments["--fastsync"] = true
	status.network = false
	globals.Arguments["--testnet"] = false
	globals.Initialize()

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(580, 300))

	rect1 := canvas.NewRectangle(color.Transparent)
	rect1.SetMinSize(fyne.NewSize(1, 1))

	btnRect := canvas.NewRectangle(color.Transparent)
	btnRect.SetMinSize(fyne.NewSize(110, 50))

	btnRect2 := canvas.NewRectangle(color.Transparent)
	btnRect2.SetMinSize(fyne.NewSize(110, 30))

	rect50 := canvas.NewRectangle(color.Transparent)
	rect50.SetMinSize(fyne.NewSize(0, 50))

	rect280 := canvas.NewRectangle(color.Transparent)
	rect280.SetMinSize(fyne.NewSize(300, 300))

	div := canvas.NewRectangle(colors.red)
	div.SetMinSize(fyne.NewSize(500, 2))

	div2 := canvas.NewRectangle(colors.gray)
	div2.SetMinSize(fyne.NewSize(500, 1))

	div3 := canvas.NewRectangle(colors.gray)
	div3.SetMinSize(fyne.NewSize(500, 1))

	frame := canvas.NewRectangle(color.Transparent)
	frame.SetMinSize(fyne.NewSize(MIN_WIDTH-10, 100))

	rectCell := canvas.NewRectangle(color.Transparent)
	rectCell.SetMinSize(fyne.NewSize(150, 5))

	rectSpacer := canvas.NewRectangle(color.Transparent)
	rectSpacer.SetMinSize(fyne.NewSize(10, 5))

	title := canvas.NewText("Netrunner", colors.red)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 28

	configTitle := canvas.NewText("Configure", colors.red)
	configTitle.TextStyle = fyne.TextStyle{Bold: true}
	configTitle.TextSize = 25

	daemonTitle := canvas.NewText("Offline", colors.gray)
	daemonTitle.TextStyle = fyne.TextStyle{Bold: true}
	daemonTitle.TextSize = 25

	minerTitle := canvas.NewText("Offline", colors.gray)
	minerTitle.TextStyle = fyne.TextStyle{Bold: true}
	minerTitle.TextSize = 25

	btnConfig := widget.NewButton("CFG", nil)

	btnExplorer := widget.NewButton("EXP", nil)
	btnExplorer.OnTapped = func() {
		a.explorer.SetContent(layoutExplorer())
		a.explorer.Show()
	}
	btnExplorer.Disable()

	btnGnomon := widget.NewButton("GNO", nil)
	btnGnomon.OnTapped = func() {

	}

	btnReturn := widget.NewButton("RTN", nil)

	btnRewind := widget.NewButton("RWD", nil)
	btnRewind.OnTapped = func() {
		globals.Logger.Info("Attempting to rewind the blockchain by 50 blocks at height ", "height", bw.chain.Get_Height())
		if bw.chain != nil {
			if bw.chain.Get_Height() <= 50 {
				globals.Logger.Info("Could not rewind chain - height <= 50")
				return
			} else {
				if m.Mission == 1 {
					closeMiner()
				}

				bw.chain.Rewind_Chain(50)
				globals.Logger.Info("Blockchain rewind complete, new height is now ", "height", bw.chain.Get_Height())
			}
		}
	}
	btnRewind.Disable()

	btnStartDaemon := widget.NewButton("RUN", nil)

	btnStartMiner := widget.NewButton("RUN", nil)
	btnStartMiner.OnTapped = func() {
		if m.Mission == 0 {
			go startRunner(bw.chain.IntegratorAddress().String(), fmt.Sprintf("127.0.0.1:%d", globals.Config.GETWORK_Default_Port), m.Threads)
			minerTitle.Text = "Running"
			minerTitle.Color = colors.red
			minerTitle.Refresh()
			res.miner.Resource = resourceMinerOnPng
			res.miner.Refresh()
		} else {
			closeMiner()
		}
	}
	btnStartMiner.Disable()

	endpoints := canvas.NewText("ENDPOINT", colors.gray)
	endpoints.TextSize = 10
	endpoints.TextStyle = fyne.TextStyle{Bold: true}

	daemonIP := canvas.NewText("---", colors.gray)
	daemonIP.TextSize = 20

	versionLabel := canvas.NewText("VERSION", colors.gray)
	versionLabel.TextSize = 11

	version := canvas.NewText("---", colors.gray)
	version.TextSize = 20

	uptimeLabel := canvas.NewText("UPTIME", colors.gray)
	uptimeLabel.TextSize = 11
	uptime := canvas.NewText("---", colors.gray)
	uptime.TextSize = 20

	minerLabel := canvas.NewText("WORK", colors.gray)
	minerLabel.TextSize = 11
	minerIP := canvas.NewText(GetIP().String()+":10100", colors.gray)
	minerIP.TextSize = 11

	netLabel := canvas.NewText("NETWORK", colors.gray)
	netLabel.TextSize = 10
	netLabel.TextStyle = fyne.TextStyle{Bold: true}

	network := canvas.NewText("---", colors.gray)
	network.TextSize = 20

	heightLabel := canvas.NewText("HEIGHT", colors.gray)
	heightLabel.TextSize = 10
	heightLabel.TextStyle = fyne.TextStyle{Bold: true}

	height := canvas.NewText("---", colors.gray)
	height.TextSize = 20

	timeLabel := canvas.NewText("BLOCK  TIME", colors.gray)
	timeLabel.TextSize = 10
	timeLabel.TextStyle = fyne.TextStyle{Bold: true}

	btime := canvas.NewText("---", colors.gray)
	btime.TextSize = 20

	diffLabel := canvas.NewText("DIFFICULTY", colors.gray)
	diffLabel.TextSize = 10
	diffLabel.TextStyle = fyne.TextStyle{Bold: true}

	diff := canvas.NewText("---", colors.gray)
	diff.TextSize = 20

	supplyLabel := canvas.NewText("SUPPLY", colors.gray)
	supplyLabel.TextSize = 10
	supplyLabel.TextStyle = fyne.TextStyle{Bold: true}

	supply := canvas.NewText("---", colors.gray)
	supply.TextSize = 20

	tpoolLabel := canvas.NewText("TRANSACTION  POOL", colors.gray)
	tpoolLabel.TextSize = 10
	tpoolLabel.TextStyle = fyne.TextStyle{Bold: true}

	tpool := canvas.NewText("---", colors.gray)
	tpool.TextSize = 20

	rpoolLabel := canvas.NewText("REGISTRATION  POOL", colors.gray)
	rpoolLabel.TextSize = 10
	rpoolLabel.TextStyle = fyne.TextStyle{Bold: true}

	rpool := canvas.NewText("---", colors.gray)
	rpool.TextSize = 20

	peersLabel := canvas.NewText("PEERS", colors.gray)
	peersLabel.TextSize = 10
	peersLabel.TextStyle = fyne.TextStyle{Bold: true}

	peers := canvas.NewText("---", colors.gray)
	peers.TextSize = 20

	noffsetLabel := canvas.NewText("NTP  OFFSET", colors.gray)
	noffsetLabel.TextSize = 10
	noffsetLabel.TextStyle = fyne.TextStyle{Bold: true}

	noffset := canvas.NewText("---", colors.gray)
	noffset.TextSize = 16

	poffsetLabel := canvas.NewText("P2P  OFFSET", colors.gray)
	poffsetLabel.TextSize = 10
	poffsetLabel.TextStyle = fyne.TextStyle{Bold: true}

	poffset := canvas.NewText("---", colors.gray)
	poffset.TextSize = 16

	minersLabel := canvas.NewText("MINERS", colors.gray)
	minersLabel.TextSize = 10
	minersLabel.TextStyle = fyne.TextStyle{Bold: true}

	miners := canvas.NewText("---", colors.gray)
	miners.TextSize = 20

	threadsLabel := canvas.NewText("THREADS", colors.gray)
	threadsLabel.TextSize = 10
	threadsLabel.TextStyle = fyne.TextStyle{Bold: true}

	threads := canvas.NewText("---", colors.gray)
	threads.TextSize = 16

	blocksLabel := canvas.NewText("BLOCKS", colors.gray)
	blocksLabel.TextSize = 10
	blocksLabel.TextStyle = fyne.TextStyle{Bold: true}

	blocks := canvas.NewText("---", colors.gray)
	blocks.TextSize = 16

	hashrateLabel := canvas.NewText("HASHRATE", colors.gray)
	hashrateLabel.TextSize = 10
	hashrateLabel.TextStyle = fyne.TextStyle{Bold: true}

	hashrate := canvas.NewText("---", colors.gray)
	hashrate.TextSize = 16

	// Config Panel
	rectLeft := canvas.NewRectangle(color.Transparent)
	rectLeft.SetMinSize(fyne.NewSize(200, 200))

	rectMid := canvas.NewRectangle(color.Transparent)
	rectMid.SetMinSize(fyne.NewSize(500, 200))

	rectSlider := canvas.NewRectangle(colors.darkmatter)
	rectSlider.SetMinSize(fyne.NewSize(500, 30))

	radNetworkLabel := canvas.NewText("NETWORK", colors.red)
	radNetworkLabel.TextSize = 10
	radNetworkLabel.TextStyle = fyne.TextStyle{Bold: true}

	radNetwork := widget.NewRadioGroup([]string{"Mainnet", "Testnet"}, nil)

	if _, ok := globals.Arguments["--testnet"]; ok && globals.Arguments["--testnet"] != nil {
		if globals.Arguments["--testnet"].(bool) {
			radNetwork.SetSelected("Testnet")
			status.network = true
		} else {
			radNetwork.SetSelected("Mainnet")
			status.network = false
		}
	} else {
		radNetwork.SetSelected("Mainnet")
		status.network = false
	}

	radSyncLabel := canvas.NewText("SYNC  MODE", colors.red)
	radSyncLabel.TextSize = 10
	radSyncLabel.TextStyle = fyne.TextStyle{Bold: true}

	radSync := widget.NewRadioGroup([]string{"Full", "Fast"}, nil)
	radSync.OnChanged = func(s string) {
		if s == "Fast" {
			status.fastsync = true
			globals.Arguments["--fastsync"] = true
			radSync.SetSelected("Fast")
		} else {
			status.fastsync = false
			radSync.SetSelected("Full")
			globals.Arguments["--fastsync"] = false
		}
	}

	if _, ok := globals.Arguments["--fastsync"]; ok && globals.Arguments["--fastsync"] != nil {
		if globals.Arguments["--fastsync"].(bool) {
			status.fastsync = true
			radSync.SetSelected("Fast")
		} else {
			status.fastsync = false
			radSync.SetSelected("Full")
		}
	} else {
		status.fastsync = true
		radSync.SetSelected("Fast")
	}

	rewardLabel := canvas.NewText("MINING  ADDRESS", colors.red)
	rewardLabel.TextSize = 10
	rewardLabel.TextStyle = fyne.TextStyle{Bold: true}

	reward := widget.NewEntry()
	reward.Validator = func(s string) error {
		addr, err := globals.ParseValidateAddress(s)
		if err != nil {
			reward.SetValidationError(err)
			reward.SetPlaceHolder(m.Address)
			return err
		} else {
			status.integrator = s

			if bw.chain != nil {
				bw.chain.SetIntegratorAddress(addr.Clone())
				m.Address = addr.String()
				reward.SetPlaceHolder(addr.String())
			}
			reward.SetValidationError(nil)
			return nil
		}
	}
	reward.SetPlaceHolder("Enter a DERO Address")

	radNetwork.OnChanged = func(s string) {
		if s == "Testnet" {
			globals.Arguments["--testnet"] = true
			status.network = true
		} else {
			globals.Arguments["--testnet"] = false
			status.network = false
		}
		globals.Initialize()
		reward.Validate()
	}

	configThreadsLabel := canvas.NewText("MINING  THREADS", colors.red)
	configThreadsLabel.TextSize = 10
	configThreadsLabel.TextStyle = fyne.TextStyle{Bold: true}

	configThreadsCount := canvas.NewText(strconv.Itoa(m.Threads), colors.red)
	configThreadsCount.TextSize = 16

	configThreads := widget.NewSlider(1, float64(runtime.GOMAXPROCS(0))-2)
	configThreads.OnChanged = func(f float64) {
		if m.Mission == 0 {
			m.Threads = int(f)
		}
		configThreadsCount.Text = strconv.Itoa(m.Threads)
		configThreadsCount.Refresh()
		configThreads.Value = float64(m.Threads)
		configThreads.Refresh()
	}

	configThreadsCount.Text = strconv.Itoa(m.Threads)
	configThreadsCount.Refresh()
	configThreads.Value = float64(m.Threads)
	configThreads.Refresh()

	// Resources
	res.background.SetMinSize(fyne.NewSize(1100, 600))
	res.background.Refresh()
	res.daemon.SetMinSize(fyne.NewSize(50, 50))
	res.daemon.Refresh()
	res.miner.SetMinSize(fyne.NewSize(50, 50))
	res.miner.Refresh()
	res.icon.SetMinSize(fyne.NewSize(45, 45))
	res.icon.Refresh()

	btnStartDaemon.OnTapped = func() {
		if status.active == 0 {
			status.uptime = "---"
			status.version = "---"
			if status.integrator != "" {
				globals.Arguments["--integrator-address"] = status.integrator
			}

			if status.network {
				globals.Arguments["--rpc-bind"] = "0.0.0.0:" + strconv.Itoa(DEFAULT_DAEMON_TESTNET_RPC_PORT)
			} else {
				globals.Arguments["--rpc-bind"] = "0.0.0.0:" + strconv.Itoa(DEFAULT_DAEMON_MAINNET_RPC_PORT)
			}

			startDaemon()
			daemonTitle.Text = "Offline"
			daemonTitle.Color = colors.gray
			daemonTitle.Refresh()
			btnStartDaemon.Disable()

			go func() {
				for bw.chain != nil {
					if m.Mission == 1 {
						hashrateLabel.Color = colors.red
						hashrateLabel.Refresh()
						hashrate.Text = m.Hashrate
						hashrate.Color = colors.red
						hashrate.Refresh()
						threadsLabel.Color = colors.red
						threadsLabel.Refresh()
						threads.Text = strconv.Itoa(m.Threads)
						threads.Color = colors.red
						threads.Refresh()
						blocksLabel.Color = colors.red
						blocksLabel.Refresh()
						blocks.Text = strconv.FormatInt(status.blocks_accepted, 10)
						blocks.Color = colors.red
						blocks.Refresh()
						minerTitle.Text = "Running"
						minerTitle.Color = colors.red
						minerTitle.Refresh()
						res.miner.Resource = resourceMinerOnPng
						res.miner.Refresh()
						btnStartMiner.Text = "        END        "
						btnStartMiner.Refresh()
					} else {
						hashrateLabel.Color = colors.gray
						hashrateLabel.Refresh()
						hashrate.Text = "---"
						hashrate.Color = colors.gray
						hashrate.Refresh()
						threadsLabel.Color = colors.gray
						threadsLabel.Refresh()
						threads.Text = "---"
						threads.Color = colors.gray
						threads.Refresh()
						blocksLabel.Color = colors.gray
						blocksLabel.Refresh()
						if status.blocks_accepted <= 0 {
							blocks.Text = "---"
						} else {
							blocks.Text = strconv.FormatInt(status.blocks_accepted, 10)
						}
						blocks.Color = colors.gray
						blocks.Refresh()
						minerTitle.Text = "Offline"
						minerTitle.Color = colors.gray
						minerTitle.Refresh()
						res.miner.Resource = resourceMinerOffPng
						res.miner.Refresh()
						m.Hashrate = ""
						btnStartMiner.Text = "        RUN        "
						btnStartMiner.Refresh()
					}

					if status.network {
						status.network = true
						radNetwork.SetSelected("Testnet")
						network.Text = "Testnet"
						network.Refresh()
						status.ip_daemon = GetIP().String() + ":" + strconv.Itoa(DEFAULT_DAEMON_TESTNET_RPC_PORT)
						daemonIP.Text = status.ip_daemon
						daemonIP.Refresh()
					} else {
						status.network = false
						radNetwork.SetSelected("Mainnet")
						network.Text = "Mainnet"
						network.Refresh()
						status.ip_daemon = GetIP().String() + ":" + strconv.Itoa(DEFAULT_DAEMON_MAINNET_RPC_PORT)
						daemonIP.Text = status.ip_daemon
						daemonIP.Refresh()
					}
					radNetwork.Disable()

					if status.fastsync {
						radSync.SetSelected("Fast")
					} else {
						radSync.SetSelected("Full")
					}
					radSync.Disable()

					bestHeight, _ := p2p.Best_Peer_Height()
					if int64(bw.chain.Get_Height()) != bestHeight {
						best, _ := p2p.Best_Peer_Height()
						progress := ""
						percent := float64(bw.chain.Get_Height()) / float64(best) * 100
						if percent <= 0 {
							progress = ""
						} else if percent >= 100 {
							progress = ""
						} else {
							progress = fmt.Sprintf("%.2f", percent) + "%"
						}

						if bw.chain.Get_Height() == -1 && status.fastsync {
							daemonTitle.Text = "Finalizing Bootstrap... "
							daemonTitle.Color = colors.white
						} else if bw.chain.Get_Height() > 20000 && status.fastsync {
							daemonTitle.Text = "Bootstrap Error - Full Syncing... " + progress
							daemonTitle.Color = colors.white
						} else {
							daemonTitle.Text = "Syncing... " + progress
							daemonTitle.Color = colors.white
						}
						daemonTitle.Refresh()
						if m.Mission == 0 {
							btnStartMiner.Disable()
						}
					} else {
						daemonTitle.Text = "Running"
						daemonTitle.Color = colors.red
						daemonTitle.Refresh()
						btnStartMiner.Enable()

						if !a.explorer.Content().Visible() {
							btnExplorer.Enable()
						} else {
							btnExplorer.Disable()
						}
					}

					ver := strings.Split(status.version, ".DEROHE")
					version.Text = ver[0]
					uptime.Text = status.uptime
					height.Text = fmt.Sprintf("%d / %d", bw.chain.Get_Height(), bestHeight)
					btime.Text = fmt.Sprintf("%.2f", status.block_time)
					diff.Text = fmt.Sprintf("%d", status.difficulty)
					supply.Text = globals.FormatMoney(status.supply)
					tpool.Text = fmt.Sprintf("%d", status.tx_pool)
					rpool.Text = fmt.Sprintf("%d", status.reg_pool)
					peers.Text = fmt.Sprintf("%d", status.peers)
					miners.Text = fmt.Sprintf("%d", status.miners)
					noffset.Text = status.offset_ntp
					poffset.Text = status.offset_p2p

					if bw.chain.Get_Height() > 50 {
						btnRewind.Enable()
					}

					//reward.Text = fmt.Sprintf("%s", bw.chain.IntegratorAddress())
					//reward.Disable()
					btnStartDaemon.Disable()
					btnConfig.Enable()
					res.daemon.Resource = resourceDaemonOnPng
					res.daemon.Refresh()
					netLabel.Color = colors.red
					netLabel.Refresh()
					network.Color = colors.red
					network.Refresh()
					endpoints.Color = colors.red
					endpoints.Refresh()
					daemonIP.Color = colors.red
					daemonIP.Refresh()
					versionLabel.Color = colors.red
					versionLabel.Refresh()
					version.Color = colors.red
					version.Refresh()
					uptimeLabel.Color = colors.red
					uptimeLabel.Refresh()
					uptime.Color = colors.red
					uptime.Refresh()
					heightLabel.Color = colors.red
					heightLabel.Refresh()
					height.Color = colors.red
					height.Refresh()
					timeLabel.Color = colors.red
					timeLabel.Refresh()
					btime.Color = colors.red
					btime.Refresh()
					supplyLabel.Color = colors.red
					supplyLabel.Refresh()
					supply.Color = colors.red
					supply.Refresh()
					peersLabel.Color = colors.red
					peersLabel.Refresh()
					peers.Color = colors.red
					peers.Refresh()
					minersLabel.Color = colors.red
					minersLabel.Refresh()
					miners.Color = colors.red
					miners.Refresh()
					diffLabel.Color = colors.red
					diffLabel.Refresh()
					diff.Color = colors.red
					diff.Refresh()
					tpoolLabel.Color = colors.red
					tpoolLabel.Refresh()
					tpool.Color = colors.red
					tpool.Refresh()
					rpoolLabel.Color = colors.red
					rpoolLabel.Refresh()
					rpool.Color = colors.red
					rpool.Refresh()
					noffsetLabel.Color = colors.red
					noffsetLabel.Refresh()
					noffset.Color = colors.red
					noffset.Refresh()
					poffsetLabel.Color = colors.red
					poffsetLabel.Refresh()
					poffset.Color = colors.red
					poffset.Refresh()

					time.Sleep(1 * time.Second)
				}
			}()
		}
	}

	statusPanel := container.NewVBox(
		rect1,
		rectSpacer,
		rectSpacer,
		container.NewHBox(
			rectSpacer,
			rectSpacer,
			rectSpacer,
			rectSpacer,
			container.NewMax(
				rect280,
				container.NewVBox(
					rectSpacer,
					netLabel,
					rectSpacer,
					network,
					rectSpacer,
					rectSpacer,
					endpoints,
					rectSpacer,
					daemonIP,
					rectSpacer,
					rectSpacer,
					versionLabel,
					rectSpacer,
					version,
					rectSpacer,
					rectSpacer,
					uptimeLabel,
					rectSpacer,
					uptime,
				),
			),
			container.NewMax(
				rect280,
				container.NewVBox(
					rectSpacer,
					heightLabel,
					rectSpacer,
					height,
					rectSpacer,
					rectSpacer,
					timeLabel,
					rectSpacer,
					btime,
					rectSpacer,
					rectSpacer,
					supplyLabel,
					rectSpacer,
					supply,
					rectSpacer,
					rectSpacer,
					peersLabel,
					rectSpacer,
					peers,
				),
			),
			container.NewMax(
				rect280,
				container.NewVBox(
					rectSpacer,
					minersLabel,
					rectSpacer,
					miners,
					rectSpacer,
					rectSpacer,
					diffLabel,
					rectSpacer,
					diff,
					rectSpacer,
					rectSpacer,
					tpoolLabel,
					rectSpacer,
					tpool,
					rectSpacer,
					rectSpacer,
					rpoolLabel,
					rectSpacer,
					rpool,
				),
			),
			layout.NewSpacer(),
			container.NewMax(
				container.NewVBox(
					container.NewMax(
						btnRect2,
						btnRewind,
					),
					container.NewMax(
						btnRect2,
						btnExplorer,
					),
					container.NewMax(
						btnRect2,
						//btnGnomon,
					),
				),
			),
			rect1,
		),
	)

	configPanel := container.NewMax(
		container.NewVBox(
			div3,
			rectSpacer,
			container.NewHBox(
				rectSpacer,
				rect50,
				configTitle,
				layout.NewSpacer(),
				container.NewMax(
					btnRect2,
					btnReturn,
				),
				rect1,
			),
			rectSpacer,
			container.NewHBox(
				rectSpacer,
				rectSpacer,
				rectSpacer,
				rectSpacer,
				container.NewMax(
					rectLeft,
					container.NewVBox(
						rectSpacer,
						radNetworkLabel,
						rectSpacer,
						radNetwork,
						rectSpacer,
						rectSpacer,
						rectSpacer,
						radSyncLabel,
						rectSpacer,
						radSync,
						rectSpacer,
					),
				),
				rectSpacer,
				rectSpacer,
				container.NewMax(
					rectMid,
					container.NewVBox(
						rectSpacer,
						rewardLabel,
						rectSpacer,
						reward,
						rectSpacer,
						rectSpacer,
						rectSpacer,
						rectSpacer,
						rectSpacer,
						rectSpacer,
						rect1,
						container.NewHBox(
							configThreadsLabel,
							rectSpacer,
							rectSpacer,
							configThreadsCount,
						),
						rectSpacer,
						rectSpacer,
						container.NewMax(
							rectSlider,
							configThreads,
						),
						rectSpacer,
					),
				),
			),
		),
	)

	bodyBox := container.NewMax(
		statusPanel,
	)

	btnConfig.OnTapped = func() {
		bodyBox.RemoveAll()
		bodyBox.AddObject(configPanel)
		bodyBox.Refresh()
	}

	btnReturn.OnTapped = func() {
		bodyBox.RemoveAll()
		bodyBox.AddObject(statusPanel)
		bodyBox.Refresh()
	}

	daemonBox := container.NewHBox(
		rectSpacer,
		res.daemon,
		rectSpacer,
		daemonTitle,
		rectSpacer,
		layout.NewSpacer(),
		rectSpacer,
		rectSpacer,
		rectSpacer,
		rectCell,
		container.NewMax(
			rectCell,
			container.NewHBox(
				noffsetLabel,
				rectSpacer,
				rectSpacer,
				noffset,
				layout.NewSpacer(),
			),
		),
		rectSpacer,
		rectSpacer,
		container.NewMax(
			rectCell,
			container.NewHBox(
				poffsetLabel,
				rectSpacer,
				rectSpacer,
				poffset,
				layout.NewSpacer(),
			),
		),
		rectSpacer,
		rectSpacer,
		rectSpacer,
		container.NewMax(
			btnRect,
			btnStartDaemon,
		),
		rectSpacer,
	)

	minerBox := container.NewHBox(
		rectSpacer,
		res.miner,
		rectSpacer,
		minerTitle,
		layout.NewSpacer(),
		rectSpacer,
		rectSpacer,
		rectSpacer,
		rectCell,
		container.NewMax(
			rectCell,
			container.NewHBox(
				hashrateLabel,
				rectSpacer,
				rectSpacer,
				hashrate,
				layout.NewSpacer(),
			),
		),
		layout.NewSpacer(),
		rectSpacer,
		rectSpacer,
		container.NewMax(
			rectCell,
			container.NewHBox(
				threadsLabel,
				rectSpacer,
				rectSpacer,
				threads,
				layout.NewSpacer(),
			),
		),
		rectSpacer,
		rectSpacer,
		container.NewMax(
			rectCell,
			container.NewHBox(
				blocksLabel,
				rectSpacer,
				rectSpacer,
				blocks,
				layout.NewSpacer(),
			),
		),
		rectSpacer,
		rectSpacer,
		rectSpacer,
		container.NewMax(
			btnRect,
			btnStartMiner,
		),
		rectSpacer,
	)

	topBox := container.NewHBox(
		rectSpacer,
		rect50,
		title,
		rectSpacer,
		layout.NewSpacer(),
		container.NewMax(
			btnRect,
			btnConfig,
		),
		rectSpacer,
	)

	form := container.NewVBox(
		div2,
		rectSpacer,
		minerBox,
		rectSpacer,
	)

	top := container.NewVBox(
		rectSpacer,
		topBox,
		rectSpacer,
		div,
		rectSpacer,
		rectSpacer,
		daemonBox,
		rectSpacer,
	)

	left := container.NewMax(
		frame,
		bodyBox,
	)

	c := container.NewBorder(
		top,
		form,
		left,
		nil,
	)

	layout := container.NewMax(
		res.background,
		c,
	)

	return layout
}

func layoutExplorer() fyne.CanvasObject {
	loadResources()

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(680, 300))

	rect50 := canvas.NewRectangle(color.Transparent)
	rect50.SetMinSize(fyne.NewSize(0, 50))

	btnRect := canvas.NewRectangle(color.Transparent)
	btnRect.SetMinSize(fyne.NewSize(110, 50))

	rect280 := canvas.NewRectangle(color.Transparent)
	rect280.SetMinSize(fyne.NewSize(300, 300))

	div := canvas.NewRectangle(colors.red)
	div.SetMinSize(fyne.NewSize(1000, 2))

	div2 := canvas.NewRectangle(colors.gray)
	div2.SetMinSize(fyne.NewSize(1000, 1))

	frame := canvas.NewRectangle(color.Transparent)
	frame.SetMinSize(fyne.NewSize(1000, 300))

	rectCell := canvas.NewRectangle(color.Transparent)
	rectCell.SetMinSize(fyne.NewSize(150, 5))

	rectSpacer := canvas.NewRectangle(color.Transparent)
	rectSpacer.SetMinSize(fyne.NewSize(10, 5))

	title := canvas.NewText("Explorer", colors.red)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.TextSize = 28

	btnReturn := widget.NewButton("END", nil)
	btnReturn.OnTapped = func() {
		a.explorer.Content().Hide()
		a.explorer.Hide()
	}

	topBox := container.NewHBox(
		rectSpacer,
		rect50,
		title,
		rectSpacer,
		layout.NewSpacer(),
		container.NewMax(
			btnRect,
			btnReturn,
		),
		rectSpacer,
	)

	top := container.NewVBox(
		rectSpacer,
		topBox,
		rectSpacer,
		div,
		rectSpacer,
		rectSpacer,
		//
		rectSpacer,
	)

	c := container.NewBorder(
		top,
		nil,
		nil,
		nil,
	)

	layout := container.NewMax(
		res.background,
		c,
	)

	return layout
}
