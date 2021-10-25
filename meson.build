project('fl2x',  version : '0.0.0')
meson.add_install_script('meson/post_install.sh')
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

# This is for old distros like Debain 10, in which case the '-v' is harmles
r = run_command('go', 'env', 'GOVERSION')
res = r.stdout().strip()
trimpath = '-trimpath'
if res == ''
  trimpath = '-v'
endif

flightlog2kml_path = join_paths(meson.current_source_dir(), 'cmd', 'flightlog2kml')
fl2mqtt_path = join_paths(meson.current_source_dir(), 'cmd', 'fl2mqtt')
log2mission_path = join_paths(meson.current_source_dir(), 'cmd', 'log2mission')
mission2kml_path = join_paths(meson.current_source_dir(), 'cmd', 'mission2kml')

r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'api','types'))
common_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'mission'))
common_files += r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'geo'))
common_files += r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'options'))
common_files += r.stdout().strip().split()

r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'bbl'))
bbl_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'otx'))
otx_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'inav'))
inav_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'kmlgen'))
kml_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'bltmqtt'))
blt_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'ltmgen'))
ltm_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'bltreader'))
bltr_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'log2mission'))
l2m_files = r.stdout().strip().split()
r = run_command('meson/find_files.sh', join_paths(meson.current_source_dir(), 'pkg', 'aplog'))
aplog_files = r.stdout().strip().split()

fl2kml_deps = [common_files, bbl_files, otx_files, inav_files, kml_files, bltr_files, aplog_files]

fl2mqtt_deps = [common_files, bbl_files, otx_files, inav_files, blt_files, bltr_files, ltm_files, aplog_files ]
log2mission_deps = [common_files, bbl_files, otx_files, inav_files, blt_files, bltr_files, ltm_files, aplog_files ]
mission2kml_deps = [common_files]

r = run_command('meson/find_files.sh', flightlog2kml_path)
flightlog2kml_files = r.stdout().strip().split()

r = run_command('meson/find_files.sh', fl2mqtt_path)
fl2mqtt_files =  r.stdout().strip().split()

r = run_command('meson/find_files.sh', log2mission_path)
log2mission_files =  r.stdout().strip().split()

r = run_command('meson/find_files.sh', mission2kml_path)
mission2kml_files =  r.stdout().strip().split()

flightlog2kml_build_path = join_paths(meson.current_source_dir(), 'cmd' , 'flightlog2kml')
flightlog2kml = custom_target(
    'flightlog2kml',
    output: 'flightlog2kml',
    env : env,
    input: [flightlog2kml_files, fl2kml_deps],
    command: [ golang, 'build', trimpath, '-o', 'flightlog2kml', '-ldflags' , ldflags, flightlog2kml_files ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

fl2mqtt_build_path = join_paths(meson.current_source_dir(), 'cmd' , 'fl2mqtt')
fl2mqtt = custom_target(
    'fl2mqtt',
    output: 'fl2mqtt',
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
    output: 'log2mission',
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
    output: 'mission2kml',
    input: [ mission2kml_files, mission2kml_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, mission2kml_files ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)
