// Read a verse using a finite personal token supplied by the application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/bibleit-labs/bibleit-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	c, err := bibleit.NewClient(os.Getenv("BIBLEIT_TOKEN"))
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	r, err := c.GetPersonalIdentity(ctx)
	if err != nil {
		return err
	}
	var identity bibleit.PersonalIdentity
	err = r.Decode(&identity)
	r.Body.Close()
	if err != nil {
		return err
	}
	r, err = c.ReadPersonalVerse(ctx, identity.PersonalAccountID, "web", "43", "3", "16")
	if err != nil {
		return err
	}
	defer r.Body.Close()
	var reading bibleit.ScriptureRead
	if err = r.Decode(&reading); err != nil {
		return err
	}
	for _, verse := range reading.Verses {
		fmt.Println(verse.Text)
	}
	return nil
}
