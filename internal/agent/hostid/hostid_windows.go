package hostid

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// machineGUID is where Windows keeps its machine id.
const machineGUID = `HKLM\SOFTWARE\Microsoft\Cryptography\MachineGuid`

func sources() []source {
	return []source{{name: machineGUID, read: func() (string, error) {
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Cryptography`, registry.QUERY_VALUE|registry.WOW64_64KEY)
		if err != nil {
			return "", fmt.Errorf("open %s: %w", machineGUID, err)
		}
		defer k.Close()
		v, _, err := k.GetStringValue("MachineGuid")
		if err != nil {
			return "", fmt.Errorf("read %s: %w", machineGUID, err)
		}
		return clean(v)
	}}}
}
