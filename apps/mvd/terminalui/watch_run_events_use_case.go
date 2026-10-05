package terminalui

// watchEvents passes every event of in on to the returned channel, in order, after
// folding it into the ledger, and calls onFailure (which must return quickly) for
// each one that reports a failure. The returned channel is closed once in is.
func watchEvents(in <-chan interface{}, ledger *failureLedger, onFailure func()) <-chan interface{} {
	out := make(chan interface{})
	go func() {
		defer close(out)
		for ev := range in {
			if ledger.apply(ev) && onFailure != nil {
				onFailure()
			}
			out <- ev
		}
	}()
	return out
}
