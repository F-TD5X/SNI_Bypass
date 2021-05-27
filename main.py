# -*- coding: utf-8 -*-
import json
import os
import sys
import atexit
import subprocess
import CM

Host = []
config_file = 'config.json'
if os.name == 'nt':
    host_filename = 'c:/windows/system32/drivers/etc/hosts'
else:
    host_filename = '/etc/hosts'
caddy = ""


def hosts(host_list):
    global Host
    Host = Host + host_list
    return ', '.join('https://%s:443' % t for t in host_list)

def upstream(upstreams):
    return ' '.join('https://%s' % t for t in upstreams)

def entry(hosts, cert, upstream):
    return """
%s {
    tls %s.crt %s.key
    reverse_proxy  %s {
        lb_policy first
        lb_try_duration 30s
        fail_duration 5s
        header_up Host {host}
        transport http {
            dial_timeout 5s
            tls_insecure_skip_verify
        }
    }
}
""" % (hosts, cert, cert, upstream)

def process_item(item):
    return entry(hosts(item['hosts']), 'SSL', upstream(item['upstream']))

def onexit():
    caddy.kill()
    content = ''
    with open(host_filename, 'r', encoding='utf-8') as f:
        content = f.read()
    content = content[:content.find(
        "# SNI Bypass")] + content[content.rfind("# SNI Bypass") + len("# SNI Bypass"):-1]
    with open(host_filename, 'w', encoding='utf-8') as f:
        f.write(content)
    try:
        os.remove("SSL.crt")
        os.remove("SSL.key")
    except:
        pass

if __name__ == "__main__":
    conf = json.loads(open(config_file, 'r', encoding='utf-8').read())
    certificate_manager = CM.CertificateManager()
    if not os.path.exists(conf["CertificateManagerFile"]):
        certificate_manager.gen_CA()
        certificate_manager.trustCA()
        certificate_manager.save(conf["CertificateManagerFile"])
    else:
        certificate_manager.load(conf["CertificateManagerFile"])

    with open("Caddyfile", "w") as f:
        f.write('\n'.join(process_item(x)
                if x['enable'] else '' for x in conf["hosts"]))

    certificate_manager.gen_cert(Host)

    with open(host_filename, 'a') as f:
        f.write('\n# SNI Bypass\n' + '\n'.join('127.0.0.1 %s' %
                x for x in Host) + '\n# SNI Bypass\n')
    caddy = subprocess.Popen(
        'caddy run -config ./Caddyfile', shell=True, stdout=sys.stdout)
    try:
        caddy.wait()
    except KeyboardInterrupt:
        onexit()
    atexit.register(onexit)
