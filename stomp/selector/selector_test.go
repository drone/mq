package selector

import "testing"

var evalTests = []struct {
	query string
	param map[string]string
	match bool
}{
	{
		query: "repo-name == 'drone'",
		param: map[string]string{"repo-name": "drone"},
		match: true,
	},
	{
		query: "repo-name == 'drone'",
		param: map[string]string{"repo-name": "drone/drone"},
		match: false,
	},
	{
		query: "repo-name == 'drone'",
		param: map[string]string{},
		match: false,
	},
	{
		query: "repo-name != 'drone'",
		param: map[string]string{"repo-name": "drone/drone"},
		match: true,
	},
	{
		query: "repo-name == 'drone' AND repo-private == true",
		param: map[string]string{"repo-name": "drone", "repo-private": "true"},
		match: true,
	},
	{
		query: "repo-name == 'drone' AND repo-private == true",
		param: map[string]string{"repo-name": "drone", "repo-private": "false"},
		match: false,
	},
	{
		query: "repo-name IN ('drone', 'coverage') OR repo-private == true",
		param: map[string]string{"repo-name": "drone", "repo-private": "false"},
		match: true,
	},
	{
		query: "repo-name IN ('drone', 'coverage') OR repo-private == false",
		param: map[string]string{"repo-name": "rocket", "repo-private": "false"},
		match: true,
	},
	{
		query: "repo-name IN ('drone', 'coverage') OR repo-private == false",
		param: map[string]string{"repo-name": "docker", "repo-private": "true"},
		match: false,
	},
	{
		query: "repo-name NOT IN ('drone', 'coverage')",
		param: map[string]string{"repo-name": "docker"},
		match: true,
	},
	{
		query: "repo-name == 'drone' AND repo-private == true AND repo-vcs == 'git'",
		param: map[string]string{"repo-name": "drone", "repo-private": "true", "repo-vcs": "git"},
		match: true,
	},
	{
		query: "repo-name == 'drone' AND repo-private == true AND repo-vcs == 'git'",
		param: map[string]string{"repo-name": "drone", "repo-private": "true", "repo-vcs": "hg"},
		match: false,
	},
	{
		query: "repo-name == 'drone' AND repo-owner == 'bradrydewski' OR repo-private == false",
		param: map[string]string{"repo-name": "drone", "repo-private": "false", "repo-owner": "bradrydewski"},
		match: true,
	},
	{
		query: "repo-owner == user-login",
		param: map[string]string{"repo-owner": "bradrydewski", "user-login": "bradrydewski"},
		match: true,
	},
	{
		query: "ram >= 2", // >= 2MB RAM
		param: map[string]string{"ram": "1.5"},
		match: false,
	},
	{
		query: "ram >= 2", // >= 2MB RAM
		param: map[string]string{"ram": "2"},
		match: true,
	},
	{
		query: "ram < 4", // < 4MB RAM
		param: map[string]string{"ram": "3.5"},
		match: true,
	},
	{
		query: "ram <= 4", // <= 4MB RAM
		param: map[string]string{"ram": "4"},
		match: true,
	},
	{
		query: "ram <= 3.5", // <= 4MB RAM
		param: map[string]string{"ram": "3.5"},
		match: true,
	},
	{
		query: "ram <= 3", // <= 3MB RAM
		param: map[string]string{"ram": "3.5"},
		match: false,
	},
	{
		query: "cores > 1", // > 1 core
		param: map[string]string{"cores": "1"},
		match: false,
	},
	{
		query: "cores > 1", // > 1 core
		param: map[string]string{"cores": "2"},
		match: true,
	},
	{
		query: "platform == 'linux/amd64'",
		param: map[string]string{"platform": "linux/amd64"},
		match: true,
	},
	{
		query: "platform == 'linux/amd64'",
		param: map[string]string{},
		match: false,
	},
	{
		query: "platform GLOB 'linux/*'",
		param: map[string]string{"platform": "linux/amd64"},
		match: true,
	},
	{
		query: "platform GLOB 'windows/amd64'",
		param: map[string]string{"platform": "linux/amd64"},
		match: false,
	},
	{
		query: "platform NOT GLOB 'windows/amd64'",
		param: map[string]string{"platform": "linux/amd64"},
		match: true,
	},
	{
		query: "platform REGEXP 'linux/(.+)'",
		param: map[string]string{"platform": "linux/amd64"},
		match: true,
	},
	{
		query: "platform REGEXP 'linux/(.+)'",
		param: map[string]string{"platform": "windows/amd64"},
		match: false,
	},
	{
		query: "platform NOT REGEXP 'linux/(.+)'",
		param: map[string]string{"platform": "windows/amd64"},
		match: true,
	},
}

func TestEval(t *testing.T) {
	for _, evalTest := range evalTests {
		query, err := Parse([]byte(evalTest.query))
		if err != nil {
			t.Error(err)
			continue
		}

		match, err := query.Eval(mapRow(evalTest.param))
		if err != nil {
			t.Error(err)
			continue
		}
		if match != evalTest.match {
			t.Errorf("wanted match [%v] for query [%s] and params [%#v]",
				evalTest.match,
				evalTest.query,
				evalTest.param,
			)
		}
	}
}

type mapRow map[string]string

func (m mapRow) Field(name []byte) []byte {
	return []byte(m[string(name)])
}

var result bool

// this benchmark measures the performance of what we expect will be
// representative of most of the real world queries that people are using.
func BenchmarkEval(b *testing.B) {
	buf := []byte("ram >= 2 AND platform == 'linux/amd64'")

	row := mapRow(map[string]string{
		"ram":      "4",
		"platform": "linux/amd64",
	})

	selector, err := Parse(buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		result, err = selector.Eval(row)
		if err != nil {
			b.Fatal(err)
		}
		if result == false {
			b.Fatalf("expected eval returns true")
		}
	}
}

// this benchmark measures the performance of using path.Glob
// to support the SQLITE GLOB keyword.
func BenchmarkEvalGlob(b *testing.B) {
	buf := []byte("platform GLOB 'linux/*'")

	row := mapRow(map[string]string{
		"ram":      "4",
		"platform": "linux/amd64",
	})

	selector, err := Parse(buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		result, err = selector.Eval(row)
		if err != nil {
			b.Fatal(err)
		}
		if result == false {
			b.Fatalf("expected eval returns true")
		}
	}
}

// this benchmark measures the performance of using regexp.Regepx
// to support the SQLITE REGEXP keyword. This performs poorly at
// the moment because we re-compile the regexp on every evaluation.
func BenchmarkEvalRegexp(b *testing.B) {
	buf := []byte("platform REGEXP 'linux/(.+)'")

	row := mapRow(map[string]string{
		"ram":      "4",
		"platform": "linux/amd64",
	})

	selector, err := Parse(buf)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		result, err = selector.Eval(row)
		if err != nil {
			b.Fatal(err)
		}
		if result == false {
			b.Fatalf("expected eval returns true")
		}
	}
}
