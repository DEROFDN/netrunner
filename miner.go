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
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/deroproject/derohe/astrobwt/astrobwtv3"
	"github.com/deroproject/derohe/block"
	"github.com/deroproject/derohe/globals"
	"github.com/deroproject/derohe/rpc"

	"github.com/gorilla/websocket"
)

type Miner struct {
	Mission     int64
	Height      int64
	Blocks      uint64
	MiniBlocks  uint64
	Hashrate    string
	NWHashrate  string
	Connection  *websocket.Conn
	Label       *canvas.Text
	LabelBlocks *canvas.Text
	Address     string
	Threads     int
	Daemon      string
	BlockList   []string
	ScrollBox   *widget.List
	Data        binding.StringList
}

var m Miner
var mutex sync.RWMutex
var job rpc.GetBlockTemplate_Result
var job_counter int64
var maxdelay int = 10000
var iterations int = 100
var max_pow_size int = 819200 //astrobwt.MAX_LENGTH
var counter uint64
var hash_rate uint64
var Difficulty uint64
var our_height int64
var block_counter uint64
var mini_block_counter uint64

func startRunner(w string, d string, t int) {
	m.Mission = 1

	globals.Arguments["--wallet-address"] = w
	globals.Arguments["--daemon-rpc-address"] = d
	globals.Arguments["--mining-threads"] = strconv.Itoa(t)

	if status.network {
		globals.Arguments["--testnet"] = true
	} else {
		globals.Arguments["--testnet"] = false
	}

	if globals.Arguments["--wallet-address"] != nil {
		addr, err := globals.ParseValidateAddress(globals.Arguments["--wallet-address"].(string))
		if err != nil {
			globals.Logger.Error(err, "[Miner] Wallet address is invalid.")
			return
		}

		m.Address = addr.String()
	}

	if globals.Arguments["--daemon-rpc-address"] != nil {
		m.Daemon = globals.Arguments["--daemon-rpc-address"].(string)
	}

	m.Threads = runtime.GOMAXPROCS(0)
	if globals.Arguments["--mining-threads"] != nil {
		if s, err := strconv.Atoi(globals.Arguments["--mining-threads"].(string)); err == nil {
			m.Threads = s
		} else {
			globals.Logger.Error(err, "[Miner] Mining threads argument cannot be parsed.")
		}

		if m.Threads > runtime.GOMAXPROCS(0) {
			globals.Logger.Info("[Miner] Mining threads is more than available CPUs. This is NOT optimal", "thread_count", m.Threads, "max_possible", runtime.GOMAXPROCS(0))
		}
	}

	globals.Logger.Info(fmt.Sprintf("[Miner] System will mine to \"%s\" with %d threads. Good Luck!!", m.Address, m.Threads))

	if m.Threads < 1 || iterations < 1 || m.Threads > 2048 {
		iterations = 1
		m.Threads = 1
	}

	// This tiny goroutine continuously updates status as required
	go func() {
		for m.Mission == 1 {
			last_our_height := int64(0)
			last_best_height := int64(0)

			last_counter := uint64(0)
			last_counter_time := time.Now()
			last_mining_state := false

			_ = last_mining_state

			mining := true
			mining_speed := 0.00
			for m.Mission == 1 {
				if m.Mission == 0 {
					mining_speed = 0
					return
				}

				best_height := int64(0)
				// only update prompt if needed
				if last_our_height != our_height || last_best_height != best_height || last_counter != counter {
					mining_string := ""

					if mining {
						mining_speed = float64(counter-last_counter) / (float64(uint64(time.Since(last_counter_time))) / 1000000000.0)
						last_counter = counter
						last_counter_time = time.Now()
						switch {
						case mining_speed > 1000000:
							mining_string = fmt.Sprintf("%.3f MH/s", float32(mining_speed)/1000000.0)
						case mining_speed > 1000:
							mining_string = fmt.Sprintf("%.3f KH/s", float32(mining_speed)/1000.0)
						case mining_speed > 0:
							mining_string = fmt.Sprintf("%.0f H/s", mining_speed)
						}
					}
					last_mining_state = mining

					hash_rate_string := ""

					switch {
					case hash_rate > 1000000000000:
						hash_rate_string = fmt.Sprintf("%.3f TH/s", float64(hash_rate)/1000000000000.0)
					case hash_rate > 1000000000:
						hash_rate_string = fmt.Sprintf("%.3f GH/s", float64(hash_rate)/1000000000.0)
					case hash_rate > 1000000:
						hash_rate_string = fmt.Sprintf("%.3f MH/s", float64(hash_rate)/1000000.0)
					case hash_rate > 1000:
						hash_rate_string = fmt.Sprintf("%.3f KH/s", float64(hash_rate)/1000.0)
					case hash_rate > 0:
						hash_rate_string = fmt.Sprintf("%d H/s", hash_rate)
					}

					m.Height = our_height
					m.Blocks = block_counter
					m.MiniBlocks = mini_block_counter
					m.NWHashrate = hash_rate_string
					m.Hashrate = mining_string

					last_our_height = our_height
					last_best_height = best_height
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	if m.Threads > 255 {
		globals.Logger.Error(nil, "[Miner] This program supports maximum 256 CPU cores.", "available", m.Threads)
		m.Threads = 255
	}

	if m.Mission == 1 {
		go getwork(m.Address)

		for i := 0; i < m.Threads; i++ {
			go mineblock(i)
		}
	}

	return
}

func random_execution(wg *sync.WaitGroup, iterations int) {
	var workbuf [255]byte

	runtime.LockOSThread()
	threadaffinity()

	rand.Read(workbuf[:])

	for i := 0; i < iterations; i++ {
		_ = astrobwtv3.AstroBWTv3(workbuf[:])
	}
	wg.Done()
	runtime.UnlockOSThread()
}

var connection_mutex sync.Mutex

func getwork(wallet_address string) {
	if m.Mission == 0 {
		m.Connection.Close()
		return
	}
	var err error

	for m.Mission == 1 {
		if m.Mission == 0 {
			break
		}
		u := url.URL{Scheme: "wss", Host: m.Daemon, Path: "/ws/" + wallet_address}
		globals.Logger.Info("[Miner] Connecting to ", "url", u.String())

		dialer := websocket.DefaultDialer
		dialer.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		m.Connection, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			globals.Logger.Info("[Miner] Will try in 10 secs", "server adress", m.Daemon)
			time.Sleep(10 * time.Second)

			continue
		}

		var result rpc.GetBlockTemplate_Result

	wait_for_another_job:

		if m.Mission == 0 {
			m.Connection.Close()
			break
		}

		if err = m.Connection.ReadJSON(&result); err != nil {
			globals.Logger.Error(err, "[Miner] Error connecting to server: %s\n", m.Daemon)
			continue
		}

		mutex.Lock()
		job = result
		job_counter++
		mutex.Unlock()
		if job.LastError != "" {
			globals.Logger.Error(nil, "[Miner] Received error", "err", job.LastError)
		}

		block_counter = job.Blocks
		mini_block_counter = job.MiniBlocks
		hash_rate = job.Difficultyuint64
		our_height = int64(job.Height)
		Difficulty = job.Difficultyuint64

		//fmt.Printf("[Miner] recv: %+v diff %d\n", result, Difficulty)
		goto wait_for_another_job
	}
}

func mineblock(tid int) {
	var diff big.Int
	var work [block.MINIBLOCK_SIZE]byte
	var random_buf [12]byte

	rand.Read(random_buf[:])

	time.Sleep(5 * time.Second)

	nonce_buf := work[block.MINIBLOCK_SIZE-5:] //since slices are linked, it modifies parent
	runtime.LockOSThread()
	threadaffinity()

	var local_job_counter int64

	i := uint32(0)

	for m.Mission == 1 {
		if m.Mission == 0 {
			break
		}
		mutex.RLock()
		myjob := job
		local_job_counter = job_counter
		mutex.RUnlock()

		n, err := hex.Decode(work[:], []byte(myjob.Blockhashing_blob))
		if err != nil || n != block.MINIBLOCK_SIZE {
			globals.Logger.Error(err, "[Miner] Blockwork could not decoded successfully", "blockwork", myjob.Blockhashing_blob, "n", n, "job", myjob)
			time.Sleep(time.Second)
			continue
		}

		copy(work[block.MINIBLOCK_SIZE-12:], random_buf[:]) // add more randomization in the mix
		work[block.MINIBLOCK_SIZE-1] = byte(tid)

		diff.SetString(myjob.Difficulty, 10)

		if work[0]&0xf != 1 { // check version
			globals.Logger.Error(nil, "[Miner] Unknown version, please check for updates", "version", work[0]&0x1f)
			time.Sleep(time.Second)
			continue
		}

		for local_job_counter == job_counter && m.Mission == 1 { // update job when it comes, expected rate 1 per second
			i++
			binary.BigEndian.PutUint32(nonce_buf, i)

			powhash := astrobwtv3.AstroBWTv3(work[:])
			atomic.AddUint64(&counter, 1)

			if CheckPowHashBig(powhash, &diff) == true { // note we are doing a local, NW might have moved meanwhile
				globals.Logger.Info("[Miner] Successfully found DERO miniblock (going to submit)", "difficulty", myjob.Difficulty, "height", myjob.Height)
				func() {
					defer globals.Recover(1)
					connection_mutex.Lock()
					defer connection_mutex.Unlock()
					m.Connection.WriteJSON(rpc.SubmitBlock_Params{JobID: myjob.JobID, MiniBlockhashing_blob: fmt.Sprintf("%x", work[:])})
				}()
			}
		}
	}
}
