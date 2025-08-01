package loadtest

import "os"

func envCleanup() {
	os.Unsetenv("TEST_1")
	os.Unsetenv("TEST_2")
	os.Unsetenv("Test3")
}
