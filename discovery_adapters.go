package main

import "itoc/internal/discovery"

type GitPathLister = discovery.GitPathLister
type DocumentPathLister = discovery.DocumentPathLister
type unsupportedDocumentFormatError = discovery.UnsupportedDocumentFormatError

var NewDocumentPathLister = discovery.NewDocumentPathLister
