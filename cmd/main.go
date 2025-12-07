package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/odlev/xray-cli/pkg/marshaller"
	"github.com/odlev/xray-cli/pkg/parser"
	"github.com/odlev/xray-cli/pkg/systemd"
)

const (
	systemSystemdUnitPath          = "/etc/systemd/system"
	userSystemdUnitPathWithoutHome = ".config/systemd/user"
	systemdUnitName                = "xray-cli.service"
	xrayConfigName                 = "config.json"
	descriptionSystemdUnit         = "Xray Core service"
	restartPolicy                  = "on-failure"
)

func main() {
	var link string
	var port int
	// var filePath string
	var xrayPath string
	var systemdUnitPath string
	var wantedByTarget string

	flag.StringVar(&link, "link", "", "link for connection (in double quote)")
	flag.StringVar(&xrayPath, "x", "", "path to the xray binary")
	flag.IntVar(&port, "p", 10808, "Inbound socks5 port")
	// flag.StringVar(&filePath, "cfgf", "config.json", "name for custom file for config")
	flag.StringVar(&systemdUnitPath, "unit-path", "", "optional directory/path for xray-cli.service (require root), default /etc/systemd/system when run as root or ~/.config/systemd/user if not root")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -l <vless link> -x <path to binary xray core> [options]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	require("link",  link)
	require("xray binary path", xrayPath)
	// isSystemdUnitPath - переменная, указывающая на то, был ли передан пользователем кастомный путь для записи systemd unit'а
	systemdUnitPath = normalizeUnitPath(systemdUnitPath)

	absXray, err := filepath.Abs(xrayPath)
	if err != nil {
		fatal("invalid xray path:", err)
	}
	xrayDir := filepath.Dir(absXray)

	cfg, err := parser.Parse(link, port)
	if err != nil {
		fatal(err)
	}
	if cfg == nil {
		fatal("parser returned empty config")
	}

	configPath := filepath.Join(xrayDir, xrayConfigName)

	if err := marshaller.Marshal(cfg, configPath); err != nil {
		fatal(err)
	}
	isSuperUser := os.Geteuid() == 0
	// choice - переменная, определяющая выбор пользователя в создании systemd в системном пространстве или в пространстве пользователя
	var choice int
	// isSystemdUnitPath - переменная, указывающая на то, был ли передан пользователем кастомный путь для записи systemd unit'а
	if systemdUnitPath == "" {
		if isSuperUser {
			systemdUnitPath = filepath.Join(systemSystemdUnitPath, systemdUnitName)
			wantedByTarget = "multi-user.target"
		} else {
			log("-------Where should I create a systemd unit?-------")
			time.Sleep(time.Millisecond * 800)
			log("1) In the system space, required root (/etc/systemd/system)")
			time.Sleep(time.Millisecond * 500)
			log("2) In the user space (~/.config/systemd/user) only for current user")
			fmt.Scan(&choice)
			switch choice {
			case 1:
				log("Restart the program with sudo to create a systemd unit in the system")
				os.Exit(1)
			case 2:
				systemdUnitPath = filepath.Join(os.Getenv("HOME"), userSystemdUnitPathWithoutHome, systemdUnitName)
				wantedByTarget = "default.target"
			default:
				fatal("incorrect choice")
			}
		}
	}
	execStart := systemd.EscapedExecStart(absXray, configPath)
	if err := systemd.WriteUnit(systemdUnitPath, descriptionSystemdUnit, xrayDir, execStart, restartPolicy, wantedByTarget); err != nil {
		fatal(err)
	}

	if err := os.Chdir(xrayDir); err != nil {
		fmt.Fprintln(os.Stderr, "failed to chdir to xray dir:", err)
		os.Exit(1)
	}

	if choice == 2 {
		if err := exec.Command("systemctl", "--user", "daemon-reload").Run(); err != nil {
			fatal(err)
		}
		if err := exec.Command("systemctl", "--user", "enable", "xray-cli.service").Run(); err != nil {
			fatal("failed to exec command for enable systemd unit autostart:", err)
		}
		
		if err := exec.Command("systemctl", "--user", "start", "xray-cli.service").Run(); err != nil {
			fatal("failed to exec command for start systemd unit :", err)
		}
		b, err := exec.Command("journalctl", "--user-unit", "xray-cli.service", "-n 50", "--no-pager").CombinedOutput()
		if err != nil {
			fatal("failed to exec command for view logs:", err)
		}
		log(string(b))
	}
	if isSuperUser {
		if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
			fatal(err)
		}
		// автостарт сервиса
		if err := exec.Command("systemctl", "enable", "xray-cli.service").Run(); err != nil {
			fatal("failed to exec command for enable systemd unit autostart (system):", err)
		}
		// старт сервис
		if err := exec.Command("systemctl", "start", "xray-cli.service").Run(); err != nil {
			fatal("failed to exec command for start systemd unit (system):", err)
		}
		// логи сервиса
		b, err := exec.Command("journalctl", "-u", "xray-cli.service", "-n 50", "--no-pager").CombinedOutput()
		if err != nil {
			fatal("failed to exec command for view logs:", err)
		}
		log(string(b))
	}
	log("Xray success started, autostart enabled, you can view logs here")

}

func require(name, val string) {
    if val == "" {
        flag.Usage()
        fatal(name + " is required")
    }
}

func normalizeUnitPath(p string) string {
    if p == "" {
        return ""
    }
    cleaned := filepath.Clean(p)
    if filepath.Ext(cleaned) == ".service" {
        return filepath.Join(filepath.Dir(cleaned), systemdUnitName)
    }
    return filepath.Join(cleaned, systemdUnitName)
}

func fatal(args ...any) {
	fmt.Fprintln(os.Stderr, args...)
	os.Exit(1)

}
func log(msg string) {
	fmt.Println(msg)
}
