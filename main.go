package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
)

type Memory struct {
	Value int    `xml:",chardata"`
	Unit  string `xml:"unit,attr"`
}

type VCPU struct {
	Value     int    `xml:",chardata"`
	Placement string `xml:"placement,attr"`
}

type GuestOsType struct {
	Text    string `xml:",chardata"`
	Arch    string `xml:"arch,attr"`
	Machine string `xml:"machine,attr"`
}

type Boot struct {
	Dev string `xml:"dev,attr"`
}

type GuestOs struct {
	GuestOsType GuestOsType `xml:"type"`
	Boot        Boot        `xml:"boot"`
}

type Clock struct {
	Offset string `xml:"offset,attr"`
}

type Driver struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type Source struct {
	File string `xml:"file,attr"`
}

type Target struct {
	Dev string `xml:"dev,attr"`
	Bus string `xml:"bus,attr"`
}

type Alias struct {
	Name string `xml:"name,attr"`
}

type Disk struct {
	Type   string `xml:"type,attr"`
	Device string `xml:"device,attr"`
	Driver Driver `xml:"driver"`
	Source Source `xml:"source"`
	Target Target `xml:"target"`
	Alias  Alias  `xml:"alias"`
}

type MacAddress struct {
	Address string `xml:"address,attr"`
}

type NetworkSource struct {
	Bridge string `xml:"bridge,attr"`
}

type NetworkTarget struct {
	Dev string `xml:"dev,attr"`
}

type NetworkModel struct {
	Type string `xml:"type,attr"`
}

type NetworkAlias struct {
	Name string `xml:"name,attr"`
}

type NetworkAddress struct {
	Type     string `xml:"type,attr"`
	Domain   string `xml:"domain,attr"`
	Bus      string `xml:"bus,attr"`
	Slot     string `xml:"slot,attr"`
	Function string `xml:"function,attr"`
}

type NetworkInterface struct {
	Type           string         `xml:"type,attr"`
	MacAddress     MacAddress     `xml:"mac"`
	NetworkSource  NetworkSource  `xml:"source"`
	NetworkTarget  NetworkTarget  `xml:"target"`
	NetworkModel   NetworkModel   `xml:"model"`
	NetworkAlias   NetworkAlias   `xml:"alias"`
	NetworkAddress NetworkAddress `xml:"address"`
}

type Listen struct {
	Type    string `xml:"type,attr"`
	Address string `xml:"address,attr"`
}

type Graphics struct {
	Type       string `xml:"type,attr"`
	Port       string `xml:"port,attr"`
	Autoport   string `xml:"autoport,attr"`
	AttrListen string `xml:"listen,attr"`
	Listen     Listen `xml:"listen"`
}

type Devices struct {
	Emulator         string           `xml:"emulator"`
	Disk             []Disk           `xml:"disk"`
	NetworkInterface NetworkInterface `xml:"interface"`
	Graphics         Graphics         `xml:"graphics"`
}

type Domain struct {
	XMLName    xml.Name `xml:"domain"`
	DomainType string   `xml:"type,attr"`
	ID         string   `xml:"id,attr"`
	Name       string   `xml:"name"`
	UUID       string   `xml:"uuid"`
	Memory     Memory   `xml:"memory"`
	VCPU       VCPU     `xml:"vcpu"`
	GuestOs    GuestOs  `xml:"os"`
	Clock      Clock    `xml:"clock"`
	Devices    Devices  `xml:"devices"`
}

func (domain Domain) writeXML() {
	output, err := xml.Marshal(domain)
	if err != nil {
		fmt.Println("error: %v\n", err)
	}
	os.Stdout.Write(output)
}

func main() {
	c, err := net.DialTimeout("tcp", "10.30.0.1:16509", 2*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
	}

	l := libvirt.New(c)
	if err := l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	v, err := l.Version()
	if err != nil {
		log.Fatalf("failed to retrieve libvirt version: %v", err)
	}
	fmt.Println("Version:", v)

	domains, err := l.Domains()
	if err != nil {
		log.Fatalf("failed to retrieve domains: %v", err)
	}

	fmt.Println("ID\tName\t\tUUID")
	fmt.Printf("--------------------------------------------------------\n")
	for _, d := range domains {
		fmt.Printf("%d\t%s\t%x\n", d.ID, d.Name, d.UUID)
	}

	fmt.Println("--------------------------------------------------------")

	fmt.Println("Printing XML...")
	memory := Memory{Value: 2048, Unit: "KiB"}
	vcpu := VCPU{Value: 2, Placement: "static"}
	clock := Clock{Offset: "utc"}

	guestOs := GuestOs{
		GuestOsType{Text: "hvm", Arch: "x86_64", Machine: "pc-i440fx-bionic"},
		Boot{Dev: "hd"},
	}

	driver := Driver{Name: "qemu", Type: "qcow2"}
	target := Target{Dev: "vda", Bus: "virtio"}

	primaryDisk := Disk{
		Type:   "file",
		Device: "disk",
		Driver: driver,
		Source: Source{File: "/var/lib/libvirt/images/mahakam-test-vm"},
		Target: target,
		Alias:  Alias{Name: "virtio-disk0"},
	}

	secondaryDisk := Disk{
		Type:   "file",
		Device: "disk",
		Driver: driver,
		Source: Source{File: "/var/lib/libvirt/images/mahakam-test-vm-secondary"},
		Target: target,
		Alias:  Alias{Name: "virtio-disk1"},
	}

	networkInterface := NetworkInterface{
		Type:          "bridge",
		MacAddress:    MacAddress{Address: "a4:58:3b:0a:fd:3b"},
		NetworkSource: NetworkSource{Bridge: "virbr0"},
		NetworkTarget: NetworkTarget{Dev: "vnet21"},
		NetworkModel:  NetworkModel{Type: "virtio"},
		NetworkAlias:  NetworkAlias{Name: "net0"},
		NetworkAddress: NetworkAddress{
			Type:     "pci",
			Domain:   "0x0000",
			Bus:      "0x00",
			Slot:     "0x03",
			Function: "0x0",
		},
	}

	graphics := Graphics{
		Type:       "spice",
		Port:       "5921",
		Autoport:   "yes",
		AttrListen: "127.0.0.1",
		Listen: Listen{
			Type:    "address",
			Address: "127.0.0.1",
		},
	}

	devices := Devices{
		Emulator: "/usr/bin/kvm-spice",
		Disk: []Disk{
			primaryDisk,
			secondaryDisk,
		},
		NetworkInterface: networkInterface,
		Graphics:         graphics,
	}

	domain := Domain{
		DomainType: "kvm",
		ID:         "1234",
		Name:       "mahakam-libvirt-spike",
		UUID:       "1234",
		Memory:     memory,
		VCPU:       vcpu,
		GuestOs:    guestOs,
		Clock:      clock,
		Devices:    devices,
	}
	domain.writeXML()

	if err := l.Disconnect(); err != nil {
		log.Fatalf("failed to disconnect: %v", err)
	}
}
