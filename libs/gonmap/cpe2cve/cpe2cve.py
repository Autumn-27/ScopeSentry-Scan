#!/usr/bin/python3

"""
Part of SubZero project: 
Written by David SORIA (Sibwara, dsoria@astar.org) in 2021
Do not sell in a commercial package
"""


from cpe import CPE
import argparse
import requests
import time
import pdfkit
import matplotlib.pyplot as plt

API = "https://services.nvd.nist.gov/rest/json/cves/1.0"

"""
format:
    {'cpe:/a:apache:tomcat:7.0.27': {
        'friendlyname': 'Tomcat 7.0.27',  # A human way to describe the porduct
        'total':15,                       # number of CVE concerning this CPE
        'highest': 9.8,                   # highest CVSS score in all the CVEs
        'remotelyExploitable': True       # True if there is an available exploit for a network vector
        'dangerous': 3                    # number of CVE with CVSS > 9 and remotely exploitable
        'cve': {
            'CVE-2021-25329': {
                'score': 4.4,                           # The CVSSv2 or CVSSv3 score
                'vector': 'AV:L/AC:M/Au:N/C:P/I:P/A:P'  # The full CVSS vector
                'exploit': False                        # True if exploits are available
                }
        }
    }
"""

def cpe2cve(cpe, api):
    # Check the format upside to understand how the data are stored in the vulns dict
    vulns = {}
    for c in cpe:
        vulns[c.cpe_str] = {"total":0, "cve":{}, "highest":0.0, "remotelyExploitable":False, "dangerous": 0, "friendlyname": f"{c.get_product()[0]} {c.get_version()[0]}"}
        # call to the NIST APÏ
        req = requests.get(f"{api}?cpeMatchString={c.cpe_str}&resultsPerPage=100")
        if req.status_code != 200:
            print(f"[!] An error occured with {c.cpe_str}: {req.text}")
            continue
        else:
            print(f"[+] {c.cpe_str} OK")
        # Sleep to avoid blacklist by the NIST servers
        time.sleep(0.5)
        vulns[c.cpe_str]['total'] = req.json().get('totalResults')
        cves = req.json()['result'].get('CVE_Items')
        if cves:
            for cve in cves:
                cveid = cve['cve']['CVE_data_meta']['ID']
                vulns[c.cpe_str]["cve"][cveid] = {"score": 0, "vector": '', "exploit": "Unknown"}
                # manage the case depending on the CVSS to be V2 or V3
                metric = "3" if "baseMetricV3" in cve['impact'] else "2"
                cvss = cve['impact'][f"baseMetricV{metric}"][f"cvssV{metric}"]
                vulns[c.cpe_str]["cve"][cveid]["score"] = cvss['baseScore']
                vulns[c.cpe_str]["cve"][cveid]["vector"] = cvss['vectorString']
                # check if vuln is exploitable from network
                if "AV:N" in cvss['vectorString']:
                    vulns[c.cpe_str]["remotelyExploitable"] = True
                    if cvss['baseScore'] > 9.0 :
                        vulns[c.cpe_str]["dangerous"] += 1
                # Update the highest risk if necessary
                vulns[c.cpe_str]['highest'] = max(vulns[c.cpe_str]['highest'], cvss['baseScore'])
    return vulns

def getCriticality(cvss):
    """ color convention fot the cells of the PDF """
    if cvss < 3.1:
        return ("low", "#ffff00", (255, 255, 0))
    if cvss < 6.1:
        return ("medium", "#ffc800", (255, 200, 0))
    if cvss < 9.1:
        return ("high", "#ff6400", (255, 100, 0))
    return ("critical", "#cc0000", (200, 0, 0))

