package pluralizer

import "github.com/goravel/framework/contracts/support/pluralizer"

var _ pluralizer.Language = (*EnglishLanguage)(nil)

type EnglishLanguage struct {
	singularRuleset pluralizer.Ruleset
	pluralRuleset   pluralizer.Ruleset
}

func NewEnglishLanguage() *EnglishLanguage {
	return &EnglishLanguage{
		singularRuleset: newEnglishSingularRuleset(),
		pluralRuleset:   newEnglishPluralRuleset(),
	}
}

func (r *EnglishLanguage) Name() string {
	return "english"
}

func (r *EnglishLanguage) SingularRuleset() pluralizer.Ruleset {
	return r.singularRuleset
}

func (r *EnglishLanguage) PluralRuleset() pluralizer.Ruleset {
	return r.pluralRuleset
}

func newEnglishPluralRuleset() pluralizer.Ruleset {
	uninflected := pluralizer.Patterns{
		NewPattern(`(?i)\b(bison|cod|deer|fish|moose|offspring|pike|salmon|sheep|species|squid|swine|trout|tuna|shrimp)\b`),
		NewPattern(`(?i)^(bison|cod|deer|fish|moose|offspring|pike|salmon|sheep|species|squid|swine|trout|tuna|shrimp)$`),
		NewPattern(`(?i)advice`), NewPattern(`(?i)aircraft`), NewPattern(`(?i)art`),
		NewPattern(`(?i)audio`), NewPattern(`(?i)baggage`), NewPattern(`(?i)butter`),
		NewPattern(`(?i)bread`), NewPattern(`(?i)cash`), NewPattern(`(?i)cattle`),
		NewPattern(`(?i)chassis`), NewPattern(`(?i)cheese`), NewPattern(`(?i)clothing`),
		NewPattern(`(?i)coal`), NewPattern(`(?i)compensation`), NewPattern(`(?i)coreopsis`),
		NewPattern(`(?i)cotton`), NewPattern(`(?i)data`), NewPattern(`(?i)education`),
		NewPattern(`(?i)equipment`), NewPattern(`(?i)evidence`), NewPattern(`(?i)feedback`),
		NewPattern(`(?i)flour`), NewPattern(`(?i)food`), NewPattern(`(?i)furniture`),
		NewPattern(`(?i)garbage`), NewPattern(`(?i)homework`),
		NewPattern(`(?i)impatience`), NewPattern(`(?i)information`), NewPattern(`(?i)jeans`),
		NewPattern(`(?i)knowledge`), NewPattern(`(?i)leather`), NewPattern(`(?i)love`),
		NewPattern(`(?i)luggage`), NewPattern(`(?i)mathematics`), NewPattern(`(?i)means`),
		NewPattern(`(?i)milk`), NewPattern(`(?i)money`),
		NewPattern(`(?i)music`), NewPattern(`(?i)news`), NewPattern(`(?i)nutrition`),
		NewPattern(`(?i)oil`), NewPattern(`(?i)pants`), NewPattern(`(?i)paper`),
		NewPattern(`(?i)patience`), NewPattern(`(?i)police`), NewPattern(`(?i)pollen`),
		NewPattern(`(?i)progress`), NewPattern(`(?i)rain`), NewPattern(`(?i)research`),
		NewPattern(`(?i)rice`), NewPattern(`(?i)series`), NewPattern(`(?i)scissors`),
		NewPattern(`(?i)shorts`), NewPattern(`(?i)software`), NewPattern(`(?i)species`),
		NewPattern(`(?i)staff`), NewPattern(`(?i)sugar`), NewPattern(`(?i)swine`),
		NewPattern(`(?i)traffic`), NewPattern(`(?i)water`), NewPattern(`(?i)weather`),
		NewPattern(`(?i)wheat`), NewPattern(`(?i)wood`), NewPattern(`(?i)wool`),
		NewPattern(`(?i)you`),
	}

	irregular := pluralizer.Substitutions{
		NewSubstitution("alumnus", "alumni"),
		NewSubstitution("apparatus", "apparatuses"),
		NewSubstitution("appendix", "appendices"),
		NewSubstitution("cactus", "cacti"),
		NewSubstitution("child", "children"),
		NewSubstitution("corpus", "corpora"),
		NewSubstitution("criterion", "criteria"),
		NewSubstitution("curriculum", "curricula"),
		NewSubstitution("die", "dice"),
		NewSubstitution("foot", "feet"),
		NewSubstitution("genus", "genera"),
		NewSubstitution("goose", "geese"),
		NewSubstitution("graffito", "graffiti"),
		NewSubstitution("index", "indices"),
		NewSubstitution("larva", "larvae"),
		NewSubstitution("louse", "lice"),
		NewSubstitution("man", "men"),
		NewSubstitution("medium", "media"),
		NewSubstitution("memorandum", "memoranda"),
		NewSubstitution("mouse", "mice"),
		NewSubstitution("nucleus", "nuclei"),
		NewSubstitution("opus", "opuses"),
		NewSubstitution("ox", "oxen"),
		NewSubstitution("person", "people"),
		NewSubstitution("phenomenon", "phenomena"),
		NewSubstitution("phylum", "phyla"),
		NewSubstitution("quiz", "quizzes"),
		NewSubstitution("stimulus", "stimuli"),
		NewSubstitution("stratum", "strata"),
		NewSubstitution("syllabus", "syllabi"),
		NewSubstitution("symposium", "symposia"),
		NewSubstitution("synopsis", "synopses"),
		NewSubstitution("testis", "testes"),
		NewSubstitution("thesis", "theses"),
		NewSubstitution("tooth", "teeth"),
		NewSubstitution("woman", "women"),
		NewSubstitution("abuse", "abuses"),
		NewSubstitution("atlas", "atlases"),
		NewSubstitution("avalanche", "avalanches"),
		NewSubstitution("axis", "axes"),
		NewSubstitution("axe", "axes"),
		NewSubstitution("beef", "beefs"),
		NewSubstitution("blouse", "blouses"),
		NewSubstitution("brother", "brothers"),
		NewSubstitution("brownie", "brownies"),
		NewSubstitution("cache", "caches"),
		NewSubstitution("cafe", "cafes"),
		NewSubstitution("canvas", "canvases"),
		NewSubstitution("cave", "caves"),
		NewSubstitution("chateau", "chateaux"),
		NewSubstitution("cookie", "cookies"),
		NewSubstitution("cow", "cows"),
		NewSubstitution("curve", "curves"),
		NewSubstitution("demo", "demos"),
		NewSubstitution("domino", "dominoes"),
		NewSubstitution("echo", "echoes"),
		NewSubstitution("emphasis", "emphases"),
		NewSubstitution("epoch", "epochs"),
		NewSubstitution("foe", "foes"),
		NewSubstitution("fungus", "fungi"),
		NewSubstitution("ganglion", "ganglions"),
		NewSubstitution("gas", "gases"),
		NewSubstitution("genie", "genies"),
		NewSubstitution("grave", "graves"),
		NewSubstitution("hippopotamus", "hippopotami"),
		NewSubstitution("hoax", "hoaxes"),
		NewSubstitution("hoof", "hoofs"),
		NewSubstitution("human", "humans"),
		NewSubstitution("iris", "irises"),
		NewSubstitution("leaf", "leaves"),
		NewSubstitution("lens", "lenses"),
		NewSubstitution("loaf", "loaves"),
		NewSubstitution("mongoose", "mongooses"),
		NewSubstitution("motto", "mottoes"),
		NewSubstitution("move", "moves"),
		NewSubstitution("mythos", "mythoi"),
		NewSubstitution("neurosis", "neuroses"),
		NewSubstitution("niche", "niches"),
		NewSubstitution("niveau", "niveaux"),
		NewSubstitution("numen", "numina"),
		NewSubstitution("oasis", "oases"),
		NewSubstitution("occiput", "occiputs"),
		NewSubstitution("octopus", "octopuses"),
		NewSubstitution("passerby", "passersby"),
		NewSubstitution("penis", "penises"),
		NewSubstitution("plateau", "plateaux"),
		NewSubstitution("runner-up", "runners-up"),
		NewSubstitution("safe", "safes"),
		NewSubstitution("save", "saves"),
		NewSubstitution("sex", "sexes"),
		NewSubstitution("sieve", "sieves"),
		NewSubstitution("soliloquy", "soliloquies"),
		NewSubstitution("son-in-law", "sons-in-law"),
		NewSubstitution("daughter-in-law", "daughters-in-law"),
		NewSubstitution("mother-in-law", "mothers-in-law"),
		NewSubstitution("father-in-law", "fathers-in-law"),
		NewSubstitution("sister-in-law", "sisters-in-law"),
		NewSubstitution("brother-in-law", "brothers-in-law"),
		NewSubstitution("attorney-general", "attorneys-general"),
		NewSubstitution("passer-by", "passers-by"),
		NewSubstitution("stadium", "stadiums"),
		NewSubstitution("thief", "thieves"),
		NewSubstitution("tornado", "tornadoes"),
		NewSubstitution("trilby", "trilbys"),
		NewSubstitution("turf", "turfs"),
		NewSubstitution("valve", "valves"),
		NewSubstitution("volcano", "volcanoes"),
		NewSubstitution("wave", "waves"),
		NewSubstitution("zombie", "zombies"),
	}

	regular := pluralizer.Transformations{
		NewTransformation(`(?i)(quiz)$`, `${1}zes`),
		NewTransformation(`(?i)^(ox)$`, `${1}en`),
		NewTransformation(`(?i)([ml])ouse$`, `${1}ice`),
		NewTransformation(`(?i)(matr|vert|ind)(ix|ex)$`, `${1}ices`),
		NewTransformation(`(?i)(x|ch|ss|sh|z)$`, `${1}es`),
		NewTransformation(`(?i)([^aeiouy]|qu)y$`, `${1}ies`),
		NewTransformation(`(?i)(hive)$`, `${1}s`),
		NewTransformation(`(?i)(lea|loa|thie|shea|shel|hal|cal|wol|kni)f$`, `${1}ves`),
		NewTransformation(`(?i)(li|wi)fe$`, `${1}ves`),
		NewTransformation(`(?i)sis$`, `ses`),
		NewTransformation(`(?i)([ti])um$`, `${1}a`),
		NewTransformation(`(?i)(tomat|potat|ech|her|vet)o$`, `${1}oes`),
		NewTransformation(`(?i)(bu)s$`, `${1}ses`),
		NewTransformation(`(?i)(alias|status)$`, `${1}es`),
		NewTransformation(`(?i)(octop|vir)us$`, `${1}i`),
		NewTransformation(`(?i)(ax|test)is$`, `${1}es`),
		NewTransformation(`(?i)s$`, `s`),
		NewTransformation(`$`, `s`),
	}

	return NewRuleset(regular, uninflected, irregular)
}

