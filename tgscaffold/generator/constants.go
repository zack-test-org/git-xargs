package generator

const (
	AWS = "aws"
	GCP = "gcp"
)

type ConstantsType struct {
	AWS string
	GCP string
}

var Constants = ConstantsType{
	AWS: AWS,
	GCP: GCP,
}
