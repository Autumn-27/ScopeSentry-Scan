// Package subdomainMode -----------------------------
// @file      : subdomainScan.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/9 23:22
// -------------------------------------------
package subdomainMode

//func SubdomainScan(Host []string) []types.SubdomainResult {
//	defer system.RecoverPanic("SubdomainScan")
//	subDomainResult := []types.SubdomainResult{}
//	resultCallback := func(Domains []string) {
//		// Do something with the msg in the context of the main function
//		system.SlogDebugLocal(fmt.Sprintf("Received message in main: %v", Domains))
//		_domain := types.SubdomainResult{}
//		_domain.Host = Domains[0]
//		_domain.Type = "A"
//		for i := 1; i < len(Domains); i++ {
//			containsSpace := strings.Contains(Domains[i], " ")
//			if containsSpace {
//				result := strings.SplitN(Domains[i], " ", 2)
//				_domain.Type = result[0]
//				_domain.Value = append(_domain.Value, result[1])
//			} else {
//				_domain.IP = append(_domain.IP, Domains[i])
//			}
//		}
//		time := system.GetTimeNow()
//		_domain.Time = time
//		subDomainResult = append(subDomainResult, _domain)
//	}
//
//	screenPrinter, _ := output.NewScreenOutput(false, resultCallback)
//
//	domains := Host
//	domainChanel := make(chan string)
//	go func() {
//		for _, d := range domains {
//			domainChanel <- d
//		}
//		close(domainChanel)
//	}()
//
//	system.KsubdomainOpt.Writer = []outputter.Output{
//		screenPrinter,
//	}
//	system.KsubdomainOpt.Domain = domainChanel
//	system.KsubdomainOpt.DomainTotal = len(domains)
//	system.KsubdomainOpt.Retry = 3
//	system.KsubdomainOpt.Check()
//	r, err := runner.New(&system.KsubdomainOpt)
//	if err != nil {
//		gologger.Fatalf(err.Error())
//	}
//	ctx := context.Background()
//	r.RunEnumeration(ctx)
//	r.Close()
//	return subDomainResult
//
//}
