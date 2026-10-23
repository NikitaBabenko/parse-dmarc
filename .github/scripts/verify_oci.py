"""Read-only OCI identity checks; public mode uses only an anonymous pull token."""
import hashlib
import json
import re
import sys
import tarfile
import urllib.error
import urllib.parse
import urllib.request


def require(condition, message):
    if not condition:
        raise ValueError(message)


def digest(data):
    return 'sha256:' + hashlib.sha256(data).hexdigest()


def verify(root, read_blob, revision, source):
    require(re.fullmatch(r'[0-9a-f]{40}', revision), 'Expected exact Git SHA')
    manifests = []

    def blob(descriptor):
        identity = descriptor['digest']
        require(re.fullmatch(r'sha256:[0-9a-f]{64}', identity), 'Invalid OCI digest')
        raw = read_blob(identity)
        require(digest(raw) == identity, 'Blob digest mismatch')
        require(len(raw) == descriptor['size'], 'Blob size mismatch')
        return json.loads(raw)

    def visit(raw, identity):
        if 'manifests' in raw:
            for descriptor in raw['manifests']:
                visit(blob(descriptor), descriptor['digest'])
        else:
            require('config' in raw and 'layers' in raw, 'Invalid image manifest')
            manifests.append((identity, raw, blob(raw['config'])))

    visit(json.loads(root), digest(root))
    apps = [(identity, manifest, config) for identity, manifest, config in manifests
            if config.get('os') == 'linux' and config.get('architecture') == 'amd64']
    require(len(apps) == 1, 'Expected exactly one linux/amd64 application')
    app_digest, app, config = apps[0]
    cfg = config['config']
    labels = cfg.get('Labels', {})
    require(labels.get('org.opencontainers.image.revision') == revision, 'Revision label mismatch')
    require(labels.get('org.opencontainers.image.source') == source, 'Source label mismatch')
    require(labels.get('org.opencontainers.image.licenses') == 'Apache-2.0', 'License label mismatch')
    require(cfg.get('Entrypoint') == ['/usr/local/bin/parse-dmarc'], 'Entrypoint mismatch')
    require('/data' in cfg.get('Volumes', {}), 'Missing /data volume')
    for setting in cfg.get('Env', []):
        key = setting.split('=', 1)[0].upper()
        require(not any(marker in key for marker in ('PASSWORD', 'TOKEN', 'SECRET', 'HTTP_AUTH')),
                'Credentials must not be embedded in image ENV')
    attestations = []
    for _, manifest, _ in manifests:
        for layer in manifest['layers']:
            if layer.get('mediaType') != 'application/vnd.in-toto+json':
                continue
            statement = blob(layer)
            if statement.get('predicateType') != 'https://slsa.dev/provenance/v0.2':
                continue
            subjects = statement.get('subject', [])
            if not any(s.get('digest', {}).get('sha256') == app_digest.split(':')[1] for s in subjects):
                continue
            predicate = statement['predicate']
            require(predicate.get('buildType') == 'https://mobyproject.org/buildkit@v1', 'Unexpected builder')
            vcs = predicate.get('metadata', {}).get('https://mobyproject.org/buildkit@v1#metadata', {}).get('vcs', {})
            require(vcs.get('revision') == revision, 'Exact clean source SHA absent from provenance')
            require(vcs.get('source', '').lower().removesuffix('.git') == source.lower(),
                    'Exact repository absent from provenance')
            attestations.append(layer['digest'])
    require(attestations, 'Missing source-bound BuildKit provenance')
    return {'indexDigest': digest(root), 'linuxAmd64ManifestDigest': app_digest,
            'configDigest': app['config']['digest'], 'revision': revision, 'source': source,
            'provenanceDigests': attestations,
            'limitation': 'BuildKit provenance is not an independent signature or runtime acceptance.'}


class SafeRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        require(urllib.parse.urlsplit(newurl).scheme == 'https', 'Non-HTTPS registry redirect')
        redirected = super().redirect_request(req, fp, code, msg, headers, newurl)
        if urllib.parse.urlsplit(req.full_url).netloc != urllib.parse.urlsplit(newurl).netloc:
            redirected.remove_header('Authorization')
        return redirected


def public(image, identity, revision, source):
    require(image == 'ghcr.io/nikitababenko/parse-dmarc', 'Unexpected publication repository')
    require(re.fullmatch(r'sha256:[0-9a-f]{64}', identity), 'Expected immutable registry digest')
    opener = urllib.request.build_opener(SafeRedirect())

    def get(url, headers=None):
        with opener.open(urllib.request.Request(url, headers=headers or {}), timeout=30) as response:
            data = response.read(4 * 1024 * 1024 + 1)
            require(len(data) <= 4 * 1024 * 1024, 'Metadata response too large')
            return data

    repository = image.split('/', 1)[1]
    query = urllib.parse.urlencode({'service': 'ghcr.io', 'scope': 'repository:' + repository + ':pull'})
    token = json.loads(get('https://ghcr.io/token?' + query))['token']
    headers = {'Authorization': 'Bearer ' + token,
               'Accept': 'application/vnd.oci.image.index.v1+json, application/vnd.oci.image.manifest.v1+json'}
    base = 'https://ghcr.io/v2/' + repository
    root = get(base + '/manifests/' + identity, headers)
    require(digest(root) == identity, 'Public manifest digest mismatch')

    def read_blob(value):
        # Manifests and config/attestation blobs are different registry resources.
        try:
            return get(base + '/blobs/' + value, headers)
        except urllib.error.HTTPError as error:
            if error.code != 404:
                raise
            return get(base + '/manifests/' + value, headers)

    return verify(root, read_blob, revision, source)


def main():
    mode, first, second, revision, source = sys.argv[1:]
    if mode == 'local':
        with open(second, 'rb') as manifest, tarfile.open(first, 'r') as archive:
            def read_blob(value):
                member = archive.getmember('blobs/sha256/' + value.split(':')[1])
                require(member.isfile() and member.size <= 4 * 1024 * 1024, 'Invalid metadata blob')
                return archive.extractfile(member).read()
            result = verify(manifest.read(), read_blob, revision, source)
    elif mode == 'public':
        result = public(first, second, revision, source)
    else:
        raise ValueError('Mode must be local or public')
    print(json.dumps(result, indent=2))


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        # Never print requests/headers or token-bearing URLs.
        print('OCI verification failed: ' + type(error).__name__ + ': ' + str(error), file=sys.stderr)
        sys.exit(1)
