project('flightlog2kml', version : '1.0.16')
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

if version == ''
  version = meson.project_version()
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

subdir('cmd/flightlog2kml')
subdir('cmd/fl2mqtt')
subdir('cmd/log2mission')
subdir('cmd/mission2kml')
subdir('cmd/fl2sitl')

#fl2mqtt_path = join_paths(meson.current_source_dir(), 'cmd', 'fl2mqtt')
#log2mission_path = join_paths(meson.current_source_dir(), 'cmd', 'log2mission')
#mission2kml_path = join_paths(meson.current_source_dir(), 'cmd', 'mission2kml')
#fl2sitl_path = join_paths(meson.current_source_dir(), 'cmd', 'fl2sitl')

common_files = []
subdir('pkg/geo')
subdir('pkg/types')
subdir('pkg/mission')
subdir('pkg/options')

# bbl_files
subdir('pkg/bbl')
# otx_files
subdir('pkg/otx')
# inav_files
subdir('pkg/inav')
# kml_files
subdir('pkg/kmlgen')
# blt_files
subdir('pkg/bltmqtt')
# ltm_files
subdir('pkg/ltmgen')
# bltr_files
subdir('pkg/bltreader')
# l2m_files
subdir('pkg/log2mission')
# aplog_files
subdir('pkg/aplog')
# sitl_files
subdir('pkg/sitlgen')
# inav_files
subdir('pkg/cli')
# inav_files
subdir('pkg/styles')

fl2kml_deps = [common_files, bbl_files, otx_files, inav_files, cli_files, style_files, kml_files, bltr_files, aplog_files]
fl2mqtt_deps = [common_files, bbl_files, otx_files, inav_files, blt_files, bltr_files, ltm_files, aplog_files ]
log2mission_deps = [common_files, bbl_files, otx_files, inav_files, blt_files, bltr_files, ltm_files, aplog_files ]
mission2kml_deps = [common_files, cli_files, style_files ]
fl2sitl_deps = [common_files, bbl_files, sitl_files]

flightlog2kml = custom_target(
    'flightlog2kml',
    output: 'flightlog2kml'+exe,
    env : env,
    input: [flightlog2kml_files, fl2kml_deps],
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags' , ldflags, flightlog2kml_path ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

fl2mqtt = custom_target(
    'fl2mqtt',
    output: 'fl2mqtt'+exe,
    input: [fl2mqtt_files, fl2mqtt_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, fl2mqtt_path],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

fl2sitl = custom_target(
    'fl2sitl',
    output: 'fl2sitl'+exe,
    input: [fl2sitl_files, fl2sitl_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, fl2sitl_path ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

log2mission = custom_target(
    'log2mission',
    output: 'log2mission'+exe,
    input: [ log2mission_files, log2mission_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, log2mission_path ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)

mission2kml = custom_target(
    'mission2kml',
    output: 'mission2kml'+exe,
    input: [ mission2kml_files, mission2kml_deps ],
    env : env,
    command: [ golang, 'build', trimpath, '-o', '@OUTPUT@', '-ldflags', ldflags, mission2kml_path ],
    build_by_default: true,
    install: true,
    install_dir: 'bin',
)
