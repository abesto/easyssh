package target

func GivenTargets(targetStrings ...string) []Target {
	targets := make([]Target, len(targetStrings))
	for i := 0; i < len(targetStrings); i++ {
		targets[i] = FromString(targetStrings[i])
	}
	return targets
}
