package pluralizer

import "regexp"

func newEnglishInflector() *inflector {
	return &inflector{rules: newEnglishRuleSet()}
}

func newEnglishRuleSet() *RuleSet {
	uncountableMap := make(map[string]bool)
	for _, word := range englishUncountableWords {
		uncountableMap[word] = true
	}

	irregularSingularMap := make(map[string]string)
	for singular, plural := range irregularEnglishPluralWords {
		irregularSingularMap[plural] = singular
	}

	return &RuleSet{
		uncountable:       uncountableMap,
		uncountableRegex:  englishUncountableRegexes,
		irregularPlural:   irregularEnglishPluralWords,
		irregularSingular: irregularSingularMap,
		plural:            englishPluralRules,
		singular:          englishSingularRules,
	}
}

var englishUncountableWords = []string{
	"adulthood", "advice", "agenda", "aid", "aircraft", "alcohol", "ammo", "analytics", "anime", "athletics",
	"audio", "bison", "blood", "bream", "buffalo", "butter", "carp", "cash", "chassis", "chess", "clothing",
	"cod", "commerce", "compensation", "cooperation", "corps", "coreopsis", "debris", "deer",
	"diabetes", "digestion", "elk", "energy", "equipment", "evidence", "excretion", "expertise", "feedback",
	"fish", "firmware", "flounder", "flour", "food", "fun", "furniture", "gallows", "garbage", "graffiti",
	"hardware", "headquarters", "health", "help", "herpes", "highjinks", "homework", "horde", "housework",
	"information", "jeans", "jewelry", "judo", "justice", "kin", "knowledge", "kudos", "labour",
	"literature", "livestock", "love", "luggage", "mackerel", "machinery", "mail", "mathematics",
	"merchandise", "mews", "money", "moose", "mud", "manga", "music", "news", "offspring", "only",
	"patience", "pep", "personnel", "physics", "pike", "plankton", "pliers", "police", "pokemon",
	"pollution", "premises", "proceedings", "progress", "rabies", "rain", "research", "rice", "salmon",
	"sand", "scissors", "series", "sewage", "shambles", "sheep", "shrimp", "software", "species", "staff",
	"sugar", "swine", "tennis", "traffic", "transportation", "trout", "tuna", "wealth", "welfare",
	"whiting", "wildebeest", "wildlife", "wood", "you", "sms",
}

var englishUncountableRegexes = []*regexp.Regexp{
	regexp.MustCompile(`(?i)pok[eé]mon$`),
	regexp.MustCompile(`(?i)[^aeiou]ese$`),
	regexp.MustCompile(`(?i)deer$`),
	regexp.MustCompile(`(?i)fish$`),
	regexp.MustCompile(`(?i)measles$`),
	regexp.MustCompile(`(?i)o[iu]s$`),
	regexp.MustCompile(`(?i)pox$`),
	regexp.MustCompile(`(?i)sheep$`),
}

var irregularEnglishPluralWords = map[string]string{
	"i":           "we",
	"me":          "us",
	"he":          "they",
	"she":         "they",
	"them":        "them",
	"myself":      "ourselves",
	"yourself":    "yourselves",
	"itself":      "themselves",
	"herself":     "themselves",
	"himself":     "themselves",
	"themself":    "themselves",
	"is":          "are",
	"was":         "were",
	"has":         "have",
	"this":        "these",
	"that":        "those",
	"my":          "our",
	"its":         "their",
	"his":         "their",
	"her":         "their",
	"thou":        "you",
	"alias":       "aliases",
	"anathema":    "anathemata",
	"analysis":    "analyses",
	"atlas":       "atlases",
	"axe":         "axes",
	"canvas":      "canvases",
	"carve":       "carves",
	"child":       "children",
	"corpus":      "corpora",
	"criterion":   "criteria",
	"datum":       "data",
	"die":         "dice",
	"dingo":       "dingoes",
	"dogma":       "dogmata",
	"eave":        "eaves",
	"echo":        "echoes",
	"foot":        "feet",
	"genus":       "genera",
	"goose":       "geese",
	"groove":      "grooves",
	"human":       "humans",
	"knife":       "knives",
	"leaf":        "leaves",
	"lemma":       "lemmata",
	"looey":       "looies",
	"louse":       "lice",
	"man":         "men",
	"medium":      "media",
	"move":        "moves",
	"mouse":       "mice",
	"opus":        "opuses",
	"ox":          "oxen",
	"parenthesis": "parentheses",
	"passerby":    "passersby",
	"person":      "people",
	"phenomenon":  "phenomena",
	"pickaxe":     "pickaxes",
	"proof":       "proofs",
	"quiz":        "quizzes",
	"radius":      "radii",
	"schema":      "schemata",
	"self":        "selves",
	"sex":         "sexes",
	"soliloquy":   "soliloquies",
	"stigma":      "stigmata",
	"stoma":       "stomata",
	"status":      "statuses",
	"testis":      "testes",
	"thesis":      "theses",
	"thief":       "thieves",
	"tooth":       "teeth",
	"torpedo":     "torpedoes",
	"tornado":     "tornadoes",
	"valve":       "valves",
	"virus":       "viruses",
	"viscus":      "viscera",
	"volcano":     "volcanoes",
	"wife":        "wives",
	"woman":       "women",
	"yes":         "yeses",
	"zombie":      "zombies",
}