def export2pdf(vulns, output):
    html = "<h1 style='font-size:40px;'>Vulnerability report</h1>"
    html += f"<p>Date: {time.ctime(time.time())}<br/></p>"
    html += f"<p style='background-color:#f8f8f8;padding:10px 10px;'><strong>{len(vulns)}</strong> products were <strong>submited</strong><br/>"
    vulnproduct = sum([vulns[i]['total']>0 for i in vulns])
    html += f"<strong>{vulnproduct}</strong> of them are <strong>vulnerable</strong><br/>"
    sumvuln = sum([vulns[i]['total'] for i in vulns])
    html += f"A total of <strong>{sumvuln}</strong> vulnerabilities were found<br/>"
    sumdangerous = sum([vulns[i]['dangerous'] for i in vulns])
    html += f"<strong>{sumdangerous}</strong> of them are critical AND remotely exploitable</p>"
    html += "<h2>Status by product</h2>"
    html += "<table><thead><tr><th>Product</th><th>CPE</th><th>CVE</th><th>Risk</th><th>Remotely exploitable</th></tr></thead><tbody>"
    fig, ax = plt.subplots()
    values = []
    labels = []
    colors = []
    maxgraph = 0
    for cpe,details in sorted(vulns.items(), reverse=True, key=lambda k_v: k_v[1]['highest']):
        crit = getCriticality(details['highest'])
        html += f"<tr><td>{details['friendlyname']}</td><td>{cpe}</td><td>{details['total']}</td><td style='background-color:{crit[1]};'>{details['highest']}</td><td>{details['remotelyExploitable']}</td></tr>"
        if sumvuln and details['total']/sumvuln > 0.02 and maxgraph < 11:
            maxgraph += 1
            values.append(details['total'])
            labels.append(details['friendlyname'])
            colors.append(crit[1])
    html += "</tbody></table>"
    ax.pie(values, labels=labels, colors=colors, autopct='%1i%%', wedgeprops={"edgecolor":"white",'linewidth': 1, 'linestyle': 'solid', 'antialiased': True} )
    plt.tight_layout()
    plt.savefig("/tmp/graph.png", bbox_inches='tight', transparent=True)
    if sumvuln:
        html += "<h2>Main vulnerabilities distribution</h2>"
        html += f"<p><img src ='/tmp/graph.png' width='65%' style='display:block;margin-left:auto;margin-right:auto;'/></p>"
    html += "<h2>Vulnerabilities details</h2>"
    html += "<table><thead><tr><th>CVE</th><th>Score</th><th>Remotely exploitable</th><th>Target</th><th>Vector</th></tr></thead><tbody>"
    for cpe, details in vulns.items():
        for cve, detailscve in details["cve"].items():
            crit = getCriticality(detailscve['score'])
            html += f"<tr><td>{cve}</td><td style='background-color:{crit[1]};'>{detailscve['score']}</td><td>{'AV:N' in detailscve['vector']}</td><td>{cpe}</td><td>{detailscve['vector']}</td></tr>"
    html += "</tbody></table>"
    if output:
        try:
            pdfkit.from_string(html, output, css="github.css", options={"enable-local-file-access": ""})
        except Exception as e:
            print(f"[!] an error occured during PDF generation: {e}\nAn html output is produced instead: {output}.html")
            with open(output+'.html', "w") as f:
                f.write(html)

def export2csv(vulns, output):
    lines = ["cpe,highestrisk,remotelyexploitable,vulnerabilities,cve,score,remote,exploit,vector"]
    for cpe,details in vulns.items():
        # counter for each vulnerability associated with a unique cpe
        cpt = 0
        # if there is no vuln, we put a line with a risk of 0.0 and 0 vuln
        if details["total"] == 0:
            lines.append(f"{cpe},0.0,False,0/0,,,,")
        for cve, detailscve in details["cve"].items():
            cpt += 1
            lines.append(f"{cpe},{details['highest']},{details['remotelyExploitable']},{cpt}/{details['total']},{cve},{detailscve['score']},{'AV:N' in detailscve['vector']},{detailscve['exploit']},{detailscve['vector']}")
    # if a filename was provided, write to a file
    if output:
        with open(output, "w") as f:
            for l in lines:
                f.write(l + '\n')
    # default is to print to stdout
    else:
        print(*(lines),sep='\n')

def export(vulns, output):
    # if a filename was provided we check its extension
    if output:
        filename = output.split('.')
        # if the extension is in PDF we use the PDF export function
        if len(filename) > 1 and filename[-1].lower() == "pdf":
            export2pdf(vulns, output)
        # if the extension is not explicitely PDF wu use CSV as default
        else:
            export2csv(vulns, output)
    # if no ouput was provided we print a CSV format to stdout
    # export2csv will recognize that output is empty and print to stdout instead of file
    else:
        export2csv(vulns, output)

def main():
    global API
    parser = argparse.ArgumentParser(description='Giving the CVEs that affect a given CPE', add_help=True)
    parser.add_argument('--cpe', '-c', action="store", default=None,
            help="Give one CPE string to the tool")
    parser.add_argument('--file', '-f', action="store", default=None,
            help="Import multiple CPE from a file (one per line)")
    parser.add_argument('--output', '-o', action="store", default=None,
            help="Write results in an output file. Extension is read to choice between PDF and CSV. Default is CSV")
    args = parser.parse_args()
    cpe = []
    print("[*] Check if CPE is well formed ...")
    try:
        if args.cpe:
            cpe.append(CPE(args.cpe))
        elif args.file:
            with open(args.file) as f:
                for line in f.readlines():
                    #remove space and newlines char from each line
                    l = line.lower().strip(' \n\r')
                    cpe.append(CPE(l))
        else:
            print("[!] indicate at least a CPE (--cpe) or an input file with one CPE per line (--file)")
            exit(1)
    except Exception as e:
        print(f"[!] Bad CPE format: {e}")
        exit(1)
    print("[+] Valid CPE")
    print(f"[*] Searching vulnerabilities for the {len(cpe)} CPE given")
    vulns = cpe2cve(cpe, API)
    print("[+] Vulnerabilities computed")
    print(f"[*] Export to {args.output if args.output else 'stdout'}")
    export(vulns, args.output)
    print("[+] Export completed !")

if __name__ == '__main__':
    main()