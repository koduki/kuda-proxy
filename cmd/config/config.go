package config

import "os"

type (
	configuration struct {
		TargetURL    string
		UseGoogleJWT bool
		WorkflowID   string
		CorsTarget   string
	}
)

func Load() (configuration, error) {
	config := configuration{
		TargetURL:    os.Getenv("TARGET_URL"),               // "https://kuda-target-dnb6froqha-uc.a.run.app",
		UseGoogleJWT: os.Getenv("USE_GOOGLE_JWT") == "true", // true,
		WorkflowID:   os.Getenv("WORKFLOW_ID"),              //	"projects/{PROJECT_ID}/locations/{LOCATION}/workflows/{WORKFLOW_NAME}"
		CorsTarget:   os.Getenv("CORS_TARGET"),              // "https://kuda-target-dnb6froqha-uc.a.run.app",
	}

	return config, nil
}
