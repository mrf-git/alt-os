package exe

import (
	"alt-os/api"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
)

// Parses command line parameters and initializes an ExeContext.
func InitContext(cmdUsage string, allowedKindRe *regexp.Regexp,
	allowedVersionRe *regexp.Regexp, kindImplMap map[string]interface{},
	respHandlerMap map[string]func(interface{}) error) *ExeContext {

	// Parse command line.
	var infile, format string
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", cmdUsage)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&infile, "i", "", "The input object(s) file to use for container runtime changes")
	flag.StringVar(&format, "f", "", "Input file format (yaml,json), inferred from extension if not specified")
	flag.Parse()

	if infile == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Read and verify input file and initialize context.
	ctxt := &ExeContext{
		ApiServiceContext: api.InitContext(kindImplMap, respHandlerMap),
	}
	if messages, err := api.UnmarshalApiProtoMessages(infile, format); err != nil {
		Fatal("unmarshaling proto messages", err, ctxt)
	} else {
		for _, msg := range messages {
			if !allowedKindRe.MatchString(msg.Kind) {
				Fatal("parsing input", errors.New("bad object kind: "+msg.Kind), ctxt)
			}
			if !allowedVersionRe.MatchString(msg.Version) {
				Fatal("parsing input", errors.New("bad object version: "+msg.Version), ctxt)
			}
		}
		ctxt.ApiServiceContext.MessageQueue = messages
	}

	return ctxt
}
