package mongodb

type Collation struct {

	// Locale defines the collation locale.
	Locale string `bson:"locale"`

	// CaseLevel defines whether to turn case sensitivity on at strength 1 or 2.
	CaseLevel bool `bson:"caseLevel,omitempty"`

	// CaseFirst may be set to "upper" or "lower" to define whether
	// to have uppercase or lowercase items first. Default is "off".
	CaseFirst string `bson:"caseFirst,omitempty"`

	// Strength defines the priority of comparison properties, as follows:
	//
	//   1 (primary)    - Strongest level, denote difference between base characters
	//   2 (secondary)  - Accents in characters are considered secondary differences
	//   3 (tertiary)   - Upper and lower case differences in characters are
	//                    distinguished at the tertiary level
	//   4 (quaternary) - When punctuation is ignored at level 1-3, an additional
	//                    level can be used to distinguish words with and without
	//                    punctuation. Should only be used if ignoring punctuation
	//                    is required or when processing Japanese text.
	//   5 (identical)  - When all other levels are equal, the identical level is
	//                    used as a tiebreaker. The Unicode code point values of
	//                    the NFD form of each string are compared at this level,
	//                    just in case there is no difference at levels 1-4
	//
	// Strength defaults to 3.
	Strength int `bson:"strength,omitempty"`

	// NumericOrdering defines whether to order numbers based on numerical
	// order and not collation order.
	NumericOrdering bool `bson:"numericOrdering,omitempty"`

	// Alternate controls whether spaces and punctuation are considered base characters.
	// May be set to "non-ignorable" (spaces and punctuation considered base characters)
	// or "shifted" (spaces and punctuation not considered base characters, and only
	// distinguished at strength > 3). Defaults to "non-ignorable".
	Alternate string `bson:"alternate,omitempty"`

	// MaxVariable defines which characters are affected when the value for Alternate is
	// "shifted". It may be set to "punct" to affect punctuation or spaces, or "space" to
	// affect only spaces.
	MaxVariable string `bson:"maxVariable,omitempty"`

	// Normalization defines whether text is normalized into Unicode NFD.
	Normalization bool `bson:"normalization,omitempty"`

	// Backwards defines whether to have secondary differences considered in reverse order,
	// as done in the French language.
	Backwards bool `bson:"backwards,omitempty"`
}
