package service

import (
	"strings"

	"github.com/AlexeytsOY/information-exchange-service/internal/entity"
	apithreat "github.com/AlexeytsOY/information-exchange-service/internal/gen/threat"
)

func enumPtrString[T ~string](e *T) *string {
	if e == nil {
		return nil
	}
	s := string(*e)
	return &s
}

func stringPtrEnum[T ~string](s *string) *T {
	if s == nil {
		return nil
	}
	e := T(*s)
	return &e
}

func lowerString(s string) *string {
	l := strings.ToLower(s)
	return &l
}

func lowerStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	return lowerString(*s)
}

func sliceFromPtr(s *[]string) []string {
	if s == nil {
		return nil
	}
	out := make([]string, len(*s))
	copy(out, *s)
	return out
}

func ptrSlice(s []string) *[]string {
	if s == nil {
		return nil
	}
	out := make([]string, len(s))
	copy(out, s)
	return &out
}

func fromCreateRequest(req *apithreat.ThreatCreateRequest) *entity.Threat {
	t := &entity.Threat{
		MD5:  strings.ToLower(req.Md5),
		Name: req.Name,
	}
	t.SHA1 = lowerStringPtr(req.Sha1)
	t.SHA256 = lowerStringPtr(req.Sha256)

	t.Aliases = sliceFromPtr(req.Aliases)
	t.FileName = sliceFromPtr(req.FileName)
	t.Platform = sliceFromPtr(req.Platform)
	t.Tags = sliceFromPtr(req.Tags)
	t.References = sliceFromPtr(req.References)
	t.Behaviors = sliceFromPtr(req.Behaviors)
	t.RelatedHashes = sliceFromPtr(req.RelatedHashes)

	t.FileSize = req.FileSize
	t.FileType = req.FileType
	t.MimeType = req.MimeType
	t.ThreatType = req.ThreatType
	t.Category = req.Category
	t.Family = req.Family
	t.Severity = enumPtrString(req.Severity)
	t.Status = enumPtrString(req.Status)
	t.Confidence = req.Confidence
	t.FirstSeen = req.FirstSeen
	t.LastSeen = req.LastSeen
	t.Source = sourceFromAPI(req.Source)
	t.Description = req.Description
	t.DetectionRatio = req.DetectionRatio
	t.RecommendedAction = req.RecommendedAction
	t.Detections = detectionsFromAPI(req.Detections)
	t.NetworkIOCs = networkIOCsFromAPI(req.NetworkIocs)
	t.FileIOCs = fileIOCsFromAPI(req.FileIocs)
	t.IsExpired = req.IsExpired
	t.TTL = req.Ttl
	t.Reporter = req.Reporter
	return t
}

// applyReplace полностью заменяет пользовательские поля существующей
// сущности значениями из запроса (семантика HTTP PUT: отсутствующие в
// запросе опциональные поля обнуляются, а не сохраняют прежние значения).
//
// Поля, управляемые сервером (MD5 из path, Version для оптимистичной
// блокировки, CreatedAt/UpdatedAt), здесь не трогаются — их выставляет
// repository и GORM.
func applyReplace(t *entity.Threat, req *apithreat.ThreatUpdateRequest) {
	t.Name = req.Name
	t.SHA1 = lowerStringPtr(req.Sha1)
	t.SHA256 = lowerStringPtr(req.Sha256)

	t.Aliases = sliceFromPtr(req.Aliases)
	t.FileName = sliceFromPtr(req.FileName)
	t.Platform = sliceFromPtr(req.Platform)
	t.Tags = sliceFromPtr(req.Tags)
	t.References = sliceFromPtr(req.References)
	t.Behaviors = sliceFromPtr(req.Behaviors)
	t.RelatedHashes = sliceFromPtr(req.RelatedHashes)

	t.FileSize = req.FileSize
	t.FileType = req.FileType
	t.MimeType = req.MimeType
	t.ThreatType = req.ThreatType
	t.Category = req.Category
	t.Family = req.Family
	t.Severity = enumPtrString(req.Severity)
	t.Status = enumPtrString(req.Status)
	t.Confidence = req.Confidence
	t.FirstSeen = req.FirstSeen
	t.LastSeen = req.LastSeen
	t.Source = sourceFromAPI(req.Source)
	t.Description = req.Description
	t.DetectionRatio = req.DetectionRatio
	t.RecommendedAction = req.RecommendedAction
	t.Detections = detectionsFromAPI(req.Detections)
	t.NetworkIOCs = networkIOCsFromAPI(req.NetworkIocs)
	t.FileIOCs = fileIOCsFromAPI(req.FileIocs)
	t.IsExpired = req.IsExpired
	t.TTL = req.Ttl
	t.Reporter = req.Reporter
}

