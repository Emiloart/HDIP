package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func DecodeJSONBody(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(destination); err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}

		return err
	}

	return errors.New("request body must contain a single JSON value")
}
