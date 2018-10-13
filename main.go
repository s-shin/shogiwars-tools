package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

type Config struct {
	AppDir   AppDir `json:"app_dir"`
	SleepSec uint   `json:"sleep"`
	Debug    bool   `json:"debug"`
}

func newConfig(cc *cli.Context) (*Config, error) {
	c := &Config{
		AppDir:   AppDir(cc.GlobalString("appdir")),
		SleepSec: cc.GlobalUint("sleep"),
		Debug:    cc.GlobalBool("debug"),
	}
	if c.AppDir == "" {
		return nil, fmt.Errorf("config: appdir is required")
	}
	return c, nil
}

func main() {
	app := cli.NewApp()
	app.Name = "shogiwars-tools"
	app.Usage = "Tool set for ShogiWars"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "appdir, d",
			Usage:  "`PATH` to application directory",
			Value:  ".shogiwars",
			EnvVar: "SHOGIWARS_APP_DIR",
		},
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "Enable debug mode",
			EnvVar: "SHOGIWARS_DEBUG",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "sync",
			Usage: "create record index by scraping history pages",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "user, u",
					Usage:  "Target `USERNAME`",
					EnvVar: "SHOGIWARS_USER",
				},
				cli.UintFlag{
					Name:   "sleep, s",
					Usage:  "Interval per HTTP request by `SECONDS`",
					Value:  3,
					EnvVar: "SHOGIWARS_SLEEP",
				},
				cli.StringFlag{
					Name:  "page, p",
					Usage: fmt.Sprintf(`"<start>" or "<start>-<end>". %d items per page.`, NumRecordItemsPerPage),
				},
			},
			Action: func(c *cli.Context) error {
				pageRange := []int{0, 20}
				{
					for i, s := range strings.Split(c.String("page"), "-") {
						if s == "" {
							break
						}
						n, err := strconv.Atoi(s)
						if err != nil {
							return err
						}
						pageRange[i] = n
						if i == 0 {
							pageRange[1] = n
						}
						if i == 1 {
							break
						}
					}
				}

				config, err := newConfig(c)
				if err != nil {
					return err
				}

				userName := c.String("user")
				if userName == "" {
					return fmt.Errorf("config: user is required")
				}

				log.Println("Initialize application directory...")
				if err := config.AppDir.Init(); err != nil {
					return err
				}

				log.Println("Load application data...")
				appData, err := config.AppDir.LoadData()
				if err != nil {
					return err
				}
				mgr := NewAppDataManager(appData)
				numNewRecordItems := 0
				for p := pageRange[0]; p <= pageRange[1]; p++ {
					page := &HistoryPage{userName, TenMinutes, p}
					log.Printf("Fetching... (user: %s, game: %s, page: %d)\n", page.UserName, page.GameType, page.Page)
					log.Printf("# URL: %s\n", page.BuildURL())
					items, err := page.FetchRecordItems()
					if err != nil {
						return err
					}
					log.Printf("%d items were fetched.\n", len(items))
					if len(items) == 0 {
						break
					}
					n := mgr.AppendRecordItems(items...)
					numNewRecordItems += n
					if n != len(items) {
						log.Println("No new items.")
						break
					}
					time.Sleep(time.Duration(config.SleepSec) * time.Second)
				}
				if numNewRecordItems > 0 {
					log.Printf("Saving... (%d new items)\n", numNewRecordItems)
					if err := config.AppDir.SaveData(appData); err != nil {
						return err
					}
				}
				log.Println("Done.")
				return nil
			},
		},
		{
			Name: "list",
			Flags: []cli.Flag{
				cli.UintFlag{
					Name:  "limit",
					Value: 100,
				},
				cli.UintFlag{
					Name:  "offset",
					Value: 0,
				},
				cli.BoolFlag{
					Name: "desc",
				},
				cli.BoolFlag{
					Name: "tsv",
				},
				cli.BoolFlag{
					Name: "skip-header",
				},
			},
			Action: func(c *cli.Context) error {
				limit := c.Uint("limit")
				offset := c.Uint("offset")
				desc := c.Bool("desc")
				tsv := c.Bool("tsv")
				skipHeader := c.Bool("skip-header")

				config, err := newConfig(c)
				if err != nil {
					return err
				}
				appData, err := config.AppDir.LoadData()
				if err != nil {
					return err
				}

				var renderer ASCIIRenderer
				if tsv {
					renderer = NewTsvRenderer(os.Stdout)
				} else {
					table := tablewriter.NewWriter(os.Stdout)
					table.SetAlignment(tablewriter.ALIGN_LEFT)
					renderer = table
				}
				if !skipHeader {
					renderer.SetHeader([]string{"Date", "Black", "White", "Winner", "Record ID"})
				}
				items := SortRecordItemsByDate(appData.RecordItems, !desc)
				n := uint(0)
				for i, item := range items {
					if uint(i) < offset {
						continue
					}
					renderer.Append([]string{
						item.Date.Format("2006-01-02 15:04:05"),
						fmt.Sprintf("%s (%s)", item.Players.Black.UserName, item.Players.Black.Rank),
						fmt.Sprintf("%s (%s)", item.Players.White.UserName, item.Players.White.Rank),
						item.Winner,
						string(item.RecordID),
					})
					n++
					if n == limit {
						break
					}
				}
				renderer.Render()
				return nil
			},
		},
		{
			Name: "get",
			Action: func(c *cli.Context) error {
				recordID := c.Args().Get(0)

				page := &GamePage{RecordID: recordID}
				log.Printf("# URL: %s\n", page.BuildURL())

				record, err := page.FetchRecord()
				if err != nil {
					return err
				}
				data, err := json.Marshal(record)
				if err != nil {
					return err
				}
				fmt.Println(string(data))

				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
