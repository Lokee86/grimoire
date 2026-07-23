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
