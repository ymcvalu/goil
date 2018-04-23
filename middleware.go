package goil

func combineChain(chain HandlerChain, handlers ...HandlerFunc) HandlerChain {
	if len(handlers) == 0 || handlers == nil {
		return chain
	}
	if len(chain) == 0 || chain == nil {
		return handlers
	}
	hc := make(HandlerChain, len(chain)+len(handlers))
	copy(hc, chain)
	copy(hc[len(chain):], handlers)
	return hc
}
