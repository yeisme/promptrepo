#!/usr/bin/env python3
"""Run the existing Go suite with owner-local, redacted run evidence."""
import datetime
import json
import pathlib
import re
import subprocess
import sys

root = pathlib.Path(__file__).resolve().parents[1]
started = datetime.datetime.now(datetime.timezone.utc)
run_id = 'integration-' + started.strftime('%Y%m%dT%H%M%S.%fZ')
output = root / 'temp' / 'integration-test-runs' / run_id
(output / 'artifacts').mkdir(parents=True)
command = ['go', 'test', '-race', './...']
result = subprocess.run(command, cwd=root, capture_output=True, text=True)
def redact(value):
    value = value.replace(str(root), '<project>').replace(str(pathlib.Path.home()), '<home>')
    return re.sub(r'(?im)^.*(?:authorization|bearer |SYNTHETIC_PRIVATE|SYNTHETIC_TEMPLATE_BODY).*$','<redacted>',value)
(output / 'stdout.log').write_text(redact(result.stdout))
(output / 'stderr.log').write_text(redact(result.stderr))
(output / 'command.txt').write_text(' '.join(command) + '\n')
(output / 'env.json').write_text(json.dumps({'network': 'local test fixtures', 'provider_calls': False}) + '\n')
summary = {'schema_version': 'yeisme.integration_test_evidence.v1', 'run_id': run_id, 'project': 'promptrepo', 'layer': 'integration', 'command': command, 'status': 'passed' if result.returncode == 0 else 'failed', 'exit_code': result.returncode, 'redaction': {'enabled': True}, 'started_at': started.isoformat()}
(output / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
print(redact(result.stdout), end='')
print(redact(result.stderr), end='', file=sys.stderr)
print('Evidence: ' + str(output.relative_to(root)))
sys.exit(result.returncode)
