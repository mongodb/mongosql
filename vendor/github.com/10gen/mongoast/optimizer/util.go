package optimizer

import "sort"

func mkStringSet(strs []string) map[string]struct{} {
	ret := make(map[string]struct{}, len(strs))
	for _, str := range strs {
		ret[str] = struct{}{}
	}
	return ret
}

func replaceStringSet(set map[string]struct{}, strs []string) map[string]struct{} {
	// This might be slightly less efficient than just replacing the whole map,
	// but this avoids the need to keep liveFields as a pointer to a map.
	for k := range set {
		delete(set, k)
	}
	return appendToStringSet(set, strs)
}

func appendToStringSet(set map[string]struct{}, strs []string) map[string]struct{} {
	for _, str := range strs {
		set[str] = struct{}{}
	}
	return set
}

func sortedFields(set map[string]struct{}) []string {
	ret := make([]string, 0, len(set))
	for str := range set {
		ret = append(ret, str)
	}
	sort.Strings(ret)
	return ret
}
