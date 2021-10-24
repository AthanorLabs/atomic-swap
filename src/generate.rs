use farcaster_core::crypto::dleq::DLEQProof;
use farcaster_core::consensus::CanonicalBytes;
use std::convert::TryInto;
use std::env;
extern crate hex;

fn zeroize_highest_bits(x: [u8; 32], highest_bit: usize) -> [u8; 32] {
    let mut x = x;
    let remainder = highest_bit % 8;
    let quotient = (highest_bit - remainder) / 8;

    for bit in x.iter_mut().skip(quotient + 1) {
        *bit = 0;
    }

    if remainder != 0 {
        let mask = (2 << (remainder - 1)) - 1;
        x[quotient] &= mask;
    }

    x
}

fn main() -> std::io::Result<()> {
    let args: Vec<String> = env::args().collect();

    // use rand::Rng;
    // let x: [u8; 32] = rand::thread_rng().gen();
    // let x_shaved = zeroize_highest_bits(x, 252);
    let x: [u8; 32] = hex::decode(args.iter().nth(1).unwrap()).expect("Decoding failed").try_into().unwrap();
    // let x: [u8; 32] = bytes!(args.first().unwrap());
    let dleq = DLEQProof::generate(x);

    let bytes = dleq.as_canonical_bytes();

    use std::fs;
    use std::io::prelude::*;

    let filename = args.iter().nth(2).unwrap();
    let file = fs::File::create(filename);
    file.unwrap().write_all(bytes.as_slice()).unwrap();
    println!("successfully wrote dleq_proof to {:?}", filename);
    Ok(())
}
