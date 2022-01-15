use farcaster_core::consensus::CanonicalBytes;
use farcaster_core::crypto::dleq::DLEQProof;
use rand_core::{OsRng, RngCore};
use std::env;
use std::fs;
use std::io::prelude::*;

fn main() -> Result<(), std::io::Error> {
    let args: Vec<String> = env::args().collect();

    let mut x = [0u8; 32];
    OsRng.fill_bytes(&mut x);
    x[31] = x[31] & 0b00001111; // zero highest four bits

    let dleq = DLEQProof::generate(x);
    let bytes = dleq.as_canonical_bytes();

    let filename = args.iter().nth(1).unwrap();
    let mut file = fs::File::create(filename)?;
    file.write_all(bytes.as_slice())?;
    let mut file = fs::File::create(filename.to_owned() + ".key")?;
    file.write_all(&x)?;
    println!(
        "successfully wrote dleq_proof to {:?} and key to {:?}.key",
        filename, filename
    );
    Ok(())
}
