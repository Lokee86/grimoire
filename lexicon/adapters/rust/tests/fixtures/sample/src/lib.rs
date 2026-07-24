pub mod child;

use crate::child::{helper as child_helper, Worker as ImportedWorker};
use crate::child::*;
use std::fmt::Debug;

pub trait Runnable {
    fn run(&self);

    fn defaulted(&self) {
        helper();
    }
}

pub struct Service {
    worker: Worker,
}

impl Service {
    pub fn new(worker: Worker) -> Self {
        Self { worker }
    }

    pub fn factory() -> Self {
        Self::new(Worker::new())
    }

    pub fn run_local(&self) {
        self.worker.work();
        helper();
    }
}

impl Runnable for Service {
    fn run(&self) {
        self.run_local();
    }
}

pub struct Alternate;

impl Runnable for Alternate {
    fn run(&self) {
        helper();
    }
}

pub enum Kind {
    Ready,
    Data(Service),
}

pub const FLOW_CONST: i32 = 1;

pub struct FlowBox {
    pub field: i32,
}

impl FlowBox {
    pub fn update(&mut self, value: i32) -> i32 {
        self.field = value;
        self.field += FLOW_CONST;
        self.field
    }
}

pub fn flow(value: i32, mut box_value: FlowBox) -> i32 {
    let mut local = value;
    local += FLOW_CONST;
    local = local + 1;
    box_value.field = local;
    local + value
}

pub fn inner(value: i32) -> i32 {
    value
}

pub fn helper() {}

pub fn invoke<F: Fn()>(callback: F) {
    callback();
}

pub fn invoke_runner<T: Runnable>(value: T) {
    value.run();
}

macro_rules! generated {
    () => {
        helper()
    };
}

pub fn top() {
    let service = Service::factory();
    service.run_local();
    <Service as Runnable>::run(&service);
    Runnable::run(&service);
    invoke(helper);
    invoke(|| child_helper());
    invoke_runner(service);
    let worker: ImportedWorker = Worker::new();
    worker.work();
    child_helper();
    let optional = Some(Service::factory());
    optional.unwrap().run_local();
    generated!();
    println!("builtin");
    std::mem::drop(Service::factory());
}

pub fn enum_build() -> Kind {
    Kind::Data(Service::factory())
}

type ByteCallback = fn(&mut Vec<u8>);

pub fn invoke_aliases(callbacks: Vec<ByteCallback>) {
    for callback in callbacks {
        let mut bytes = Vec::new();
        callback(&mut bytes);
    }
}

type BoolCallback = fn(bool) -> bool;

pub fn invoke_tuple_aliases(cases: Vec<(&str, ByteCallback, BoolCallback)>) {
    for (_label, callback, expected) in cases {
        let mut bytes = Vec::new();
        callback(&mut bytes);
        expected(true);
    }
}

pub fn intentionally_missing() {
    missing();
}

pub mod first {
    pub fn duplicate() {}
}

pub mod second {
    pub fn duplicate() {}
}

pub fn ambiguous() {
    duplicate();
}

pub struct Snapshot;

impl Snapshot {
    pub fn open() -> Result<Self, ()> {
        Ok(Self)
    }

    pub fn info(&self) {}
}

pub fn map_snapshot() {
    Snapshot::open()
        .map(|snapshot| snapshot.info())
        .unwrap();
}

#[derive(Eq, PartialEq)]
pub struct Ordered(u32);

impl Ord for Ordered {
    fn cmp(&self, other: &Self) -> std::cmp::Ordering {
        self.0.cmp(&other.0)
    }
}

impl PartialOrd for Ordered {
    fn partial_cmp(&self, other: &Self) -> Option<std::cmp::Ordering> {
        Some(self.cmp(other))
    }
}

pub fn builtin_collections() -> usize {
    let mut values = std::collections::HashMap::new();
    values.insert("one", 1_u32);
    values.len()
}

pub fn external_call() {
    external_fixture::run();
}

pub enum LocalError {
    Failure(u32),
}

impl LocalError {
    pub fn failure(value: u32) -> Self {
        Self::Failure(value)
    }
}

#[derive(Default)]
pub struct GeneratedDefault;

pub fn generated_default() -> GeneratedDefault {
    GeneratedDefault::default()
}

pub fn captured_snapshot(snapshot: Snapshot) {
    [0_u8].iter().map(|_| snapshot.info()).count();
}

static BUILTIN_MAP: std::sync::OnceLock<
    std::sync::RwLock<std::collections::HashMap<usize, usize>>,
> = std::sync::OnceLock::new();

fn builtin_map(
) -> &'static std::sync::RwLock<std::collections::HashMap<usize, usize>> {
    BUILTIN_MAP.get_or_init(|| {
        std::sync::RwLock::new(std::collections::HashMap::new())
    })
}

pub fn builtin_type_text(pointer: *const u8, value: usize) -> bool {
    let guard = builtin_map().read().expect("builtin map poisoned");
    pointer.is_null() || value.checked_add(guard.len()).is_some()
}

const BUILTIN_SIZE: usize = 1;
static BUILTIN_COUNTER: std::sync::atomic::AtomicUsize =
    std::sync::atomic::AtomicUsize::new(0);

thread_local! {
    static BUILTIN_LOCAL: std::cell::RefCell<String> =
        const { std::cell::RefCell::new(String::new()) };
}

pub struct CallbackRecord {
    pub id: String,
}

impl LocalError {
    pub fn code(&self) -> u32 {
        match self {
            Self::Failure(value) => *value,
        }
    }
}

pub fn builtin_value_flow(records: &[CallbackRecord]) -> usize {
    let total = records
        .iter()
        .try_fold(BUILTIN_SIZE, |total, record| {
            total.checked_add(record.id.len())
        })
        .expect("builtin length overflow");
    BUILTIN_COUNTER.fetch_add(total, std::sync::atomic::Ordering::Relaxed);
    BUILTIN_LOCAL.with(|slot| slot.borrow().len());
    total
}

pub fn map_error_value(value: Result<(), LocalError>) -> Result<(), u32> {
    value.map_err(|error| error.code())
}

#[derive(Default)]
pub struct DefaultBucket {
    pub values: std::collections::HashMap<String, usize>,
}

pub struct LocalHasher;

impl LocalHasher {
    pub fn new() -> Self {
        Self
    }

    pub fn finish(&self) -> usize {
        1
    }
}

pub struct LocalIter {
    next: usize,
}

impl Iterator for LocalIter {
    type Item = usize;

    fn next(&mut self) -> Option<Self::Item> {
        let value = self.next;
        self.next += 1;
        Some(value)
    }
}

pub fn final_runtime_calibration(flag: bool, missing: Option<String>) -> usize {
    use std::fmt::Write as _;

    macro_rules! field {
        ($value:expr) => {
            $value
        };
    }

    let mut output = String::new();
    write!(&mut output, "value").expect("writing to String cannot fail");

    let mut bucket = DefaultBucket::default();
    *bucket.values.entry("value".to_owned()).or_default() += 1;

    let finished = flag
        .then(LocalHasher::new)
        .map(|hasher| hasher.finish())
        .unwrap_or_default();
    let values: Vec<_> = LocalIter { next: 0 }
        .take(1)
        .map(|value| value + 1)
        .collect();
    let supplied = missing.ok_or(()).unwrap_or_default();
    let exists = std::path::PathBuf::from(".")
        .try_exists()
        .unwrap_or_default();
    field!(finished + values.len() + supplied.len() + output.len() + usize::from(exists))
}