func sourceFromAPI(in *apithreat.ThreatSource) *entity.ThreatSource {
	if in == nil {
		return nil
	}
	return &entity.ThreatSource{
		Name: in.Name,
		Type: enumPtrString(in.Type),
		URL:  in.Url,
	}
}

func detectionsFromAPI(in *[]apithreat.Detection) []entity.Detection {
	if in == nil {
		return nil
	}
	out := make([]entity.Detection, len(*in))
	for i, d := range *in {
		out[i] = entity.Detection{
			Engine:    d.Engine,
			Version:   d.Version,
			Signature: d.Signature,
		}
	}
	return out
}

func networkIOCsFromAPI(in *[]apithreat.NetworkIOC) []entity.NetworkIOC {
	if in == nil {
		return nil
	}
	out := make([]entity.NetworkIOC, len(*in))
	for i, n := range *in {
		out[i] = entity.NetworkIOC{
			Type:        enumPtrString(n.Type),
			Value:       n.Value,
			Description: n.Description,
		}
	}
	return out
}

func fileIOCsFromAPI(in *[]apithreat.FileIOC) []entity.FileIOC {
	if in == nil {
		return nil
	}
	out := make([]entity.FileIOC, len(*in))
	for i, f := range *in {
		out[i] = entity.FileIOC{
			Type:        enumPtrString(f.Type),
			Value:       f.Value,
			Description: f.Description,
		}
	}
	return out
}

func toAPI(t *entity.Threat) *apithreat.Threat {
	if t == nil {
		return nil
	}
	version := t.Version
	createdAt := t.CreatedAt
	updatedAt := t.UpdatedAt

	return &apithreat.Threat{
		Md5:               t.MD5,
		Sha1:              t.SHA1,
		Sha256:            t.SHA256,
		Name:              t.Name,
		Aliases:           ptrSlice(t.Aliases),
		FileName:          ptrSlice(t.FileName),
		Platform:          ptrSlice(t.Platform),
		Tags:              ptrSlice(t.Tags),
		References:        ptrSlice(t.References),
		Behaviors:         ptrSlice(t.Behaviors),
		RelatedHashes:     ptrSlice(t.RelatedHashes),
		FileSize:          t.FileSize,
		FileType:          t.FileType,
		MimeType:          t.MimeType,
		ThreatType:        t.ThreatType,
		Category:          t.Category,
		Family:            t.Family,
		Severity:          stringPtrEnum[apithreat.ThreatSeverity](t.Severity),
		Status:            stringPtrEnum[apithreat.ThreatStatus](t.Status),
		Confidence:        t.Confidence,
		FirstSeen:         t.FirstSeen,
		LastSeen:          t.LastSeen,
		Source:            sourceToAPI(t.Source),
		Description:       t.Description,
		DetectionRatio:    t.DetectionRatio,
		RecommendedAction: t.RecommendedAction,
		Detections:        detectionsToAPI(t.Detections),
		NetworkIocs:       networkIOCsToAPI(t.NetworkIOCs),
		FileIocs:          fileIOCsToAPI(t.FileIOCs),
		IsExpired:         t.IsExpired,
		Ttl:               t.TTL,
		Reporter:          t.Reporter,
		Version:           &version,
		CreatedAt:         &createdAt,
		UpdatedAt:         &updatedAt,
	}
}

func sourceToAPI(in *entity.ThreatSource) *apithreat.ThreatSource {
	if in == nil {
		return nil
	}
	return &apithreat.ThreatSource{
		Name: in.Name,
		Type: stringPtrEnum[apithreat.ThreatSourceType](in.Type),
		Url:  in.URL,
	}
}

func detectionsToAPI(in []entity.Detection) *[]apithreat.Detection {
	if in == nil {
		return nil
	}
	out := make([]apithreat.Detection, len(in))
	for i, d := range in {
		out[i] = apithreat.Detection{
			Engine:    d.Engine,
			Version:   d.Version,
			Signature: d.Signature,
		}
	}
	return &out
}

func networkIOCsToAPI(in []entity.NetworkIOC) *[]apithreat.NetworkIOC {
	if in == nil {
		return nil
	}
	out := make([]apithreat.NetworkIOC, len(in))
	for i, n := range in {
		out[i] = apithreat.NetworkIOC{
			Type:        stringPtrEnum[apithreat.NetworkIOCType](n.Type),
			Value:       n.Value,
			Description: n.Description,
		}
	}
	return &out
}

func fileIOCsToAPI(in []entity.FileIOC) *[]apithreat.FileIOC {
	if in == nil {
		return nil
	}
	out := make([]apithreat.FileIOC, len(in))
	for i, f := range in {
		out[i] = apithreat.FileIOC{
			Type:        stringPtrEnum[apithreat.FileIOCType](f.Type),
			Value:       f.Value,
			Description: f.Description,
		}
	}
	return &out
}
