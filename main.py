# -*- coding: utf-8 -*-
import json
import os
import sys
import atexit
from OpenSSL import crypto
import random
import subprocess

Host = []
if os.name == 'nt':
    host_filename = 'c:/windows/system32/drivers/etc/hosts'
else:
    host_filename ='/etc/hosts'
issuer = {
    "C": "CN",
    "ST": "CN",
    "L": "CN",
    "O": "SNI Bypass",
    "CN": "SNI Bypass"
}
CA_SERIAL= ""
caddy = ""

def hosts(host_list):
    global Host 
    Host = Host + host_list
    return ', '.join('https://%s:443' % t for t in host_list)

def upstream(upstreams):
    return ' '.join('https://%s' % t for t in upstreams)

def entry(hosts,cert,upstream):
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
        }
    }
}
""" % (hosts,cert,cert,upstream)
def process_item(item):
    return entry(hosts(item['hosts']),'SSL',upstream(item['upstream']))

def gen_CA():
    ca_key = crypto.PKey()
    ca_key.generate_key(crypto.TYPE_RSA, 2048)

    ca_cert = crypto.X509()
    ca_cert.set_version(2)
    ca_cert.gmtime_adj_notBefore(0)
    ca_cert.gmtime_adj_notAfter(10 * 365 * 24 * 60 * 60)
    ca_cert.set_serial_number(random.randint(50000000, 100000000))

    ca_cert.get_subject().commonName = "This CA is my Root CA"
    ca_cert.get_subject().C = issuer['C']
    ca_cert.get_subject().ST = issuer["ST"]
    ca_cert.get_subject().L = issuer["L"]
    ca_cert.get_subject().O = issuer["O"]
    ca_cert.get_subject().CN = issuer["CN"]

    ca_cert.set_issuer(ca_cert.get_subject())
    ca_cert.set_pubkey(ca_key)

    ca_cert.add_extensions([
        crypto.X509Extension(b"subjectKeyIdentifier", False, b"hash", subject=ca_cert),
    ])

    ca_cert.add_extensions([
        crypto.X509Extension(b"authorityKeyIdentifier", False, b"keyid:always,issuer", issuer=ca_cert),
    ])

    ca_cert.add_extensions([
        crypto.X509Extension(b"basicConstraints", True, b"CA:TRUE"),
        crypto.X509Extension(b"keyUsage", True, b"digitalSignature, keyCertSign, cRLSign"),
    ])

    ca_cert.sign(ca_key, 'sha256')

    # Save certificate
    with open("ca.crt", "wt") as f:
        f.write(crypto.dump_certificate(crypto.FILETYPE_PEM, ca_cert).decode("utf-8"))

    # Save private key
    with open("ca.key", "wt") as f:
        f.write(crypto.dump_privatekey(crypto.FILETYPE_PEM, ca_key).decode("utf-8"))

def gen_cert(hosts):
    CA_key = crypto.load_privatekey(crypto.FILETYPE_PEM,open('ca.key','r').read())
    CA_crt = crypto.load_certificate(crypto.FILETYPE_PEM,open('ca.crt','r').read())

    cert = crypto.X509()
    key = crypto.PKey()
    key.generate_key(crypto.TYPE_RSA,2048)
    cert.gmtime_adj_notBefore(0)
    cert.gmtime_adj_notAfter(10 * 365 * 24 * 60 * 60)
    cert.set_serial_number(random.randint(50000000, 100000000))

    cert.set_version(2)
    cert.set_pubkey(key)
    cert.set_subject(CA_crt.get_subject())
    cert.set_issuer(CA_crt.get_issuer())
    
    cert.add_extensions([
        crypto.X509Extension(b'subjectAltName',False,','.join('DNS:%s' % x for x in hosts).encode()),
        crypto.X509Extension(b"basicConstraints", True, b"CA:FALSE"),
        crypto.X509Extension(b"keyUsage", True, b"Digital Signature, Non Repudiation, Key Encipherment, Data Encipherment"),
        crypto.X509Extension(b"extendedKeyUsage",True,b"TLS Web Server Authentication, TLS Web Client Authentication, Code Signing, E-mail Protection"),
        crypto.X509Extension(b"authorityKeyIdentifier", False, b"keyid", issuer=CA_crt),
        crypto.X509Extension(b"subjectKeyIdentifier", False, b"hash", subject=cert),
    ])

    
    cert.sign(CA_key,'sha256')
    open("SSL.crt",'wb').write(crypto.dump_certificate(crypto.FILETYPE_PEM,cert))
    open("SSL.key",'wb').write(crypto.dump_privatekey(crypto.FILETYPE_PEM,key))


def trustCA():
    if os.name=='nt':
        os.system('certutil.exe -addstore "ROOT" ./ca.crt')
    else:
        os.system('cp ./ca.crt /usr/share/ca-certificates/trust-source/anchors/SNI_BYPASS.crt')
        if os.path.exists('/usr/bin/update-ca-trust'):
            os.system('update-ca-trust')
        else:
            os.system('update-ca-certificates')

def delCA():
    if os.name == 'nt':
        os.system(f'certutil -delstore "ROOT" {CA_SERIAL}')
    else:
        if os.path.exists('/usr/bin/update-ca-trust'):
            os.system('rm /usr/share/ca-certificates/trust-source/anchors/SNI_BYPASS.crt')
            os.system('update-ca-trust extract')
        else:
            os.system('update-ca-certificates --fresh')

def onexit():
    caddy.kill()
    content = ''
    with open(host_filename,'r') as f:
        content = f.read()
    content = content[:content.find("# SNI Bypass")] + content[content.rfind("# SNI Bypass") + len("# SNI Bypass"):-1]
    with open(host_filename,'w') as f:
        f.write(content)

if __name__ == "__main__":
    config = json.loads(open("host.json","r",encoding="utf-8").read())
    with open("Caddyfile","w") as f:
        f.write('\n'.join( process_item(x) if x['enable'] else '' for x in config))
    if not os.path.exists("ca.key") or os.path.exists("ca.pem"):
        gen_CA()
        trustCA()
    CA_SERIAL = crypto.load_certificate(crypto.FILETYPE_PEM,open('ca.crt','r').read()).get_serial_number()
    gen_cert(Host)
    with open(host_filename,'a') as f:
        f.write('\n# SNI Bypass\n' + '\n'.join('127.0.0.1 %s'% x for x in Host) + '\n# SNI Bypass\n')
    caddy = subprocess.Popen(["caddy", "run", "-config", "./Caddyfile"],shell=True,stdout=sys.stdout)
    try:
        caddy.wait()
    except KeyboardInterrupt:
        onexit()
    atexit.register(onexit)