var englishPluralRules = []Rule{
	{Pattern: regexp.MustCompile(`(?i)(pe)(?:rson|ople)$`), Replacement: "${1}ople"},
	{Pattern: regexp.MustCompile(`(?i)(m)an$`), Replacement: "${1}en"},
	{Pattern: regexp.MustCompile(`(?i)(child)(?:ren)?$`), Replacement: "${1}ren"},
	{Pattern: regexp.MustCompile(`(?i)(matr|cod|mur|sil|vert|ind|append)(?:ix|ex)$`), Replacement: "${1}ices"},
	{Pattern: regexp.MustCompile(`(?i)\b((?:tit)?m|l)(?:ice|ouse)$`), Replacement: "${1}ice"},
	{Pattern: regexp.MustCompile(`(?i)(x|ch|ss|sh|zz)$`), Replacement: "${1}es"},
	{Pattern: regexp.MustCompile(`(?i)([^aeiouy]|qu)y$`), Replacement: "${1}ies"},
	{Pattern: regexp.MustCompile(`(?i)(?:(kni|wi|li)fe|(ar|l|ea|eo|oa|hoo)f)$`), Replacement: "${1}${2}ves"},
	{Pattern: regexp.MustCompile(`(?i)sis$`), Replacement: "ses"},
	{Pattern: regexp.MustCompile(`(?i)(her|at|gr)o$`), Replacement: "${1}oes"},
	{Pattern: regexp.MustCompile(`(?i)(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|automat|quor)(?:a|um)$`), Replacement: "${1}a"},
	{Pattern: regexp.MustCompile(`(?i)(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)(?:a|on)$`), Replacement: "${1}a"},
	{Pattern: regexp.MustCompile(`(?i)(alumn|syllab|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)(?:us|i)$`), Replacement: "${1}i"},
	{Pattern: regexp.MustCompile(`(?i)(alumn|alg|vertebr)(?:a|ae)$`), Replacement: "${1}ae"},
	{Pattern: regexp.MustCompile(`(?i)(seraph|cherub)(?:im)?$`), Replacement: "${1}im"},
	{Pattern: regexp.MustCompile(`(?i)(alias|[^aou]us|t[lm]as|gas|ris)$`), Replacement: "${1}es"},
	{Pattern: regexp.MustCompile(`(?i)(e[mn]u)s?$`), Replacement: "${1}s"},
	{Pattern: regexp.MustCompile(`(?i)s?$`), Replacement: "s"},
}

var englishSingularRules = []Rule{
	{Pattern: regexp.MustCompile(`(?i)(pe)(rson|ople)$`), Replacement: "${1}rson"},
	{Pattern: regexp.MustCompile(`(?i)(m)en$`), Replacement: "${1}an"},
	{Pattern: regexp.MustCompile(`(?i)(child)ren$`), Replacement: "${1}"},
	{Pattern: regexp.MustCompile(`(?i)(matr|append)ices$`), Replacement: "${1}ix"},
	{Pattern: regexp.MustCompile(`(?i)(cod|mur|sil|vert|ind)ices$`), Replacement: "${1}ex"},
	{Pattern: regexp.MustCompile(`(?i)\b((?:tit)?m|l)ice$`), Replacement: "${1}ouse"},
	{Pattern: regexp.MustCompile(`(?i)(x|ch|ss|sh|zz)es$`), Replacement: "${1}"},
	{Pattern: regexp.MustCompile(`(?i)([^aeiouy]|qu)ies$`), Replacement: "${1}y"},
	{Pattern: regexp.MustCompile(`(?i)(wi|kni|(?:after|half|high|low|mid|non|night|[^\w]|^)li)ves$`), Replacement: "${1}fe"},
	{Pattern: regexp.MustCompile(`(?i)(ar|(?:wo|[ae])l|[eo][ao])ves$`), Replacement: "${1}f"},
	{Pattern: regexp.MustCompile(`(?i)(analy|diagno|parenthe|progno|synop|the|empha|cri|ne)(?:sis|ses)$`), Replacement: "${1}sis"},
	{Pattern: regexp.MustCompile(`(?i)(agend|addend|millenni|dat|extrem|bacteri|desiderat|strat|candelabr|errat|ov|symposi|curricul|quor)a$`), Replacement: "${1}um"},
	{Pattern: regexp.MustCompile(`(?i)(apheli|hyperbat|periheli|asyndet|noumen|phenomen|criteri|organ|prolegomen|hedr|automat)a$`), Replacement: "${1}on"},
	{Pattern: regexp.MustCompile(`(?i)(alumn|alg|vertebr)ae$`), Replacement: "${1}a"},
	{Pattern: regexp.MustCompile(`(?i)(alumn|syllab|vir|radi|nucle|fung|cact|stimul|termin|bacill|foc|uter|loc|strat)i$`), Replacement: "${1}us"},
	{Pattern: regexp.MustCompile(`(?i)(seraph|cherub)im$`), Replacement: "${1}"},
	{Pattern: regexp.MustCompile(`(?i)(alias|[^aou]us|t[lm]as|gas|ris)es$`), Replacement: "${1}"},
	{Pattern: regexp.MustCompile(`(?i)(movie|twelve|abuse|e[mn]u)s$`), Replacement: "${1}"},
	{Pattern: regexp.MustCompile(`(?i)s$`), Replacement: ""},
}
