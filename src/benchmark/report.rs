use std::fmt;
use std::time::Duration;

use super::error::BenchmarkError;

/// The storage backend used for one benchmark sample.
#[derive(Clone, Copy, Debug, Eq, Ord, PartialEq, PartialOrd)]
pub enum Backend {
    Packed,
    Sqlite,
}

impl fmt::Display for Backend {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str(match self {
            Self::Packed => "packed",
            Self::Sqlite => "sqlite",
        })
    }
}

/// The operation measured by a benchmark sample.
#[derive(Clone, Copy, Debug, Eq, Ord, PartialEq, PartialOrd)]
pub enum BenchmarkMetric {
    Build,
    Reopen,
    Query,
    Mutation,
    DiskSize,
}

impl fmt::Display for BenchmarkMetric {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str(match self {
            Self::Build => "build",
            Self::Reopen => "reopen",
            Self::Query => "query",
            Self::Mutation => "mutation",
            Self::DiskSize => "disk_size",
        })
    }
}

/// One measured benchmark observation.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct BenchmarkSample {
    pub graph: String,
    pub backend: Backend,
    pub metric: BenchmarkMetric,
    pub workload: String,
    pub sample: u32,
    pub duration: Duration,
    pub operations: u64,
    pub items: u64,
    pub file_size: u64,
    pub fingerprint: u64,
}

impl BenchmarkSample {
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        graph: impl Into<String>,
        backend: Backend,
        metric: BenchmarkMetric,
        workload: impl Into<String>,
        sample: u32,
        duration: Duration,
        operations: u64,
        items: u64,
        file_size: u64,
        fingerprint: u64,
    ) -> Self {
        Self {
            graph: graph.into(),
            backend,
            metric,
            workload: workload.into(),
            sample,
            duration,
            operations,
            items,
            file_size,
            fingerprint,
        }
    }
}

/// Collected benchmark observations and their serializable summaries.
#[derive(Clone, Debug, Default, Eq, PartialEq)]
pub struct BenchmarkReport {
    samples: Vec<BenchmarkSample>,
}

impl BenchmarkReport {
    pub const fn new() -> Self {
        Self {
            samples: Vec::new(),
        }
    }

    pub fn push(&mut self, sample: BenchmarkSample) {
        self.samples.push(sample);
    }

    pub fn samples(&self) -> &[BenchmarkSample] {
        &self.samples
    }

    pub fn to_csv(&self) -> String {
        let mut csv = String::from(
            "graph,backend,metric,workload,sample,duration_ns,operations,items,file_size,fingerprint\n",
        );
        for sample in &self.samples {
            append_csv_field(&mut csv, &sample.graph);
            csv.push(',');
            append_csv_field(&mut csv, &sample.backend.to_string());
            csv.push(',');
            append_csv_field(&mut csv, &sample.metric.to_string());
            csv.push(',');
            append_csv_field(&mut csv, &sample.workload);
            csv.push_str(&format!(
                ",{},{},{},{},{},{}\n",
                sample.sample,
                sample.duration.as_nanos(),
                sample.operations,
                sample.items,
                sample.file_size,
                sample.fingerprint
            ));
        }
        csv
    }

    pub fn write_csv<W: std::io::Write>(&self, mut writer: W) -> Result<(), BenchmarkError> {
        writer.write_all(self.to_csv().as_bytes())?;
        Ok(())
    }

    pub fn human_summary(&self) -> String {
        let mut groups = self
            .samples
            .iter()
            .map(|sample| (sample.graph.clone(), sample.workload.clone(), sample.metric))
            .collect::<Vec<_>>();
        groups.sort_unstable();
        groups.dedup();

        let mut summary = String::new();
        for (graph, workload, metric) in groups {
            let samples = self.samples.iter().filter(|sample| {
                sample.graph == graph && sample.workload == workload && sample.metric == metric
            });
            let mut packed = Vec::new();
            let mut sqlite = Vec::new();
            for sample in samples {
                match sample.backend {
                    Backend::Packed => packed.push(sample),
                    Backend::Sqlite => sqlite.push(sample),
                }
            }

            summary.push_str(&format!("{graph}/{workload}/{metric}:"));
            append_median_timing(&mut summary, "packed", &packed);
            append_median_timing(&mut summary, "sqlite", &sqlite);
            append_speedup(&mut summary, &packed, &sqlite);
            append_file_sizes(&mut summary, &packed, &sqlite);
            if metric == BenchmarkMetric::Query {
                append_throughput(&mut summary, "packed", &packed);
                append_throughput(&mut summary, "sqlite", &sqlite);
            }
            summary.push('\n');
        }
        summary
    }
}

