use super::Runnable;

pub struct Worker;

pub trait LocalTrait {}

impl LocalTrait for Worker {}

pub mod nested {
    pub fn nested_fn() {}
}
