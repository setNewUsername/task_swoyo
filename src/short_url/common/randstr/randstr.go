package randstr

import "math/rand"

//returns random string

func CreateRandomString(StringLength int) string {
	const letters = "0123456789abcdefghijklmnopqrstuvwxyz"

	Result := make([]byte, StringLength)

	for i := range Result {
		Result[i] = letters[rand.Intn(len(letters))]
	}

	return string(Result)
}
