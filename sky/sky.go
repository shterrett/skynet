package main

import (
	"flag"
	"fmt"
	"github.com/bketelsen/skynet/skylib"
	"os"
	"strconv"
  "bufio"
  "syscall"
  "strings"
)

var VersionFlag *string = flag.String("version", "", "service version")
var ServiceNameFlag *string = flag.String("service", "", "service name")
var HostFlag *string = flag.String("host", "", "host")
var RegionFlag *string = flag.String("region", "", "region")
var RegisteredFlag *string = flag.String("registered", "", "registered")

var DC skylib.DoozerConnection

func main() {
	flag.Parse()

	query := &skylib.Query{
		DoozerConn: Doozer(),
		Service:    *ServiceNameFlag,
		Version:    *VersionFlag,
		Host:       *HostFlag,
		Region:     *RegionFlag,
	}

	switch flag.Arg(0) {
	case "help", "h":
		CommandLineHelp()
	case "services":
		ListServices(query)
	case "hosts":
		ListHosts(query)
	case "regions":
		ListRegions(query)
	case "instances":
		ListInstances(query)
	case "versions":
		ListServiceVersions(query)
	case "topology":
		PrintTopology(query)
  case "cli":
    InteractiveShell()

	default:
		CommandLineHelp()
	}
}

func Doozer() (skylib.DoozerConnection) {
  if DC == nil {
    DC = Connect()
  }

  return DC
}

func Connect() (skylib.DoozerConnection) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Failed to connect to Doozer")
			os.Exit(1)
		}
	}()

  // TODO: This needs to come from command line, or environment variable
  conn := skylib.NewDoozerConnection("127.0.0.1:8046", "", false, nil) // nil as the last param will default to a Stdout logger
  conn.Connect()

  return conn
}

func ListInstances(q *skylib.Query) {
  var regFlag *bool

  if *RegisteredFlag == "true" {
    b := true
    regFlag = &b
  } else if *RegisteredFlag == "false" {
    b := false
    regFlag = &b
  }

  q.Registered = regFlag

	results := q.FindInstances()

	for _, instance := range *results {
    registered := ""

    if(instance.Registered){
      registered = " [Registered]"
    }

		fmt.Println(instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port) + " - " + instance.Config.Name + " " + instance.Config.Version + registered)
	}
}

func ListHosts(q *skylib.Query) {
	results := q.FindHosts()

	for _, host := range *results {
		fmt.Println(host)
	}
}

func ListRegions(q *skylib.Query) {
	results := q.FindRegions()

	for _, region := range *results {
		fmt.Println(region)
	}
}

func ListServices(q *skylib.Query) {
	results := q.FindServices()

	for _, service := range *results {
		fmt.Println(service)
	}
}

func ListServiceVersions(q *skylib.Query) {
	if *ServiceNameFlag == "" {
		fmt.Println("Service name is required")
    return
	}

	results := q.FindServiceVersions()

	for _, version := range *results {
		fmt.Println(version)
	}
}

func PrintTopology(q *skylib.Query) {
	topology := make(map[string]map[string]map[string]map[string][]*skylib.Service)

	results := q.FindInstances()

	// Build topology hash first
	for _, instance := range *results {
		if topology[instance.Config.Region] == nil {
			topology[instance.Config.Region] = make(map[string]map[string]map[string][]*skylib.Service)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress] = make(map[string]map[string][]*skylib.Service)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name] = make(map[string][]*skylib.Service)
		}

		if topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] == nil {
			topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] = make([]*skylib.Service, 0)
		}

		topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version] = append(topology[instance.Config.Region][instance.Config.ServiceAddr.IPAddress][instance.Config.Name][instance.Config.Version], instance)
	}

	// Now we can print the correct heirarchy
	for regionName, region := range topology {
		fmt.Println("Region: " + regionName)

		for hostName, host := range region {
			fmt.Println("\tHost: " + hostName)

			for serviceName, service := range host {
				fmt.Println("\t\tService: " + serviceName)

				for versionName, version := range service {
					fmt.Println("\t\t\tVersion: " + versionName)

					for _, instance := range version {
            registered := ""

            if(instance.Registered){
              registered = " [Registered]"
            }

						fmt.Println("\t\t\t\t" + instance.Config.ServiceAddr.IPAddress + ":" + strconv.Itoa(instance.Config.ServiceAddr.Port) + registered)
					}
				}
			}
		}
	}
}

