project('fl2x',  version : '0.0.0')
meson.add_install_script('meson/post_install.py')
version = get_option('version')
commit = get_option('commit')

golang = find_program('go')
env = environment()

env.set('CGO_ENABLED', '0')

ldflags =  '-extldflags -static -s -w'
if commit != ''
  ldflags += ' -X main.GitCommit=' + commit
endif

if version != ''
  ldflags += ' -X main.GitTag='+ version
endif

# This is beyond stupid ....
pymodule = import('python3')
python3 = pymodule.find_python()
exe=''
res = run_command(python3, '-c', 'import os; print(os.environ["EXE"])', check: false)
if res.returncode() == 0
  exe = res.stdout().strip()
endif

# This is for old distros like Debain 10, in which case the '-v' is harmles
r = run_command('go', 'env', 'GOVERSION', check: false)
res = r.stdout().strip()
trimpath = '-trimpath'
if res == ''
  trimpath = '-v'
endif

flightlog2kml_path = join_paths(meson.current_source_dir(), 'cmd', 'flightlog2kml')
fl2mqtt_path = join_paths(meson.current_source_dir(), 'cmd', 'fl2mqtt')
log2mission_path = join_paths(meson.current_source_dir(), 'cmd', 'log2mission')
mission2kml_path = join_paths(meson.current_source_dir(), 'cmd', 'mission2kml')

r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'api','types'), check: true)
common_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'mission'), check: true)
common_files += r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'geo'), check: true)
common_files += r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'options'), check: true)
common_files += r.stdout().strip().split()

r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'bbl'), check: true)
bbl_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'otx'), check: true)
otx_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'inav'), check: true)
inav_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'kmlgen'), check: true)
kml_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'bltmqtt'), check: true)
blt_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'ltmgen'), check: true)
ltm_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'bltreader'), check: true)
bltr_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'log2mission'), check: true)
l2m_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'aplog'), check: true)
aplog_files = r.stdout().strip().split()

fl2kml_deps = [common_files, bbl_files, otx_files, inav_files, kml_files, bltr_files, aplog_files]

fl2mqtt_deps = [common_files, bbl_files, otx_files, inav_files, blt_files, bltr_files, ltm_files, aplog_files ]
log2mission_deps = [common_files, bbl_files, otx_files, inav_files, blt_files, bltr_files, ltm_files, aplog_files ]
mission2kml_deps = [common_files]

r = run_command('meson/find_files.sh', flightlog2kml_path, check: true)
flightlog2kml_files = r.stdout().strip().split()

r = run_command('meson/find_files.sh', fl2mqtt_path, check: true)
fl2mqtt_files =  r.stdout().strip().split()

r = run_command('meson/find_files.sh', log2mission_path, check: true)
log2mission_files =  r.stdout().strip().split()

r = run_command('meson/find_files.sh', mission2kml_path, check: true)
mission2kml_files =  r.stdout().strip().split()

flightlog2kml_build_path = join_paths(meson.current_source_dir(), 'cmd' , 'flightlog2kml')
flightlog2kml = custom_target(
    'flightlog2kml',
    output: 'flightlog2kml'+exe,
    env : env,
    input: [flightlog2kml_files, fl2kml_deps],
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags' , ldflags, flightlog2kml_files ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

fl2mqtt_build_path = join_paths(meson.current_source_dir(), 'cmd' , 'fl2mqtt')
fl2mqtt = custom_target(
    'fl2mqtt',
    output: 'fl2mqtt'+exe,
    input: [fl2mqtt_files, fl2mqtt_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, fl2mqtt_files ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

log2mission_build_path = join_paths(meson.current_source_dir(), 'cmd' , 'log2mission')
log2mission = custom_target(
    'log2mission',
    output: 'log2mission'+exe,
    input: [ log2mission_files, log2mission_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, log2mission_files ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

mission2kml_build_path = join_paths(meson.current_source_dir(), 'cmd' , 'mission2kml')
mission2kml = custom_target(
    'mission2kml',
    output: 'mission2kml'+exe,
    input: [ mission2kml_files, mission2kml_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, mission2kml_files ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)
