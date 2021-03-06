package dns

import (
	"bytes"
	"fmt"
	"github.com/code-ready/crc/pkg/crc/logging"
	"github.com/code-ready/crc/pkg/crc/network"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/code-ready/crc/pkg/crc/services"
	crcos "github.com/code-ready/crc/pkg/os"
)

const (
	resolverFileTemplate = `port {{ .Port }}
nameserver {{ .IP }}
search_order {{ .SearchOrder }}`
)

type resolverFileValues struct {
	Port        int
	IP          string
	SearchOrder int
}

func runPostStartForOS(serviceConfig services.ServicePostStartConfig, result *services.ServicePostStartResult) (services.ServicePostStartResult, error) {
	// Write resolver file
	success, err := createResolverFile(serviceConfig.IP, filepath.Join("/", "etc", "resolver", serviceConfig.BundleMetadata.ClusterInfo.BaseDomain))
	if !success {
		result.Success = success
		return *result, err
	}
	// Restart the Network on mac
	logging.InfoF("Restarting the host network")
	success, err = restartNetwork()
	// we pass the result and error on
	result.Success = success
	return *result, err
}

func createResolverFile(InstanceIP string, path string) (bool, error) {
	var resolverFile bytes.Buffer

	values := resolverFileValues{
		Port:        dnsServicePort,
		IP:          InstanceIP,
		SearchOrder: 1,
	}

	t, err := template.New("resolver").Parse(resolverFileTemplate)
	if err != nil {
		return false, err
	}
	err = t.Execute(&resolverFile, values)
	if err != nil {
		return false, err
	}

	err = ioutil.WriteFile(path, resolverFile.Bytes(), 0644)
	if err != nil {
		return false, err
	}

	return true, nil
}

func restartNetwork() (bool, error) {
	// https://medium.com/@kumar_pravin/network-restart-on-mac-os-using-shell-script-ab19ba6e6e99
	netDeviceList, _, err := crcos.RunWithDefaultLocale("networksetup", "-listallnetworkservices")
	netDeviceList = strings.TrimSpace(netDeviceList)
	if err != nil {
		return false, err
	}
	for _, netdevice := range strings.Split(netDeviceList, "\n")[1:] {
		time.Sleep(1 * time.Second)
		stdout, stderr, err := crcos.RunWithDefaultLocale("networksetup", "-setnetworkserviceenabled", netdevice, "off")
		logging.DebugF("Disabling the %s Device (stdout: %s), (stderr: %s)", netdevice, stdout, stderr)
		stdout, stderr, err = crcos.RunWithDefaultLocale("networksetup", "-setnetworkserviceenabled", netdevice, "on")
		logging.DebugF("Enabling the %s Device (stdout: %s), (stderr: %s)", netdevice, stdout, stderr)
		if err != nil {
			return false, fmt.Errorf("%s: %v", stderr, err)
		}
	}

	waitForNetwork()

	return true, err
}

// Wait for Network wait till the network is up, since it is required to resolve external dnsquery
func waitForNetwork() {
	hostResolv, err := network.GetResolvValuesFromHost()
	if err != nil {
		logging.WarnF("Unable to read host resolv file (%v)", err)
	}
	// retry for 5 times
	for i := 0; i < 5; i++ {
		for _, ns := range hostResolv.NameServers {
			_, _, err := crcos.RunWithDefaultLocale("ping", "-c4", ns.IPAddress)
			if err == nil {
				return
			}
		}
	}
	logging.Warn("Host is not connected to internet.")
}
