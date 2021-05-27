import os
from OpenSSL import crypto
import random
import pickle


class CertificateManager:
    issuer = {
        "C": "CN",
        "ST": "CN",
        "L": "CN",
        "O": "SNI Bypass",
        "CN": "SNI Bypass"
    }
    CA_KEY_STRING=""
    CA_CRT_STRING=""
    CA_SERIAL=""
    CA_CRT_FILENAME="SNI_BYPASS.crt"
    def __init__(self) -> None:
        pass

    def load(self,filename):
        with open(filename,'rb') as f:
            self.__dict__ = pickle.loads(f.read())
    
    def save(self,filename):
        with open(filename,'wb') as f:
            f.write(pickle.dumps(self.__dict__))
    
    def gen_CA(self):
        CA_KEY = crypto.PKey()
        CA_KEY.generate_key(crypto.TYPE_RSA, 2048)

        CA_CRT = crypto.X509()
        CA_CRT.set_version(2)
        CA_CRT.gmtime_adj_notBefore(0)
        CA_CRT.gmtime_adj_notAfter(10 * 365 * 24 * 60 * 60)
        CA_CRT.set_serial_number(random.randint(50000000, 100000000))

        CA_CRT.get_subject().commonName = "SNI Bypass Root CA"
        CA_CRT.get_subject().C = self.issuer['C']
        CA_CRT.get_subject().ST = self.issuer["ST"]
        CA_CRT.get_subject().L = self.issuer["L"]
        CA_CRT.get_subject().O = self.issuer["O"]
        CA_CRT.get_subject().CN = self.issuer["CN"]

        CA_CRT.set_issuer(CA_CRT.get_subject())
        CA_CRT.set_pubkey(CA_KEY)

        CA_CRT.add_extensions([
            crypto.X509Extension(b"subjectKeyIdentifier",
                                 False, b"hash", subject=CA_CRT),
        ])

        CA_CRT.add_extensions([
            crypto.X509Extension(b"authorityKeyIdentifier",
                                 False, b"keyid:always,issuer", issuer=CA_CRT),
        ])

        CA_CRT.add_extensions([
            crypto.X509Extension(b"basicConstraints", True, b"CA:TRUE"),
            crypto.X509Extension(
                b"keyUsage", True, b"digitalSignature, keyCertSign, cRLSign"),
        ])

        CA_CRT.sign(CA_KEY, 'sha256')
        self.CA_CRT_STRING = crypto.dump_certificate(crypto.FILETYPE_PEM,CA_CRT)
        self.CA_KEY_STRING = crypto.dump_privatekey(crypto.FILETYPE_PEM,CA_KEY)

        self.CA_SERIAL = CA_CRT.get_serial_number()

    def gen_cert(self,hosts):
        CA_KEY = crypto.load_privatekey(crypto.FILETYPE_PEM,self.CA_KEY_STRING)
        CA_CRT =crypto.load_certificate(crypto.FILETYPE_PEM,self.CA_CRT_STRING)

        cert = crypto.X509()
        key = crypto.PKey()
        key.generate_key(crypto.TYPE_RSA, 2048)
        cert.gmtime_adj_notBefore(0)
        cert.gmtime_adj_notAfter(10 * 365 * 24 * 60 * 60)
        cert.set_serial_number(random.randint(50000000, 100000000))

        cert.set_version(2)
        cert.set_pubkey(key)
        cert.set_subject(CA_CRT.get_subject())
        cert.set_issuer(CA_CRT.get_issuer())

        cert.add_extensions([
            crypto.X509Extension(b'subjectAltName', False, ','.join(
                'DNS:%s' % x for x in hosts).encode()),
            crypto.X509Extension(b"basicConstraints", True, b"CA:FALSE"),
            crypto.X509Extension(
                b"keyUsage", True, b"Digital Signature, Non Repudiation, Key Encipherment, Data Encipherment"),
            crypto.X509Extension(b"extendedKeyUsage", True,
                                 b"TLS Web Server Authentication, TLS Web Client Authentication, Code Signing, E-mail Protection"),
            crypto.X509Extension(b"authorityKeyIdentifier",
                                 False, b"keyid", issuer=CA_CRT),
            crypto.X509Extension(b"subjectKeyIdentifier",
                                 False, b"hash", subject=cert),
        ])

        cert.sign(CA_KEY, 'sha256')
        open("SSL.crt", 'wb').write(
            crypto.dump_certificate(crypto.FILETYPE_PEM, cert))
        open("SSL.key", 'wb').write(
            crypto.dump_privatekey(crypto.FILETYPE_PEM, key))

    def trustCA(self):
        with open(self.CA_CRT_FILENAME,'wb') as f:
            f.write(self.CA_CRT_STRING)
        if os.name == 'nt':
            os.system(f'certutil.exe -addstore "ROOT" {self.CA_CRT_FILENAME}')
        else:
            os.system(
                f'cp {self.CA_CRT_FILENAME} /usr/share/ca-certificates/trust-source/anchors/{self.CA_CRT_FILENAME}')
            if os.path.exists('/usr/bin/update-ca-trust'):
                os.system('update-ca-trust')
            else:
                os.system('update-ca-certificates')
        os.remove(self.CA_CRT_FILENAME)

    def delCA(self):
        if os.name == 'nt':
            os.system(f'certutil -delstore "ROOT" {CA_SERIAL}')
        else:
            if os.path.exists('/usr/bin/update-ca-trust'):
                os.system(
                    f'rm /usr/share/ca-certificates/trust-source/anchors/{self.CA_CRT_FILENAME}')
                os.system('update-ca-trust extract')
            else:
                os.system('update-ca-certificates --fresh')

if __name__ == '__main__':
    cm = CertificateManager()
    cm.load("CertificateManager.bin")
    cm.delCA()

