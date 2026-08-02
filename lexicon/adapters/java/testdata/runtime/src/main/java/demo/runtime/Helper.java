package demo.runtime;

class Helper {
    static int shared;

    Helper(int seed) {
        shared = seed;
    }

    static void staticCall(int value) {
        shared = value;
    }
}

class Choice {
    Choice(int value) {
    }

    Choice(String value) {
    }
}
