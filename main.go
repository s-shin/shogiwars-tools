package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
)

type Config struct {
	AppDir   AppDir `json:"app_dir"`
	User     string `json:"user"`
	SleepSec uint   `json:"sleep"`
	Debug    bool   `json:"debug"`
}

func newConfig(cc *cli.Context) (*Config, error) {
	c := &Config{
		AppDir:   AppDir(cc.GlobalString("appdir")),
		User:     cc.GlobalString("user"),
		SleepSec: cc.GlobalUint("sleep"),
		Debug:    cc.GlobalBool("debug"),
	}
	if c.AppDir == "" {
		return nil, fmt.Errorf("config: appdir is required")
	}
	if c.User == "" {
		return nil, fmt.Errorf("config: user is required")
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
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "Enable debug mode",
			EnvVar: "SHOGIWARS_DEBUG",
		},
	}

	app.Commands = []cli.Command{
		{
			Name: "sync",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "page, p",
					Usage: "'<start>' or '<start>,<end>'",
				},
			},
			Action: func(c *cli.Context) error {
				pageRange := []int{0, 20}
				{
					for i, s := range strings.Split(c.String("page"), ",") {
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
					page := &HistoryPage{config.User, TenMinutes, p}
					log.Printf("Fetching... (%s:%s:%d)\n", page.UserName, page.GameType, page.Page)
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
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
