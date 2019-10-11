//+build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
)

func main() {
	newVersion := getNewVersion()
	releaseType := getReleaseType(newVersion)

	log("running release type %q", releaseType)

	switch releaseType {
	case firstBetaRelease:
		releaseFirstBeta(newVersion)
	case subsequentBetaRelease:
		releaseSubsequentBeta(newVersion)
	case minorRelease:
		releaseMinor(newVersion)
	case patchRelease:
		releasePatch(newVersion)
	default:
		fail("unknown release type %q", releaseType)
	}
}

func check(err error) {
	if err == nil {
		return
	}

	msg := err.Error()
	switch typedE := err.(type) {
	case *exec.ExitError:
		stderr := string(typedE.Stderr)
		if stderr != "" {
			msg = fmt.Sprintf("%v: %s", err)
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

func fail(msg string, args ...interface{}) {
	check(fmt.Errorf(msg, args...))
}

func assert(asst bool, msg string) {
	if !asst {
		fail(msg)
	}
}

func log(msg string, args ...interface{}) {
	fmt.Printf(msg, args...)
	fmt.Println()
}

func releaseFirstBeta(v *version) {
	// checkout master
	git("checkout", "master")

	// validate:
	//   - major.minor matches new version
	//   - patch is 0
	//   - pre-release is "dev"
	log("validating version on master")
	mv := getCurrentVersion()
	assert(mv.majorMinorString() == v.majorMinorString(), "expected same major and minor version as master")
	assert(mv.patch == 0, "expected master to have no patch version")
	assert(mv.preRelease.isDev, "expected master to have a dev pre-release")

	// checkout version branch
	git("checkout", gitBranchName(v))

	// write new version to version.txt
	updateVersionTxt(v)

	// commit & tag
	gitBumpAndTagRelease(v)
}

func releaseSubsequentBeta(v *version) {
	// checkout version branch
	git("checkout", gitBranchName(v))

	// validate:
	//   - major.minor.patch matches new version
	//   - pre-release is "beta(N-1)"
	log("validating version on release branch")
	bv := getCurrentVersion()
	assert(bv.majorMinorPatchString() == v.majorMinorPatchString(), "expected same major, minor, and patch version as existing release branch")
	assert(bv.preRelease.betaNumber == v.preRelease.betaNumber-1, "expected beta number to be incremented by 1")

	// write new version to version.txt
	updateVersionTxt(v)

	// commit & tag
	gitBumpAndTagRelease(v)
}

func releaseMinor(v *version) {
	// checkout version branch
	git("checkout", "-b", gitBranchName(v))

	// validate:
	//   - major.minor.patch matches new version
	//   - pre-release is "betaN"
	log("validating version on release branch")
	cv := getCurrentVersion()
	if cv.preRelease.isBeta {
		log("releasing a minor version after a beta")
		assert(cv.majorMinorPatchString() == v.majorMinorPatchString(), "expected same major, minor, and patch version as existing release branch")
	} else {
		log("releasing a minor without a prior beta")
		assert(cv.majorMinorString() == v.majorMinorString(), "expected same major and minor version as existing release branch")
		assert(cv.patch == 0, "expected existing release branch to have no patch version")
		assert(cv.preRelease.isDev, "expected existing release branch to have a dev pre-release")
	}

	// write new version to version.txt
	updateVersionTxt(v)

	// commit & tag
	gitBumpAndTagRelease(v)

	// checkout master
	git("checkout", "master")

	// validate:
	//   - major.minor matches new version
	//   - patch is 0
	//   - pre-release is "dev"
	log("validating version on master")
	mv := getCurrentVersion()
	assert(mv.majorMinorString() == v.majorMinorString(), "expected same major and version as master")
	assert(mv.patch == 0, "expected master to have a patch version of zero")
	assert(mv.preRelease.isDev, "expected master to be a dev release")

	// increment minor version
	mv.minor++

	// update version.txt with dev version
	updateVersionTxt(mv)

	// update build-msi.ps1 with dev version
	updateBuildMSI(mv)

	// commit & tag with dev version
	gitBumpAndTagMaster(mv)
}

func releasePatch(v *version) {
	// checkout version branch
	git("checkout", gitBranchName(v))

	// validate:
	//   - major.minor matches new version
	//   - patch is (N-1)
	//   - pre-release is ""
	log("validating version on release branch")
	bv := getCurrentVersion()
	assert(bv.majorMinorString() == v.majorMinorString(), "expected same major and minor version as existing release branch")
	assert(bv.patch == v.patch-1, "expected patch version to be incremented by 1")
	assert(bv.preRelease.isEmpty(), "expected no pre-release version")

	// write new version to version.txt
	updateVersionTxt(v)

	// commit & tag
	gitBumpAndTagRelease(v)
}

func git(args ...string) {
	log("running git with args %s", args)
	err := exec.Command("git", args...).Run()
	check(err)
}

func gitBumpAndTagMaster(v *version) {
	// commit
	msg := fmt.Sprintf("UPDATE v%s", v)
	git("commit", "-am", msg)

	// tag
	tagMsg := v.String()
	tagName := fmt.Sprintf("v%s", v)
	git("tag", "-am", tagMsg, tagName, "master")
}
func gitBumpAndTagRelease(v *version) {
	// commit
	msg := fmt.Sprintf("BUMP v%s", v)
	git("commit", "-am", msg)

	// tag
	tagMsg := v.String()
	tagName := fmt.Sprintf("v%s", v)
	branchName := gitBranchName(v)
	git("tag", "-am", tagMsg, tagName, branchName)
}

func gitBranchName(v *version) string {
	return fmt.Sprintf("%s.x", v.majorMinorString())
}

func getCurrentVersion() *version {
	b, err := ioutil.ReadFile("version.txt")
	check(err)
	versionStr := strings.TrimSpace(string(b))

	if versionStr[0:1] != "v" {
		fail("expected version.txt version to start with 'v', but found %q", versionStr)
	}
	versionStr = versionStr[1:]

	v, err := parseVersion(versionStr)
	check(err)

	return v
}

func getNewVersion() *version {
	if len(os.Args) != 2 {
		fail("expected exactly one command-line arg, got %d", len(os.Args)-1)
	}
	newVersionArg := os.Args[1]

	v, err := parseVersion(newVersionArg)
	check(err)

	return v
}

func updateVersionTxt(v *version) {
	log("updating version.txt")
	newVersionString := fmt.Sprintf("v%s", v)
	err := ioutil.WriteFile("version.txt", []byte(newVersionString), 0644)
	check(err)
}

func updateBuildMSI(v *version) {
	log("updating build-msi.ps1")

	b, err := ioutil.ReadFile("testdata/bin/build-msi.ps1")
	check(err)
	contents := string(b)

	re, err := regexp.Compile("version -gt 2.1[0-9]")
	check(err)
	contents = re.ReplaceAllString(contents, fmt.Sprintf("version -gt %s", v.majorMinorString()))

	newUUID, err := uuid.NewV4()
	check(err)

	re, err = regexp.Compile("[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}")
	check(err)
	contents = re.ReplaceAllString(contents, newUUID.String())

	err = ioutil.WriteFile("testdata/bin/build-msi.ps1", []byte(contents), 0644)
	check(err)
}

type version struct {
	major      int
	minor      int
	patch      int
	preRelease preRelease
}

func parseVersion(s string) (*version, error) {
	parts := strings.Split(s, "-")

	release := parts[0]

	pre := preRelease{}
	if len(parts) > 1 {
		var err error
		pre, err = parsePreRelease(parts[1])
		if err != nil {
			return nil, err
		}
	}

	versionParts := strings.Split(release, ".")
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return nil, err
	}

	v := &version{
		major:      major,
		minor:      minor,
		patch:      patch,
		preRelease: pre,
	}

	return v, nil
}

func (v *version) majorMinorString() string {
	return fmt.Sprintf("%d.%d", v.major, v.minor)
}

func (v *version) majorMinorPatchString() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func (v *version) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
	pre := v.preRelease.String()
	if pre != "" {
		s += "-" + pre
	}
	return s
}

type preRelease struct {
	isDev      bool
	isBeta     bool
	betaNumber int
}

func parsePreRelease(s string) (preRelease, error) {
	pre := preRelease{}

	if s == "dev" {
		pre.isDev = true
		return pre, nil
	}

	if s[:4] == "beta" {
		pre.isBeta = true
		betaNumber, err := strconv.Atoi(s[4:])
		if err != nil {
			return pre, err
		}
		pre.betaNumber = betaNumber
		return pre, nil
	}

	return pre, fmt.Errorf("invalid pre-release %q", s)
}

func (p preRelease) isEmpty() bool {
	return !(p.isDev || p.isBeta || p.betaNumber > 0)
}

func (p preRelease) String() string {
	if p.isDev {
		return "dev"
	}
	if p.isBeta {
		return fmt.Sprintf("beta%d", p.betaNumber)
	}
	return ""
}

type releaseType string

const (
	firstBetaRelease      releaseType = "first-beta"
	subsequentBetaRelease releaseType = "subsequent-beta"
	minorRelease          releaseType = "minor"
	patchRelease          releaseType = "patch"
)

func getReleaseType(newVersion *version) releaseType {

	// dev -> beta :: firstBetaRelease
	if newVersion.preRelease.betaNumber == 1 {
		return firstBetaRelease
	}

	// beta -> beta :: subsequentBetaRelease
	if newVersion.preRelease.isBeta {
		return subsequentBetaRelease
	}

	// beta -> minor :: minorRelease
	if newVersion.patch == 0 {
		return minorRelease
	}

	// minor.patch -> minor.patch+1 :: patchRelease
	return patchRelease
}
