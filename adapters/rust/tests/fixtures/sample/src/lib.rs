pub mod child;

use crate::child::Worker;
use std::fmt::Debug;

pub trait Runnable {
    fn run(&self);
}

pub struct Service;

impl Runnable for Service {
    fn run(&self) {}
}

impl Service {
    pub fn new() -> Self {
        Service
    }
}

pub enum Kind {
    Ready,
}

pub fn build() -> Service {
    Service::new()
}

macro_rules! generated {
    () => {};
}

pub fn use_worker(_worker: Worker) {}
