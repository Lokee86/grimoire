package demo.runtime;

class RuntimeSlice extends Base implements Contract {
    int field;

    RuntimeSlice() {
        this(1);
    }

    RuntimeSlice(int seed) {
        super(seed);
        this.field = seed;
    }

    public void match(String value) {
    }

    void definite(int value) {
    }

    void overloaded(int value) {
    }

    void overloaded(String value) {
    }

    int exercise(int value, ExternalReceiver external) {
        definite(value);
        this.definite(value);
        overloaded(value);
        Helper.staticCall(value);
        new Helper(value);
        new Choice(value);
        new RuntimeSlice(value);
        vendor.External.staticCall();
        new vendor.External(value);
        external.run();
        missing(value);

        field = value;
        this.field += value;
        value++;
        Helper.shared = value;
        return field + this.field + value + Helper.shared;
    }

    void shadow(int value) {
        int field = value;
        field++;
    }

    void nestedBodies() {
        Runnable lambda = () -> definite(1);
        Runnable anonymous = new Runnable() {
            public void run() {
                definite(1);
            }
        };
        class Local {
            void run() {
                definite(1);
            }
        }
    }
}
