use crate::synthetic::{GraphSpec, ScaleTier, Topology};
use std::fmt;
use std::path::PathBuf;

const DEFAULT_QUERY_COUNT: usize = 1_000;
const DEFAULT_SAMPLE_COUNT: u32 = 3;
const DEFAULT_SEED: u64 = 0;
const DEFAULT_WORK_DIR: &str = "target/arcana-benchmark";

/// Parsed options for the benchmark command.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct BenchmarkCommand {
    pub graph: GraphSpec,
    pub query_count: usize,
    pub sample_count: u32,
    pub csv_path: Option<PathBuf>,
    pub work_dir: PathBuf,
    pub keep_files: bool,
}

impl Default for BenchmarkCommand {
    fn default() -> Self {
        Self {
            graph: topology_preset(ScaleTier::Small, "modular", DEFAULT_SEED)
                .expect("the default topology preset is valid"),
            query_count: DEFAULT_QUERY_COUNT,
            sample_count: DEFAULT_SAMPLE_COUNT,
            csv_path: None,
            work_dir: PathBuf::from(DEFAULT_WORK_DIR),
            keep_files: false,
        }
    }
}
impl BenchmarkCommand {
    /// Parses benchmark options, excluding the executable name.
    pub fn parse<I, S>(arguments: I) -> Result<Self, BenchmarkParseError>
    where
        I: IntoIterator<Item = S>,
        S: AsRef<str>,
    {
        let mut command = Self::default();
        let mut arguments = arguments
            .into_iter()
            .map(|argument| argument.as_ref().to_owned())
            .peekable();
        if arguments
            .peek()
            .is_some_and(|argument| argument == "benchmark")
        {
            arguments.next();
        }

        let mut tier = ScaleTier::Small;
        let mut topology = "modular".to_owned();
        let mut seed = DEFAULT_SEED;
        while let Some(argument) = arguments.next() {
            let (option, inline_value) = split_option(&argument);
            match option {
                "--tier" => tier = parse_tier(&value(&mut arguments, "--tier", inline_value)?)?,
                "--topology" => {
                    topology = value(&mut arguments, "--topology", inline_value)?;
                }
                "--queries" => {
                    command.query_count = parse_number(
                        "--queries",
                        &value(&mut arguments, "--queries", inline_value)?,
                    )?;
                }
                "--samples" => {
                    command.sample_count = parse_number(
                        "--samples",
                        &value(&mut arguments, "--samples", inline_value)?,
                    )?;
                }
                "--seed" => {
                    seed = parse_number("--seed", &value(&mut arguments, "--seed", inline_value)?)?;
                }
                "--csv" => {
                    command.csv_path =
                        Some(PathBuf::from(value(&mut arguments, "--csv", inline_value)?));
                }
                "--work-dir" => {
                    command.work_dir =
                        PathBuf::from(value(&mut arguments, "--work-dir", inline_value)?);
                }
                "--keep-files" if inline_value.is_none() => command.keep_files = true,
                option if option.starts_with('-') => {
                    return Err(BenchmarkParseError::UnknownFlag(argument));
                }
                _ => return Err(BenchmarkParseError::UnexpectedArgument(argument)),
            }
        }

        command.graph = topology_preset(tier, &topology, seed)?;
        Ok(command)
    }
}
/// Builds a valid graph specification for a named topology and scale tier.
pub fn topology_preset(
    tier: ScaleTier,
    topology: &str,
    seed: u64,
) -> Result<GraphSpec, BenchmarkParseError> {
    let (cluster_count, hub_count, layer_count, dense_node_count) = match tier {
        ScaleTier::Small => (8, 4, 8, 100),
        ScaleTier::Medium => (32, 16, 16, 1_000),
        ScaleTier::Large => (128, 64, 32, 10_000),
        ScaleTier::Stress => (512, 256, 64, 50_000),
    };
    let topology = match topology {
        "modular" => Topology::Modular {
            cluster_count,
            cross_cluster_ratio: 2_500,
        },
        "entangled" => Topology::Entangled {
            cluster_count,
            hub_count,
        },
        "hub-heavy" => Topology::HubHeavy { hub_count },
        "layered" => Topology::Layered { layer_count },
        "dense-subsystem" => Topology::DenseSubsystem { dense_node_count },
        name => return Err(BenchmarkParseError::UnsupportedTopology(name.to_owned())),
    };
    let spec = GraphSpec::for_tier(topology, tier, seed);
    spec.validate()
        .map_err(|error| BenchmarkParseError::InvalidPreset {
            topology: topology_name(topology),
            error: error.to_string(),
        })?;
    Ok(spec)
}

