package types

type S3ObjectInfo struct {
	Bucket string
	Key    string
}

type S3Record struct {
	S3 struct {
		Bucket struct {
			Name string `json:"name"`
		} `json:"bucket"`
		Object struct {
			Key string `json:"key"`
		} `json:"object"`
	} `json:"s3"`
}

type S3ObjectCreatedEvent struct {
	Records []S3Record `json:"Records"`
}
