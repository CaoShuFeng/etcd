// Copyright 2015 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io"
	"math"
	"os"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/report"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/time/rate"
	"gopkg.in/cheggaaa/pb.v1"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put",
	Short: "Benchmark put",

	Run: putFunc,
}

var (
	putTotal int
	putRate  int

	compactInterval   time.Duration
	compactIndexDelta int64
	reportFile        string
	keyFile           string
)

func init() {
	RootCmd.AddCommand(putCmd)
	putCmd.Flags().IntVar(&putRate, "rate", 0, "Maximum puts per second (0 is no limit)")

	putCmd.Flags().IntVar(&putTotal, "total", 10000, "Total number of put requests")
	putCmd.Flags().DurationVar(&compactInterval, "compact-interval", 0, `Interval to compact database (do not duplicate this with etcd's 'auto-compaction-retention' flag) (e.g. --compact-interval=5m compacts every 5-minute)`)
	putCmd.Flags().Int64Var(&compactIndexDelta, "compact-index-delta", 1000, "Delta between current revision and compact revision (e.g. current revision 10000, compact at 9000)")
	putCmd.Flags().StringVar(&reportFile, "report-file", "-", "where report file is saved, - means stdout")
	putCmd.Flags().StringVar(&keyFile, "key-file", "-", "where key file is saved, - means stdout")
}

func putFunc(cmd *cobra.Command, args []string) {
	requests := make(chan v3.Op, totalClients)
	if putRate == 0 {
		putRate = math.MaxInt32
	}
	limit := rate.NewLimiter(rate.Limit(putRate), 1)
	clients := mustCreateClients(totalClients, totalConns)

	bar = pb.New(putTotal)
	bar.Format("Bom !")
	bar.Start()

	r := newReport()
	for i := range clients {
		wg.Add(1)
		go func(c *v3.Client) {
			defer wg.Done()
			for op := range requests {
				limit.Wait(context.Background())

				st := time.Now()
				_, err := c.Do(context.Background(), op)
				r.Results() <- report.Result{Err: err, Start: st, End: time.Now()}
				bar.Increment()
			}
		}(clients[i])
	}
	// 100 must be big enough
	keys := make(chan string, 100)
	go func() {
		var file io.Writer
		if keyFile == "-" {
			file = os.Stdout
		} else {
			f, err := os.OpenFile(keyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				panic(err)
			}
			defer f.Close()
			file = f
		}
		for key := range keys {
			fmt.Fprintf(file, "%s\n", key)
		}
	}()
	go func() {
		for i := 0; i < putTotal; i++ {
			k8sKey := getAuditKey()
			k8sValue := getAuditJson()
			keys <- k8sKey
			requests <- v3.OpPut(k8sKey, k8sValue)

		}
		close(requests)
		close(keys)
	}()

	if compactInterval > 0 {
		go func() {
			for {
				time.Sleep(compactInterval)
				compactKV(clients)
			}
		}()
	}

	rc := r.Run()
	wg.Wait()
	close(r.Results())
	bar.Finish()
	res := <-rc
	if reportFile == "-" {
		fmt.Println(res)
	} else {
		f, err := os.OpenFile(reportFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		fmt.Fprintln(f, res)
	}
}

func compactKV(clients []*v3.Client) {
	var curRev int64
	for _, c := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		resp, err := c.KV.Get(ctx, "foo")
		cancel()
		if err != nil {
			panic(err)
		}
		curRev = resp.Header.Revision
		break
	}
	revToCompact := max(0, curRev-compactIndexDelta)
	for _, c := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err := c.KV.Compact(ctx, revToCompact)
		cancel()
		if err != nil {
			panic(err)
		}
		break
	}
}

func max(n1, n2 int64) int64 {
	if n1 > n2 {
		return n1
	}
	return n2
}
