package com.acme;

import com.acme.support.Helper;
import com.acme.support.*;
import java.util.List;
import static com.acme.support.Helper.DEFAULT_NAME;
import static com.acme.support.Helper.*;

@Marker
public class Service implements Worker {
    private final Helper helper;
    private int count, limit = 10;

    public Service(Helper helper) {
        this.helper = helper;
    }

    public String execute(@Marker String input, int... retries) {
        return helper.decorate(input + retries.length + count + DEFAULT_NAME);
    }

    static {
        System.getProperty("lexicon.fixture");
    }

    class Nested {
        long value;

        Nested(long value) {
            this.value = value;
        }
    }
}
