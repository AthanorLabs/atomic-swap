use farcaster_core::crypto::dleq::DLEQProof;
use farcaster_core::consensus::CanonicalBytes;
use std::env;
use ecdsa_fun::fun::Point as secp256k1Point;
use secp256kfun::{g, marker::*};

fn main() {
    let args: Vec<String> = env::args().collect();

    use std::fs;
    use std::io::prelude::*;

    let filename = args.iter().nth(1).unwrap();
    let mut f = fs::File::open(&filename).expect("no file found");
    let metadata = fs::metadata(&filename).expect("unable to read metadata");
    let mut buffer = vec![0; metadata.len() as usize];

    f.read(&mut buffer).expect("buffer overflow");
    let dleq = DLEQProof::from_canonical_bytes(buffer.as_slice()).unwrap();

    let commitment_agg_ed25519 = dleq.c_g.iter().sum();
    let commitment_agg_secp256k1 = dleq
        .c_h
        .iter()
        .fold(secp256k1Point::zero(), |acc, bit_commitment| {
            g!(acc + bit_commitment).mark::<Normal>()
        });

    let verification = dleq.verify(commitment_agg_ed25519, commitment_agg_secp256k1.mark::<NonZero>().unwrap()).unwrap();
    println!("DLEQ proof successfully verified for:\ned25519:{:?}\nsecp256k1{:?}", commitment_agg_ed25519.compress().as_bytes(), commitment_agg_secp256k1.mark::<NonZero>().unwrap().to_bytes())
    // Ok((*commitment_agg_ed25519.compress().as_bytes(), commitment_agg_secp256k1.mark::<NonZero>().unwrap().to_bytes()))
}
