mod types;
mod block_processing;
use std::env;
use std::fs::read_to_string;
use crate::types::{Claim, ClaimVerificationContext};
use std::time::Instant;

fn main() {
    let args: Vec<String> = env::args().collect();
    if args.len() != 3 {
        eprintln!("Usage: {} <claim_path> <context_path>", args[0]);
        std::process::exit(1);
    }
    let claim_path = &args[1];
    let context_path = &args[2];

    let claim: Claim = match serde_json::from_str(&read_to_string(claim_path).expect("Failed to read claim file")) {
        Ok(c) => c,
        Err(e) => {
            eprintln!("Failed to parse claim: {}", e);
            std::process::exit(1);
        }
    };
    let context: ClaimVerificationContext = match serde_json::from_str(&read_to_string(context_path).expect("Failed to read context file")) {
        Ok(c) => c,
        Err(e) => {
            eprintln!("Failed to parse context: {}", e);
            std::process::exit(1);
        }
    };

    let start = Instant::now();
    let result = block_processing::verify(&claim, &context);
    let duration = start.elapsed();

    match result {
        Ok(_) => {
            println!("{{\"verification_time\": \"{:?}\"}}", duration);
        }
        Err(e) => {
            eprintln!("Verification failed: {:?}", e);
            std::process::exit(1);
        }
    }
}