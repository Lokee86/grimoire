package com.acme;

enum Mode {
    FAST,
    SAFE;

    private int code = 1;
}

record Result(String value, int count) {
    public Result {
        if (count < 0) {
            throw new IllegalArgumentException("count");
        }
    }

    public String label() {
        return value + count;
    }
}