func CommandLineHelp() {
	fmt.Println("Usage:\n\t sky -option1=value -option2=value command <arguments>")

	fmt.Print(
		"\nCommands:\n" +
			"\n\tcli: Interactive shell for executing commands against skynet cluster" +
			"\n\thosts: List all hosts available that meet the specified criteria" +
			"\n\t\t-service - limit results to hosts running the specified service" +
			"\n\t\t-version - limit results to hosts running the specified version of the service (-service required)" +
			"\n\t\t-region - limit results to hosts in the specified region" +

			"\n\tinstances: List all instances available that meet the specified criteria" +
			"\n\t\t-service - limit results to instances of the specified service" +
			"\n\t\t-version - limit results to instances of the specified version of service" +
			"\n\t\t-region - limit results to instances in the specified region" +
			"\n\t\t-host - limit results to instances on the specified host" +
			"\n\t\t-registered - (true, false) limit results to instances that are registered (accepting requests)" +

			"\n\tregions: List all regions available that meet the specified criteria" +

			"\n\tservices: List all services available that meet the specified criteria" +
			"\n\t\t-host - limit results to the specified host" +
			"\n\t\t-region - limit results to hosts in the specified region" +

			"\n\n\tversions: List all services available that meet the specified criteria" +
			"\n\t\t-service - service name (required)" +
			"\n\t\t-host - limit results to the specified host" +
			"\n\t\t-region - limit results to hosts in the specified region" +

			"\n\n\ttopology: Print detailed heirarchy of regions/hosts/services/versions/instances" +
			"\n\t\t-service - limit results to instances of the specified service" +
			"\n\t\t-version - limit results to instances of the specified version of service" +
			"\n\t\t-region - limit results to instances in the specified region" +
			"\n\t\t-host - limit results to instances on the specified host" +

			"\n\n\n")
}


/*
 * CLI Logic
 */

func InteractiveShell(){
	lineReader := bufio.NewReader(os.Stdin)
  doozer := Doozer()

  fmt.Println("Skynet Interactive Shell")
  prompt()

  query := &skylib.Query{
    DoozerConn: doozer,
  }

	for {
		l, _, e := lineReader.ReadLine()
		if e != nil { break }

		s := string(l)
    parts := strings.Split(s, " ")

    switch(parts[0]) {
    case "exit":
      syscall.Exit(0)
    case "help", "h":
      InteractiveShellHelp()
    case "services":
      ListServices(query)
    case "hosts":
      ListHosts(query)
    case "regions":
      ListRegions(query)
    case "instances":
      ListInstances(query)
    case "versions":
      ListServiceVersions(query)
    case "topology":
      PrintTopology(query)

    case "service":
      if len(parts) >= 2 {
        query.Service = parts[1]
      }
      
      fmt.Printf("Service: %v\n", query.Service)

    case "host":
      if len(parts) >= 2 {
        query.Host = parts[1]
      }
      
      fmt.Printf("Host: %v\n", query.Host)

    case "version":
      if len(parts) >= 2 {
        query.Version = parts[1]
      }
      
      fmt.Printf("Version: %v\n", query.Version)

    case "region":
      if len(parts) >= 2 {
        query.Region = parts[1]
      }
      
      fmt.Printf("Region: %v\n", query.Region)

    case "registered":
      if len(parts) >= 2 {
        var reg bool

        if(parts[1] == "true"){
          reg = true
        } else {
          reg = false
        }

        query.Registered = &reg
      }

      registered := ""
      if query.Registered != nil {
        registered = strconv.FormatBool(*query.Registered)
      }
      
      fmt.Printf("Registered: %v\n", registered)


    case "reset":
      if len(parts) == 1 || parts[1] == "service" {
        query.Service = ""
      }

      if len(parts) == 1 || parts[1] == "version" {
        query.Version = ""
      }

      if len(parts) == 1 || parts[1] == "host" {
        query.Host = ""
      }

      if len(parts) == 1 || parts[1] == "region" {
        query.Region = ""
      }

      if len(parts) == 1 || parts[1] == "registered" {
        query.Registered = nil
      }
    case "filters":
      registered := ""
      if query.Registered != nil {
        registered = strconv.FormatBool(*query.Registered)
      }

      fmt.Printf("Region: %v\nHost: %v\nService:%v\nVersion: %v\nRegistered: %v\n", query.Region, query.Host, query.Service, query.Version, registered)
    default:
      fmt.Println("Unknown Command - type 'help' for a list of commands")
    }

    prompt()
	}
}

func InteractiveShellHelp(){
	fmt.Print(
		"\nCommands:" +
			"\n\thosts: List all hosts available that meet the specified criteria" +
			"\n\tinstances: List all instances available that meet the specified criteria" +
			"\n\tregions: List all regions available that meet the specified criteria" +
			"\n\tservices: List all services available that meet the specified criteria" +
			"\n\tversions: List all services available that meet the specified criteria" +

			"\n\ttopology: Print detailed heirarchy of regions/hosts/services/versions/instances" +

      "\n\nFilters:" +
      "\n\tfilters - list current filters" +
      "\n\treset <filter> - reset all filters or specified filter" +
      "\n\tregion <region> - Set region filter, all commands will be scoped to this region until reset" +
      "\n\tservice <service> - Set service filter, all commands will be scoped to this service until reset" +
      "\n\tversion <version> - Set version filter, all commands will be scoped to this version until reset" +
      "\n\thost <host> - Set host filter, all commands will be scoped to this host until reset" +

			"\n\n")
}

func prompt(){
  fmt.Printf("> ")
}