package validation

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
)

var (
	md5Re    = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	sha1Re   = regexp.MustCompile(`^[a-fA-F0-9]{40}$`)
	sha256Re = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)
)

func MD5(s string) error {
	if !md5Re.MatchString(s) {
		return errors.New("md5 must match ^[a-fA-F0-9]{32}$")
	}
	return nil
}

func CreateRequest(req *apithreat.ThreatCreateRequest) error {
	if err := MD5(req.Md5); err != nil {
		return fmt.Errorf("md5: %w", err)
	}
	if req.Name == "" {
		return errors.New("name: must not be empty")
	}
	if err := validateHashesAndNumbers(
		req.Sha1, req.Sha256, req.FileSize, req.Confidence, req.Ttl, req.Version,
	); err != nil {
		return err
	}
	if err := validateReferences(req.References); err != nil {
		return err
	}
	return validateAssociations(
		req.Severity, req.Status, req.Source, req.NetworkIocs, req.FileIocs,
	)
}

func UpdateRequest(req *apithreat.ThreatUpdateRequest) error {
	if req.Name == "" {
		return errors.New("name: must not be empty")
	}
	if err := validateHashesAndNumbers(
		req.Sha1, req.Sha256, req.FileSize, req.Confidence, req.Ttl, req.Version,
	); err != nil {
		return err
	}
	if err := validateReferences(req.References); err != nil {
		return err
	}
	return validateAssociations(
		req.Severity, req.Status, req.Source, req.NetworkIocs, req.FileIocs,
	)
}

func validateHashesAndNumbers(
	sha1, sha256 *string,
	fileSize *int64,
	confidence *int,
	ttl *int64,
	version *int32,
) error {
	if err := validateSHA1(sha1); err != nil {
		return err
	}
	if err := validateSHA256(sha256); err != nil {
		return err
	}
	if err := validateFileSize(fileSize); err != nil {
		return err
	}
	if err := validateConfidence(confidence); err != nil {
		return err
	}
	if err := validateTTL(ttl); err != nil {
		return err
	}
	return validateVersion(version)
}

func validateAssociations[Sev enum, Stat enum](
	severity *Sev,
	status *Stat,
	source *apithreat.ThreatSource,
	networkIOCs *[]apithreat.NetworkIOC,
	fileIOCs *[]apithreat.FileIOC,
) error {
	if err := validateEnum(severity, "severity"); err != nil {
		return err
	}
	if err := validateEnum(status, "status"); err != nil {
		return err
	}
	if err := validateSource(source); err != nil {
		return err
	}
	if err := validateNetworkIOCs(networkIOCs); err != nil {
		return err
	}
	return validateFileIOCs(fileIOCs)
}

func validateReferences(refs *[]string) error {
	if refs == nil {
		return nil
	}
	for i, r := range *refs {
		if err := validateURI(r); err != nil {
			return fmt.Errorf("references[%d]: %w", i, err)
		}
	}
	return nil
}

func validateURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("must be a valid uri: %w", err)
	}
	if u.Scheme == "" {
		return errors.New("must be an absolute uri (missing scheme)")
	}
	return nil
}

func validateSHA1(v *string) error {
	if v == nil {
		return nil
	}
	if !sha1Re.MatchString(*v) {
		return errors.New("sha1: must match ^[a-fA-F0-9]{40}$")
	}
	return nil
}

func validateSHA256(v *string) error {
	if v == nil {
		return nil
	}
	if !sha256Re.MatchString(*v) {
		return errors.New("sha256: must match ^[a-fA-F0-9]{64}$")
	}
	return nil
}

func validateFileSize(v *int64) error {
	if v == nil {
		return nil
	}
	if *v < 0 {
		return errors.New("file_size: must be >= 0")
	}
	return nil
}

func validateConfidence(v *int) error {
	if v == nil {
		return nil
	}
	if *v < 0 || *v > 100 {
		return errors.New("confidence: must be between 0 and 100")
	}
	return nil
}

func validateTTL(v *int64) error {
	if v == nil {
		return nil
	}
	if *v < 0 {
		return errors.New("ttl: must be >= 0")
	}
	return nil
}

func validateVersion(v *int32) error {
	if v == nil {
		return nil
	}
	if *v < 1 {
		return errors.New("version: must be >= 1")
	}
	return nil
}

type enum interface {
	~string
	Valid() bool
}

func validateEnum[T enum](v *T, field string) error {
	if v == nil {
		return nil
	}
	if !(*v).Valid() {
		return fmt.Errorf("%s: invalid value %q", field, string(*v))
	}
	return nil
}

func validateSource(src *apithreat.ThreatSource) error {
	if src == nil {
		return nil
	}
	if err := validateEnum(src.Type, "source.type"); err != nil {
		return err
	}
	if src.Url != nil {
		if err := validateURI(*src.Url); err != nil {
			return fmt.Errorf("source.url: %w", err)
		}
	}
	return nil
}

func validateNetworkIOCs(iocs *[]apithreat.NetworkIOC) error {
	if iocs == nil {
		return nil
	}
	for i, ioc := range *iocs {
		if err := validateEnum(ioc.Type, fmt.Sprintf("network_iocs[%d].type", i)); err != nil {
			return err
		}
	}
	return nil
}

func validateFileIOCs(iocs *[]apithreat.FileIOC) error {
	if iocs == nil {
		return nil
	}
	for i, ioc := range *iocs {
		if err := validateEnum(ioc.Type, fmt.Sprintf("file_iocs[%d].type", i)); err != nil {
			return err
		}
	}
	return nil
}
