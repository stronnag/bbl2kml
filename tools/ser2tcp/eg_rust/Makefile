
build:
	cargo build --release

linux-static:
	cargo build --target=x86_64-unknown-linux-musl

install:
	cargo install --path . --root $$HOME/.local/ --force

debug:
	cargo build

windows:
	cargo build --target x86_64-pc-windows-gnu --release

clean:
	cargo clean
