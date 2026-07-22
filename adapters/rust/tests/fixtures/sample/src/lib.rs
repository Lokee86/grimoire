pub mod child;

use crate::child::Worker;
use std::fmt::Debug;

pub trait Runnable {
    fn run(&self);
}

pub struct Service;

impl Runnable for Service {
    fn run(&self) {
        build();
        self.run();
    }
}

impl Service {
    pub fn new() -> Self {
        helper();
        Service
    }
}

pub enum Kind {
    Ready,
}

pub fn helper() {}

pub fn build() -> Service {
    helper();
    crate::helper();
    Service::new()
}

macro_rules! generated {
    () => {};
}

pub fn use_worker(_worker: Worker) {}

pub fn call_unresolved_forms() {
    missing();
    std::mem::drop(Service::new());
    duplicate();
    generated!();
    (|| {})();
}

pub mod first {
    pub fn duplicate() {}
}

pub mod second {
    pub fn duplicate() {}
}
