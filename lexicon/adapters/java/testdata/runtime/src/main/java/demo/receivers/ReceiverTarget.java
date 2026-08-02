package demo.receivers;

class ReceiverBase {
    void inherited() {
    }
}

class ReceiverTarget extends ReceiverBase {
    ReceiverTarget() {
    }

    void unique(int value) {
    }

    void overloaded(int value) {
    }

    void overloaded(String value) {
    }

    static void staticOnly() {
    }
}
