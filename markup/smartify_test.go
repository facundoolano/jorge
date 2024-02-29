package markup

import (
	"io"
	"strings"
	"testing"
)

func TestSmartify(t *testing.T) {
	input := `<html>
<head>
<script type="text/javascript">
const url = 'http://localhost:4001/_events/';
const string = "joe's garage";
</script>
</head>
<body>
<p>the album is "Joe's Garage" --by Frank Zappa...</p>
</body>
</html>`

	output, err := Smartify(".html", strings.NewReader(input))
	assertEqual(t, err, nil)
	buf := new(strings.Builder)
	_, err = io.Copy(buf, output)
	assertEqual(t, err, nil)

	assertEqual(t, buf.String(), `<html><head>
<script type="text/javascript">
const url = 'http://localhost:4001/_events/';
const string = "joe's garage";
</script>
</head>
<body>
<p>the album is “Joe’s Garage” –by Frank Zappa…</p>

</body></html>`)
}
