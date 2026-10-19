package bibleit

import (
	"strings"
	"testing"
)

func TestCommandArgumentsCannotChangeOperations(t *testing.T) {
	for _, input := range []string{"x\nLIVE DELETE ALL", "x\rPING", "bad\"quote", `He said \"hello\"`, "bad\x00value", ""} {
		constructors := []func() (Command, error){
			func() (Command, error) { return SearchCommand(input, "query") },
			func() (Command, error) { return SearchCommand("web", input) },
			func() (Command, error) { return LiveCommand(input, LiveStart) },
			func() (Command, error) { return ReadCommand(Reference{Translation: "web", Book: input}) },
		}
		for _, build := range constructors {
			if _, err := build(); err == nil {
				t.Fatalf("accepted %q", input)
			}
		}
	}
	c, err := SearchCommand("web", `love \ peace`)
	if err != nil || c.String() != `SEARCH web "love \ peace"` {
		t.Fatalf("%q, %v", c.String(), err)
	}
}

func TestCommandEnumsAreValidated(t *testing.T) {
	if _, err := LiveCommand("abc", LiveAction("DELETE ALL")); err == nil {
		t.Fatal("accepted invalid action")
	}
	if _, err := SetLiveCommand("abc", LiveField("SECRET"), "value"); err == nil {
		t.Fatal("accepted invalid field")
	}
	if _, err := LiveSecretCommand("abc", SecretAction("SET"), "value"); err == nil {
		t.Fatal("accepted unsupported secret set")
	}
	if _, err := ReadCommand(Reference{Translation: "web", Book: "John", Verse: 16}); err == nil {
		t.Fatal("accepted verse without chapter")
	}
}

func TestSecretCommandsMatchServerGrammar(t *testing.T) {
	for action, expected := range map[SecretAction]string{SecretAuthenticate: "LIVE abc SECRET audience-secret", SecretCreate: "LIVE abc SECRET CREATE", SecretRotate: "LIVE abc SECRET ROTATE", SecretDelete: "LIVE abc SECRET DELETE"} {
		value := ""
		if action == SecretAuthenticate {
			value = "audience-secret"
		}
		c, err := LiveSecretCommand("abc", action, value)
		if err != nil || c.String() != expected {
			t.Fatalf("%q, %v", c.String(), err)
		}
	}
}

func TestCommandPreservesLiteralBackslashes(t *testing.T) {
	for _, name := range []string{`plain\path`, `path\`, `literal \n \t \\ path`, `Sunday Service\`} {
		c, err := CreateLiveCommand(name)
		expected := "LIVE CREATE " + name
		if strings.ContainsAny(name, " \t") {
			expected = "LIVE CREATE \"" + name + "\""
		}
		if err != nil || c.String() != expected {
			t.Fatalf("name=%q command=%q error=%v", name, c.String(), err)
		}
	}
}

func TestPopLiveCountBounds(t *testing.T) {
	for _, count := range []int{-1, -2147483648} {
		if _, err := PopLiveCommand("live", count); err == nil {
			t.Fatalf("accepted negative count %d", count)
		}
	}
	for count, wire := range map[int]string{0: "LIVE live STACK POP", 1: "LIVE live STACK POP 1", 2147483647: "LIVE live STACK POP 2147483647"} {
		command, err := PopLiveCommand("live", count)
		if err != nil || command.String() != wire {
			t.Fatalf("count %d: %s %v", count, command.String(), err)
		}
	}
}
