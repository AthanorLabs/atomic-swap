use farcaster_core::crypto::dleq::DLEQProof;
use farcaster_core::consensus::{self, CanonicalBytes, Decodable, Encodable};
use std::env;
use curve25519_dalek::{
    constants::ED25519_BASEPOINT_POINT as G, edwards::CompressedEdwardsY as ed25519PointCompressed,
    edwards::EdwardsPoint as ed25519Point, scalar::Scalar as ed25519Scalar,
};
use ecdsa_fun::fun::{Point as secp256k1Point, Scalar as secp256k1Scalar, G as H};
use secp256kfun::{g, marker::*, s as sc};

fn _zeroize_highest_bits(x: [u8; 32], highest_bit: usize) -> [u8; 32] {
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

// fn main() -> Result<([u8; 32], [u8; 33]), std::io::Error> {
// fn main() -> std::io::Result<()> {
fn main() {
    let args: Vec<String> = env::args().collect();

    use rand::Rng;
    let x: [u8; 32] = rand::thread_rng().gen();
    let x_shaved = _zeroize_highest_bits(x, 252);
    let dleq = DLEQProof::generate(x_shaved);

    let bytes = dleq.as_canonical_bytes();

    use std::fs;
    use std::io::prelude::*;

    let filename = "dleq_proof.dat";
    let file = fs::File::create("dleq_proof.dat");
    file.unwrap().write_all(bytes.as_slice()).unwrap();
    // Ok(())

    let mut f = fs::File::open(&filename).expect("no file found");
    let metadata = fs::metadata(&filename).expect("unable to read metadata");
    let mut buffer = vec![0; metadata.len() as usize];

    f.read(&mut buffer).expect("buffer overflow");
    let dleq2 = DLEQProof::from_canonical_bytes(bytes.as_slice()).unwrap();

    let commitment_agg_ed25519 = dleq2.c_g.iter().sum();
    let commitment_agg_secp256k1 = dleq2
        .c_h
        .iter()
        .fold(secp256k1Point::zero(), |acc, bit_commitment| {
            g!(acc + bit_commitment).mark::<Normal>()
        });

    let verification = dleq2.verify(commitment_agg_ed25519, commitment_agg_secp256k1.mark::<NonZero>().unwrap()).unwrap();
    println!("DLEQ proof successfully verified for:\ned25519:{:?}\nsecp256k1{:?}", commitment_agg_ed25519.compress().as_bytes(), commitment_agg_secp256k1.mark::<NonZero>().unwrap().to_bytes())
    // Ok((*commitment_agg_ed25519.compress().as_bytes(), commitment_agg_secp256k1.mark::<NonZero>().unwrap().to_bytes()))
}
