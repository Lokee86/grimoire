package com.acme;

public interface Worker {
    String execute(String input, int... retries);
}
