pub struct Worker;

impl Worker {
    pub fn new() -> Self {
        Self
    }

    pub fn work(&self) {
        helper();
    }
}

pub fn helper() {}
