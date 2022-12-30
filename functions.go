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
	"net"
	"os"
	"runtime"
	"time"

	"fyne.io/fyne/v2/canvas"
	derodpkg "github.com/civilware/derodpkg/cmd"
	"github.com/deroproject/derohe/blockchain"
	"github.com/deroproject/derohe/cmd/derod/rpc"
	"github.com/deroproject/derohe/config"
	"github.com/deroproject/derohe/cryptography/crypto"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/p2p"
)

type Blackwall struct {
	chain  *blockchain.Blockchain
	server *rpc.RPCServer
	status int64
}

type resources struct {
	background *canvas.Image
	daemon     *canvas.Image
	load       *canvas.Image
	miner      *canvas.Image
	icon       *canvas.Image
}

const (
	DEFAULT_DAEMON_LOCAL_ADDRESS     = "127.0.0.1"
	DEFAULT_DAEMON_WORK_ADDRESS      = "0.0.0.0"
	DEFAULT_DAEMON_MAINNET_RPC_PORT  = 10102
	DEFAULT_DAEMON_MAINNET_P2P_PORT  = 10101
	DEFAULT_DAEMON_MAINNET_WORK_PORT = 10100
	DEFAULT_DAEMON_TESTNET_RPC_PORT  = 40402
	DEFAULT_DAEMON_TESTNET_P2P_PORT  = 40401
	DEFAULT_DAEMON_TESTNET_WORK_PORT = 40400
)

var res resources
var command_line string = `derod 
DERO : A secure, private blockchain with smart-contracts
Usage:
  derod [--version] [--testnet] [--debug]  [--sync-node] [--timeisinsync] [--fastsync] [--socks-proxy=<socks_ip:port>] [--data-dir=<directory>] [--p2p-bind=<0.0.0.0:18089>] [--add-exclusive-node=<ip:port>]... [--add-priority-node=<ip:port>]... [--min-peers=<11>] [--max-peers=<100>] [--rpc-bind=<127.0.0.1:9999>] [--getwork-bind=<0.0.0.0:18089>] [--node-tag=<unique name>] [--prune-history=<50>] [--integrator-address=<address>] [--clog-level=1] [--flog-level=1]
  derod --version
Options:
  --version     Show version.
  --testnet  	Run in testnet mode.
  --debug       Debug mode enabled, print more log messages
  --clog-level=1	Set console log level (0 to 127) 
  --flog-level=1	Set file log level (0 to 127)
  --fastsync      Fast sync mode (this option has effect only while bootstrapping)
  --timeisinsync  Confirms to daemon that time is in sync, so daemon doesn't try to sync
  --socks-proxy=<socks_ip:port>  Use a proxy to connect to network.
  --data-dir=<directory>    Store blockchain data at this location
  --rpc-bind=<127.0.0.1:9999>    RPC listens on this ip:port
  --p2p-bind=<0.0.0.0:18089>    p2p server listens on this ip:port, specify port 0 to disable listening server
  --getwork-bind=<0.0.0.0:10100>    getwork server listens on this ip:port, specify port 0 to disable listening server
  --add-exclusive-node=<ip:port>	Connect to specific peer only 
  --add-priority-node=<ip:port>	Maintain persistant connection to specified peer
  --sync-node       Sync node automatically with the seeds nodes. This option is for rare use.
  --node-tag=<unique name>	Unique name of node, visible to everyone
  --integrator-address	if this node mines a block,Integrator rewards will be given to address.default is dev's address.
  --min-peers=<31>	  Node will try to maintain atleast this many connections to peers
  --max-peers=<101>	  Node will maintain maximim this many connections to peers and will stop accepting connections
  --prune-history=<50>	prunes blockchain history until the specific topo_height
  `

