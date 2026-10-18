package official

// SetOfficialOnly says whether to leave out the choice of the best quality upload when
// there is no official video. It may be called while the resolver is at work: the lookups
// that start afterwards follow it.
func (r *Resolver) SetOfficialOnly(on bool) { r.officialOnly.Store(on) }