fn append_csv_field(csv: &mut String, value: &str) {
    if value
        .bytes()
        .any(|byte| matches!(byte, b',' | b'"' | b'\n' | b'\r'))
    {
        csv.push('"');
        for character in value.chars() {
            if character == '"' {
                csv.push('"');
            }
            csv.push(character);
        }
        csv.push('"');
    } else {
        csv.push_str(value);
    }
}

fn median(values: &mut [u128]) -> Option<u128> {
    if values.is_empty() {
        return None;
    }
    values.sort_unstable();
    let middle = values.len() / 2;
    if values.len() % 2 == 1 {
        Some(values[middle])
    } else {
        Some(values[middle - 1] + (values[middle] - values[middle - 1]) / 2)
    }
}

fn append_median_timing(summary: &mut String, backend: &str, samples: &[&BenchmarkSample]) {
    let mut durations = samples
        .iter()
        .filter_map(|sample| {
            let duration = sample.duration.as_nanos();
            (duration > 0).then_some(duration)
        })
        .collect::<Vec<_>>();
    if let Some(duration) = median(&mut durations) {
        summary.push_str(&format!(" {backend} median {}", format_duration(duration)));
    }
}

fn append_speedup(summary: &mut String, packed: &[&BenchmarkSample], sqlite: &[&BenchmarkSample]) {
    let mut packed_durations = packed
        .iter()
        .filter_map(|sample| {
            let duration = sample.duration.as_nanos();
            (duration > 0).then_some(duration)
        })
        .collect::<Vec<_>>();
    let mut sqlite_durations = sqlite
        .iter()
        .filter_map(|sample| {
            let duration = sample.duration.as_nanos();
            (duration > 0).then_some(duration)
        })
        .collect::<Vec<_>>();
    if let (Some(packed), Some(sqlite)) =
        (median(&mut packed_durations), median(&mut sqlite_durations))
        && packed > 0
    {
        summary.push_str(&format!(" speedup {:.2}x", sqlite as f64 / packed as f64));
    }
}

fn append_file_sizes(
    summary: &mut String,
    packed: &[&BenchmarkSample],
    sqlite: &[&BenchmarkSample],
) {
    let mut packed_sizes = packed
        .iter()
        .map(|sample| u128::from(sample.file_size))
        .collect::<Vec<_>>();
    let mut sqlite_sizes = sqlite
        .iter()
        .map(|sample| u128::from(sample.file_size))
        .collect::<Vec<_>>();
    let packed_size = median(&mut packed_sizes);
    let sqlite_size = median(&mut sqlite_sizes);
    if packed_size.is_some() || sqlite_size.is_some() {
        summary.push_str(&format!(
            " file_size packed={}B sqlite={}B",
            packed_size.map_or_else(|| "-".to_owned(), |size| size.to_string()),
            sqlite_size.map_or_else(|| "-".to_owned(), |size| size.to_string())
        ));
    }
}

fn append_throughput(summary: &mut String, backend: &str, samples: &[&BenchmarkSample]) {
    let mut throughputs = samples
        .iter()
        .filter_map(|sample| {
            let nanos = sample.duration.as_nanos();
            (nanos > 0).then_some(sample.operations as f64 * 1_000_000_000.0 / nanos as f64)
        })
        .collect::<Vec<_>>();
    if throughputs.is_empty() {
        return;
    }
    throughputs.sort_by(f64::total_cmp);
    let middle = throughputs.len() / 2;
    let throughput = if throughputs.len() % 2 == 1 {
        throughputs[middle]
    } else {
        (throughputs[middle - 1] + throughputs[middle]) / 2.0
    };
    summary.push_str(&format!(" {backend} throughput={throughput:.1} ops/s"));
}

fn format_duration(nanos: u128) -> String {
    if nanos < 1_000 {
        format!("{nanos}ns")
    } else if nanos < 1_000_000 {
        format!("{:.3}us", nanos as f64 / 1_000.0)
    } else if nanos < 1_000_000_000 {
        format!("{:.3}ms", nanos as f64 / 1_000_000.0)
    } else {
        format!("{:.3}s", nanos as f64 / 1_000_000_000.0)
    }
}