/// Returns the benchmark command help text.
pub fn benchmark_usage() -> &'static str {
    "Usage: arcana benchmark [OPTIONS]\n\nOptions:\n    --tier <small|medium|large|stress>\n    --topology <modular|entangled|hub-heavy|layered|dense-subsystem>\n    --queries <COUNT>\n    --samples <COUNT>\n    --seed <NUMBER>\n    --csv <PATH>\n    --work-dir <PATH>\n    --keep-files"
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum BenchmarkParseError {
    MissingValue {
        option: &'static str,
    },
    UnknownFlag(String),
    UnexpectedArgument(String),
    InvalidNumber {
        option: &'static str,
        value: String,
    },
    UnsupportedTier(String),
    UnsupportedTopology(String),
    InvalidPreset {
        topology: &'static str,
        error: String,
    },
}

impl fmt::Display for BenchmarkParseError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::MissingValue { option } => write!(formatter, "missing value for {option}"),
            Self::UnknownFlag(flag) => write!(formatter, "unknown flag '{flag}'"),
            Self::UnexpectedArgument(argument) => {
                write!(formatter, "unexpected argument '{argument}'")
            }
            Self::InvalidNumber { option, value } => {
                write!(formatter, "invalid number '{value}' for {option}")
            }
            Self::UnsupportedTier(tier) => write!(formatter, "unsupported tier '{tier}'"),
            Self::UnsupportedTopology(topology) => {
                write!(formatter, "unsupported topology '{topology}'")
            }
            Self::InvalidPreset { topology, error } => {
                write!(formatter, "invalid {topology} topology preset: {error}")
            }
        }
    }
}

impl std::error::Error for BenchmarkParseError {}

fn split_option(argument: &str) -> (&str, Option<&str>) {
    argument
        .split_once('=')
        .map_or((argument, None), |(option, value)| (option, Some(value)))
}

fn value<I>(
    arguments: &mut I,
    option: &'static str,
    inline_value: Option<&str>,
) -> Result<String, BenchmarkParseError>
where
    I: Iterator<Item = String>,
{
    if let Some(value) = inline_value {
        if !value.is_empty() {
            return Ok(value.to_owned());
        }
        return Err(BenchmarkParseError::MissingValue { option });
    }
    match arguments.next() {
        Some(value) if !value.starts_with("--") => Ok(value),
        _ => Err(BenchmarkParseError::MissingValue { option }),
    }
}

fn parse_number<T: std::str::FromStr>(
    option: &'static str,
    value: &str,
) -> Result<T, BenchmarkParseError> {
    value
        .parse()
        .map_err(|_| BenchmarkParseError::InvalidNumber {
            option,
            value: value.to_owned(),
        })
}

fn parse_tier(value: &str) -> Result<ScaleTier, BenchmarkParseError> {
    match value {
        "small" => Ok(ScaleTier::Small),
        "medium" => Ok(ScaleTier::Medium),
        "large" => Ok(ScaleTier::Large),
        "stress" => Ok(ScaleTier::Stress),
        tier => Err(BenchmarkParseError::UnsupportedTier(tier.to_owned())),
    }
}
fn topology_name(topology: Topology) -> &'static str {
    match topology {
        Topology::Modular { .. } => "modular",
        Topology::Entangled { .. } => "entangled",
        Topology::HubHeavy { .. } => "hub-heavy",
        Topology::Layered { .. } => "layered",
        Topology::DenseSubsystem { .. } => "dense-subsystem",
    }
}
