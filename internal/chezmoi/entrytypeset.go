package chezmoi

// FIXME Add IncludeEncrypted

import (
	"fmt"
	"os"
	"strings"
)

// An EntryTypeSet is a set of entry types. It parses and prints as a
// comma-separated list of strings, but is internally represented as a bitmask.
// *EntryTypeSet implements the github.com/spf13/pflag.Value interface.
type EntryTypeSet struct {
	bits EntryTypeBits
}

// An EntryTypeBits is a bitmask of entry types.
type EntryTypeBits int

// Entry type bits.
const (
	EntryTypeDirs EntryTypeBits = 1 << iota
	EntryTypeFiles
	EntryTypeRemove
	EntryTypeScripts
	EntryTypeSymlinks
	EntryTypeEncrypted

	// EntryTypesAll is all entry types.
	EntryTypesAll EntryTypeBits = EntryTypeDirs | EntryTypeFiles | EntryTypeRemove | EntryTypeScripts | EntryTypeSymlinks | EntryTypeEncrypted

	// EntryTypesNone is no entry types.
	EntryTypesNone EntryTypeBits = 0
)

// entryTypeBits is a map from human-readable strings to EntryTypeBits.
var entryTypeBits = map[string]EntryTypeBits{
	"all":       EntryTypesAll,
	"dirs":      EntryTypeDirs,
	"files":     EntryTypeFiles,
	"remove":    EntryTypeRemove,
	"scripts":   EntryTypeScripts,
	"symlinks":  EntryTypeSymlinks,
	"encrypted": EntryTypeEncrypted,
}

// NewEntryTypeSet returns a new IncludeSet.
func NewEntryTypeSet(bits EntryTypeBits) *EntryTypeSet {
	return &EntryTypeSet{
		bits: bits,
	}
}

// IncludeActualStateEntry returns true if the type of actualStateEntry is a member.
func (s *EntryTypeSet) IncludeActualStateEntry(actualStateEntry ActualStateEntry) bool {
	switch actualStateEntry.(type) {
	case *ActualStateAbsent:
		return s.bits&EntryTypeRemove != 0
	case *ActualStateDir:
		return s.bits&EntryTypeDirs != 0
	case *ActualStateFile:
		return s.bits&EntryTypeFiles != 0
	case *ActualStateSymlink:
		return s.bits&EntryTypeSymlinks != 0
	default:
		return false
	}
}

// IncludeEncrypted returns true if s includes encrypted files.
func (s *EntryTypeSet) IncludeEncrypted() bool {
	return s.bits&EntryTypeEncrypted != 0
}

// IncludeFileInfo returns true if the type of info is a member.
func (s *EntryTypeSet) IncludeFileInfo(info os.FileInfo) bool {
	switch {
	case info.IsDir():
		return s.bits&EntryTypeDirs != 0
	case info.Mode().IsRegular():
		return s.bits&EntryTypeFiles != 0
	case info.Mode()&os.ModeType == os.ModeSymlink:
		return s.bits&EntryTypeSymlinks != 0
	default:
		return false
	}
}

// IncludeTargetStateEntry returns true if type of targetStateEntry is a member.
func (s *EntryTypeSet) IncludeTargetStateEntry(targetStateEntry TargetStateEntry) bool {
	switch targetStateEntry.(type) {
	case *TargetStateDir:
		return s.bits&EntryTypeDirs != 0
	case *TargetStateFile:
		return s.bits&EntryTypeFiles != 0
	case *TargetStateRemove:
		return s.bits&EntryTypeRemove != 0
	case *TargetStateScript:
		return s.bits&EntryTypeScripts != 0
	case *TargetStateSymlink:
		return s.bits&EntryTypeSymlinks != 0
	case *targetStateRenameDir:
		return s.bits&EntryTypeDirs != 0
	default:
		return false
	}
}

// Set implements github.com/spf13/pflag.Value.Set.
func (s *EntryTypeSet) Set(str string) error {
	if str == "none" {
		s.bits = EntryTypesNone
		return nil
	}

	var bits EntryTypeBits
	for i, element := range strings.Split(str, ",") {
		if element == "" {
			continue
		}
		exclude := false
		if strings.HasPrefix(element, "no") {
			exclude = true
			element = element[2:]
		}
		bit, ok := entryTypeBits[element]
		if !ok {
			return fmt.Errorf("%s: unknown entry type", element)
		}
		if i == 0 && exclude {
			bits = EntryTypesAll
		}
		if exclude {
			bits &^= bit
		} else {
			bits |= bit
		}
	}
	s.bits = bits
	return nil
}

func (s *EntryTypeSet) String() string {
	//nolint:exhaustive
	switch s.bits {
	case EntryTypesAll:
		return "all"
	case EntryTypesNone:
		return "none"
	}
	var elements []string
	for i, element := range []string{
		"dirs",
		"files",
		"remove",
		"scripts",
		"symlinks",
	} {
		if s.bits&(1<<i) != 0 {
			elements = append(elements, element)
		}
	}
	return strings.Join(elements, ",")
}

// Sub returns a copy of s with the elements of other removed.
func (s *EntryTypeSet) Sub(other *EntryTypeSet) *EntryTypeSet {
	return &EntryTypeSet{
		bits: (s.bits &^ other.bits) & EntryTypesAll,
	}
}

// Type implements github.com/spf13/pflag.Value.Type.
func (s *EntryTypeSet) Type() string {
	return "entry type set"
}
