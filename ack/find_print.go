package ack

import "io"

// FindPrint combines the Find and Print functionalities.
// It searches for occurrences of term in the input reader r using the specified findOpts,
// and then prints the results to the writer w using the specified printOpts.
func FindPrint(r io.Reader, w io.Writer, location, term string, findOpts FindOptions, printOpts PrintOptions) error {
	ms, err := Find(r, term, findOpts)
	if err != nil {
		return err
	}
	ms.Print(w, location, printOpts)
	return nil
}
