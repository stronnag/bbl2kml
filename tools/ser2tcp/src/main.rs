// Serial <-> TCP Bridge
// (c) Jonathan Hudson 2023 GPL V2 or later

extern crate getopts;
use getopts::Options;
use std::env;

use std::io::{Read, Write};
use std::sync::Arc;

use std::net::TcpStream;
use serial2::SerialPort;
use std::thread;
use std::time::Duration;

const VERSION: &str = env!("CARGO_PKG_VERSION");

fn print_usage(program: &str, opts: &Options) {
    let brief = format!(
        "Usage: {} [options]\nVersion: {}",
        program, VERSION
    );
    print!("{}", opts.usage(&brief));
}

fn get_random_device() -> Option<String> {
    let res: Option<String> = match serial2::SerialPort::available_ports() {
        Ok(ports) => {
	    if ports.len() == 0 {
		None
	    } else {
		Some(ports[ports.len()-1].clone().display().to_string())
	    }
        },
        Err(_e) => {
            std::process::exit(1);
        }
    };
    res
}

fn do_main() -> Result<(), ()> {
    let args: Vec<_> = std::env::args().collect();
    let prog_name = args[0].rsplit_once('/').map(|(_parent, name)| name).unwrap_or(&args[0]);

    let mut opts = Options::new();
    opts.optflag("h", "help", "print this help menu");
    opts.optflag("V", "version", "print version and exit");
    opts.optflag("v", "verbose", "print I/O read sizes");
    opts.optopt("c", "comport", "serial device name (mandatory)", "");
    opts.optopt("b", "baudrate", "serial baud rate", "<115200>");
    opts.optopt("d", "databits", "serial databits 5|6|7|8", "<8>");
    opts.optopt("s", "stopbits", "serial stopbits [None|One|Two]", "<One>");
    opts.optopt("p", "parity", "serial parity [Even|None|Odd]", "<None>");
    opts.optopt("i", "ip", "Host name / Address", "<localhost>");
    opts.optopt("t", "tcpport", "IP port", "<5761>");
    opts.optopt("z", "buffersize", "Buffersize (ignored)", "");

    if args.len() == 1 {
        print_usage(prog_name, &opts);
        std::process::exit(1);
    }

    let matches = match opts.parse(&args[1..]) {
        Ok(m) => m,
        Err(_) => {
            print_usage(prog_name, &opts);
            return Ok(());
        }
    };

    if matches.opt_present("h") {
        print_usage(prog_name, &opts);
        return Ok(());
    }
    if matches.opt_present("V") {
	println!("{}", VERSION);
        return Ok(());
    }

    let verbose: bool = matches.opt_present("v");

    let hostname: String = match matches.opt_str("i") {
	Some(o) => o,
	None => "localhost".to_string(),
    };

    let ipport: u16 = if let Ok(Some(o)) = matches.opt_get::<u32>("t") {
	o.try_into().unwrap()
    } else {
	5761
    };

    let mut port_name: String = match matches.opt_str("c") {
	Some(o) => o,
        None => {
	    eprintln!("Serial device name is required");
            std::process::exit(1);
	},
    };

    let baud_rate: u32 = if let Ok(Some(o)) = matches.opt_get::<u32>("b") {
	o
    } else {
	115200
    };

    let parity: serial2::Parity = match matches.opt_str("s") {
	Some(o) => {
	    match o.as_str() {
		"Odd" => serial2::Parity::Odd,
		"Even" => serial2::Parity::Even,
		_ => serial2::Parity::None,
	    }},
        None => serial2::Parity::None,
    };

    let databits: serial2::CharSize = if let Ok(Some(o)) = matches.opt_get::<u8>("d") {
	match o {
	    5 => serial2::CharSize::Bits5,
	    6 => serial2::CharSize::Bits6,
	    7 => serial2::CharSize::Bits7,
	    _ => serial2::CharSize::Bits8
	}
    } else {
	serial2::CharSize::Bits8
    };

    let stopbits: serial2::StopBits = match matches.opt_str("s") {
	Some(o) => {
	    match o.as_str() {
		"Two" => serial2::StopBits::Two,
		_ => serial2::StopBits::One,
	    }},
        None => serial2::StopBits::One,
    };

    let mut port: SerialPort;
    match SerialPort::open(&port_name, baud_rate) {
	Ok(p) => port=p,
	Err(_) => {
	    match get_random_device() {
		Some(p) => {
		    port_name = p;
		    port = SerialPort::open(&port_name, baud_rate)
			.map_err(|e| eprintln!("Error: Failed to open {}: {}", port_name, e))?;
		},
		None => {
		    eprintln!("Invalid port");
		    std::process::exit(127);
		},
	    }
	}
    }

    let mut settings = port.get_configuration().unwrap();
    settings.set_char_size(databits);
    settings.set_stop_bits(stopbits);
    settings.set_parity(parity);
    port.set_configuration(&settings).unwrap();

    let port = Arc::new(port);

    let mut ostream: Option<TcpStream> = None;

    'a: for _ in 0..20  {
	match TcpStream::connect((hostname.as_str(), ipport)) {
	    Ok(t) => {
		ostream = Some(t);
		break 'a
	    },
	    Err(_) => {
		thread::sleep(Duration::from_millis(250));
	    },
	}
    }

    let mut stream: TcpStream;
    if let Some(s) = ostream {
	stream = s
    } else {
	eprintln!("Failed to connect to socket");
	std::process::exit(127);
    }
    let mut rstream = stream.try_clone().unwrap();

    print!("Connect {} to {:?}\n", port_name, stream.peer_addr().unwrap());

    std::thread::spawn({
	let port = port.clone();
	move || {
	    if let Err(()) = read_tcp_loop(port, &mut rstream, verbose) {
		std::process::exit(1);
	    }
	}
    });

    read_serial_loop(port, &mut stream, verbose)?;
    Ok(())
}

fn read_tcp_loop(port: Arc<SerialPort>, stream: &mut TcpStream, verbose: bool) -> Result<(), ()> {

    let mut buffer = [0; 512];
    loop {
	let nread = stream
	    .read(&mut buffer)
	    .map_err(|e| eprintln!("Error: Failed to read from socket: {}", e))?;
	if nread == 0 {
	    return Ok(());
	} else {
	    port.write(&buffer[..nread])
		.map_err(|e| eprintln!("Error: Failed to write to serial: {}", e))?;
	    if verbose {
		println!("TCP read: {}", nread);
	    }
	}

    }
}

fn read_serial_loop(port: Arc<SerialPort>, stream: &mut TcpStream, verbose: bool) -> Result<(), ()> {
    let mut buffer = [0; 512];
    loop {
	match port.read(&mut buffer) {
	    Ok(0) => return Ok(()),
	    Ok(n) => {
		stream
		    .write_all(&buffer[..n])
		    .map_err(|e| eprintln!("Error: Failed to write to socket: {}", e))?;
		if verbose {
		    println!("SER read: {}", n);
		}
	    },
	    Err(ref e) if e.kind() == std::io::ErrorKind::TimedOut => continue,
	    Err(e) => {
		eprintln!("Error: Failed to read from serial: {}", e);
		return Err(());
	    },
	}
    }
}

fn main() {
    if let Err(()) = do_main() {
	std::process::exit(1);
    }
}
