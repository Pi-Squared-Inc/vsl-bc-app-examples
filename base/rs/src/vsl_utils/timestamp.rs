//! # Timestamp Module
//!
//! This module provides the `Timestamp` struct, a compact representation of time with
//! second and nanosecond precision. It supports natural ordering (`Ord`, `PartialOrd`),
//! formatted display, string parsing, and JSON serialization.
//!
use alloy_rlp::{RlpDecodable, RlpEncodable};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};
use std::cmp::Ordering;
use std::fmt;
use std::time::{SystemTime, UNIX_EPOCH};

#[derive(
    Debug,
    Clone,
    Copy,
    PartialEq,
    Eq,
    Serialize,
    Deserialize,
    JsonSchema,
    RlpEncodable,
    RlpDecodable,
)]
pub struct Timestamp {
    seconds: u64,
    nanos: u32,
}

impl Timestamp {
    pub fn now() -> Self {
        let duration = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .expect("Time went backwards!");
        Timestamp {
            seconds: duration.as_secs(),
            nanos: duration.subsec_nanos(),
        }
    }

    pub fn seconds(&self) -> u64 {
        self.seconds
    }

    pub fn nanos(&self) -> u32 {
        self.nanos
    }

    pub fn from_seconds(seconds: u64) -> Self {
        Timestamp { seconds, nanos: 0 }
    }

    /// The timestamp as a formatted string: "seconds.nanos"
    pub fn to_string(&self) -> String {
        format!("{}.{}", self.seconds, self.nanos)
    }

    /// Returns the next timestamp, incrementing the nanos by one.
    /// If the nanos are already at their maximum value, the seconds
    /// are incremented and the nanos are reset to zero.
    pub fn tick(&self) -> Self {
        if self.nanos == 999_999_999 {
            Timestamp {
                seconds: self.seconds + 1,
                nanos: 0,
            }
        } else {
            Timestamp {
                seconds: self.seconds,
                nanos: self.nanos + 1,
            }
        }
    }
}

use std::str::FromStr;

#[derive(Debug, PartialEq, Eq)]
pub struct ParseTimestampError;

impl FromStr for Timestamp {
    type Err = ParseTimestampError;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        let parts: Vec<&str> = s.split('.').collect();
        if parts.len() != 2 {
            return Err(ParseTimestampError);
        }

        let seconds = u64::from_str(parts[0]).map_err(|_| ParseTimestampError)?;
        let nanos = u32::from_str(parts[1]).map_err(|_| ParseTimestampError)?;

        if nanos >= 1_000_000_000 {
            return Err(ParseTimestampError);
        }

        Ok(Timestamp { seconds, nanos })
    }
}

/// Allow printing the timestamp directly
impl fmt::Display for Timestamp {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        write!(f, "{}.{}", self.seconds, self.nanos)
    }
}

/// Allow comparing timestamps
impl PartialOrd for Timestamp {
    fn partial_cmp(&self, other: &Self) -> Option<Ordering> {
        Some(self.cmp(other))
    }
}

impl Ord for Timestamp {
    fn cmp(&self, other: &Self) -> Ordering {
        match self.seconds.cmp(&other.seconds) {
            Ordering::Equal => self.nanos.cmp(&other.nanos),
            other => other,
        }
    }
}
