package main

import "testing"

func TestCompilerEligibleSourceSkipsEmbeddedFixtureProjects(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{`C:\repo\impl\src\main\java\example\Main.java`, true},
		{`C:\repo\impl\src\test\java\example\MainTest.java`, true},
		{`C:\repo\its\src\test\resources\sample\src\main\java\Fixture.java`, false},
		{`C:\repo\impl\src\test\projects\sample\src\main\java\Fixture.java`, false},
	}
	for _, test := range tests {
		if got := compilerEligibleSource(test.path); got != test.want {
			t.Fatalf("compilerEligibleSource(%q) = %v, want %v", test.path, got, test.want)
		}
	}
}
