package demo.receivers;

import demo.receiver.one.*;
import demo.receiver.two.*;

class ReceiverCalls {
    void parameter(ReceiverTarget receiver) {
        receiver.unique(1);
        receiver.inherited();
    }

    void local() {
        ReceiverTarget receiver = new ReceiverTarget();
        receiver.unique(1);
    }

    void overloaded(ReceiverTarget receiver) {
        receiver.overloaded(1);
    }

    void external(vendor.External receiver) {
        receiver.run();
    }

    void ambiguous(SharedReceiver receiver) {
        receiver.run();
    }

    void scope() {
        future.unique(1);
        ReceiverTarget future = new ReceiverTarget();
        future.unique(1);

        {
            ReceiverTarget inner = new ReceiverTarget();
            inner.unique(1);
        }
        inner.unique(1);

        ReceiverTarget ReceiverTarget = new ReceiverTarget();
        ReceiverTarget.unique(1);
        ReceiverTarget.staticOnly();

        dynamic().unique(1);
        unknown.unique(1);
    }

    ReceiverTarget dynamic() {
        return new ReceiverTarget();
    }
}