// Load the resources as images from bundled.go
func loadResources() {
	res.load = canvas.NewImageFromResource(resourceLoadPng)
	res.load.FillMode = canvas.ImageFillStretch
	res.background = canvas.NewImageFromResource(resourceNetrunnerPng)
	res.background.FillMode = canvas.ImageFillStretch
	res.daemon = canvas.NewImageFromResource(resourceDaemonOffPng)
	res.daemon.FillMode = canvas.ImageFillContain
	res.miner = canvas.NewImageFromResource(resourceMinerOffPng)
	res.miner.FillMode = canvas.ImageFillContain
	res.icon = canvas.NewImageFromResource(resourceIconPng)
	res.icon.FillMode = canvas.ImageFillContain
}

func startDaemon() {
	if bw.chain == nil {
		// Initialize DERO blockchain
		bw.chain = derodpkg.InitializeDerod(globals.Arguments)

		// Initialize RPC server
		bw.server = derodpkg.StartDerod(bw.chain)

		res.daemon.Resource = resourceDaemonOnPng
		res.daemon.Refresh()

		status.active = 1

		globals.Logger.Info("Netrunner ", "Version", version)
		globals.Logger.Info("Copyright 2020-2023 DERO Foundation. All rights reserved.")
		globals.Logger.Info("OS", runtime.GOOS, "ARCH", runtime.GOARCH, "GOMAXPROCS", runtime.GOMAXPROCS(0))

		// Run the update routine
		go update()
	}
}

// Always call this for a graceful close
func appClose() {
	closeMiner()
	closeDaemon()
	os.Exit(0)
}

func closeDaemon() {
	if bw.server != nil {
		bw.server.RPCServer_Stop()
		bw.server = nil
	}

	if bw.chain != nil {
		bw.chain.Shutdown()
		bw.chain = nil
	}
}

func closeMiner() {
	if m.Mission == 1 {
		m.Mission = 0
	}
}

func getStatus() {
	if bw.chain == nil {
		return
	}

	var zerohash crypto.Hash

	blid, err := bw.chain.Load_Block_Topological_order_at_index(bw.chain.Get_Height())
	blid50, err := bw.chain.Load_Block_Topological_order_at_index(bw.chain.Get_Height() - 50)
	if err == nil {
		if blid50 != zerohash {
			now := bw.chain.Load_Block_Timestamp(blid)
			now50 := bw.chain.Load_Block_Timestamp(blid50)
			status.block_time = float32(now-now50) / (50.0 * 1000)
		} else {
			status.block_time = 0
		}
	} else {
		status.block_time = 0
	}

	status.height = bw.chain.Get_Height()
	status.difficulty = bw.chain.Get_Difficulty()
	status.stable_height = bw.chain.Get_Stable_Height()
	status.peers = p2p.Peer_Count()
	status.peer_height, _ = p2p.Best_Peer_Height()
	status.miners = rpc.CountMiners()
	status.estimate_1d = rpc.HashrateEstimatePercent_1day()
	status.estimate_1hr = rpc.HashrateEstimatePercent_1hr()
	status.estimate_7d = rpc.HashrateEstimatePercent_7day()
	status.blocks_accepted = rpc.CountMinisAccepted
	status.blocks_rejected = rpc.CountMinisRejected
	status.total_blocks = rpc.CountBlocks
	status.supply = config.PREMINE + blockchain.CalcBlockReward(uint64(bw.chain.Get_Height()))*uint64(bw.chain.Get_Height())
	status.tx_pool = len(bw.chain.Mempool.Mempool_List_TX())
	status.reg_pool = len(bw.chain.Regpool.Regpool_List_TX())
	status.uptime = time.Now().Sub(globals.StartTime).Round(time.Second).String()
	status.version = config.Version.String()
	status.offset_p2p = globals.GetOffsetP2P().Round(time.Millisecond).String()
	status.offset_ntp = globals.GetOffsetNTP().Round(time.Millisecond).String()

	return
}

// Routine to update status per second
func update() {
	for bw.chain != nil {
		getStatus()

		status.last_height = bw.chain.Get_Height()

		time.Sleep(1 * time.Second)
	}
}

// Find the IP address for endpoint
func GetIP() net.IP {
	conn, err := net.Dial("udp", "1.1.1.1:80")
	if err != nil {
		fmt.Printf("[Netrunner]  Failed to dial UDP - Check firewall settings...\n")
		return nil
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)

	return addr.IP
}
