package main

type PathLister interface {
	List(input string) ([]cochangeEvidence, string, error)
}

func discover(input string) ([]candidate, error) {
	return discoverWithLister(input, GitPathLister{})
}

func discoverWithLister(input string, lister PathLister) ([]candidate, error) {
	evidence, target, err := lister.List(input)
	if err != nil {
		return nil, err
	}
	return analyzeCandidates(target, evidence), nil
}
