use ecdsa_fun::fun::Point as secp256k1Point;
use farcaster_core::consensus::CanonicalBytes;
use farcaster_core::crypto::dleq::DLEQProof;
use hex;
use secp256kfun::{g, marker::*};
use std::env;

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

    let _verification = dleq
        .verify(
            commitment_agg_ed25519,
            commitment_agg_secp256k1.mark::<NonZero>().unwrap(),
        )
        .unwrap();
    let ed25519_pub = *commitment_agg_ed25519.compress().as_bytes();
    let ed25519_pub_hex = hex::encode(ed25519_pub);
    let secp256k1_pub = commitment_agg_secp256k1
        .mark::<NonZero>()
        .unwrap()
        .to_bytes();
    let sec256k1 = commitment_agg_secp256k1
        .mark::<NonZero>()
        .unwrap();
    let (secp256k1_x, secp256k1_y) = sec256k1.coordinates();
    let secp256k1_x_hex = hex::encode(secp256k1_x);
    let secp256k1_y_hex = hex::encode(secp256k1_y);

    println!("{} {} {}", ed25519_pub_hex, secp256k1_x_hex, secp256k1_y_hex)
    // Ok((*commitment_agg_ed25519.compress().as_bytes(), commitment_agg_secp256k1.mark::<NonZero>().unwrap().to_bytes()))
}