func newEnglishSingularRuleset() pluralizer.Ruleset {
	uninflected := newEnglishPluralRuleset().Uninflected()

	irregular := pluralizer.Substitutions{}
	for _, sub := range newEnglishPluralRuleset().Irregular() {
		irregular = append(irregular, NewSubstitution(sub.To(), sub.From()))
	}
	irregular = append(irregular, NewSubstitution("trousers", "trouser"))

	regular := pluralizer.Transformations{
		NewTransformation(`(?i)(quiz)zes$`, `${1}`),
		NewTransformation(`(?i)(matr)ices$`, `${1}ix`),
		NewTransformation(`(?i)(vert|ind)ices$`, `${1}ex`),
		NewTransformation(`(?i)^(ox)en`, `${1}`),
		NewTransformation(`(?i)(alias|status)(es)?$`, `${1}`),
		NewTransformation(`(?i)(octop|vir)(i|uses)$`, `${1}us`),
		NewTransformation(`(?i)(cris|ax|test)es$`, `${1}is`),
		NewTransformation(`(?i)(shoe)s$`, `${1}`),
		NewTransformation(`(?i)(o)es$`, `${1}`),
		NewTransformation(`(?i)(bus)es$`, `${1}`),
		NewTransformation(`(?i)([ml])ice$`, `${1}ouse`),
		NewTransformation(`(?i)(x|ch|ss|sh|z)es$`, `${1}`),
		NewTransformation(`(?i)(m)ovies$`, `${1}ovie`),
		NewTransformation(`(?i)(s)eries$`, `${1}eries`),
		NewTransformation(`(?i)([^aeiouy]|qu)ies$`, `${1}y`),
		NewTransformation(`(?i)(lea|loa|thie|shea|shel|hal|cal|wol|kni)ves$`, `${1}f`),
		NewTransformation(`(?i)(li|wi)ves$`, `${1}fe`),
		NewTransformation(`(?i)(tive)s$`, `${1}`),
		NewTransformation(`(?i)(hive)s$`, `${1}`),
		NewTransformation(`(?i)(^analy)ses$`, `${1}sis`),
		NewTransformation(`(?i)((a)naly|(b)a|(d)iagno|(p)arenthe|(p)rogno|(s)ynop|(t)he)ses$`, `${1}${2}sis`),
		NewTransformation(`(?i)([ti])a$`, `${1}um`),
		NewTransformation(`(?i)(n)ews$`, `${1}ews`),
		NewTransformation(`(?i)(ss)$`, `${1}`),
		NewTransformation(`(?i)s$`, ``),
	}

	return NewRuleset(regular, uninflected, irregular)
}
