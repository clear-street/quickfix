package quickfix

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func storedTradeReport() string {
	return assembleStored("8=FIX.4.4",
		[]string{"35=AE", "34=2", "49=CLSTFIX1", "52=20260706-08:00:00.104", "56=NSCC", "115=CLST"},
		[]string{
			"22=1", "30=100", "31=639.18", "32=2", "552=2",
			"54=1", "453=2", "448=695", "452=4", "448=ALG1", "452=1", "136=2", "137=1.50", "139=10", "137=2.25", "139=1",
			"54=2", "453=1", "448=9132", "452=18",
			"15=USD", "381=1278.36", "118=1278.36", "570=N", "571=1",
		})
}

func TestParseMessageForResendSendsTheStoredBodyWithTheSessionHeader(t *testing.T) {
	stored := storedTradeReport()
	msg, err := ParseMessageForResend([]byte(stored))
	require.NoError(t, err)

	msg.Header.SetInt(tagMsgSeqNum, 4711)
	msg.Header.SetString(tagSendingTime, "20261123-19:00:00.000")
	out := msg.build()

	declaredLen, actualLen, declaredCk, computedCk := frameCheck(t, out)
	require.Equal(t, actualLen, declaredLen)
	require.Equal(t, computedCk, declaredCk)
	require.Contains(t, string(out), "\x0134=4711\x01")
	require.Contains(t, string(out), "\x0152=20261123-19:00:00.000\x01")
	require.NotContains(t, string(out), "\x0134=2\x01")

	body := stored[strings.Index(stored, "\x0122=")+1 : strings.LastIndex(stored, "\x0110=")+1]
	require.Contains(t, string(out), body)
}

func TestParseMessageForResendStringReflectsTheHeader(t *testing.T) {
	msg, err := ParseMessageForResend([]byte(storedTradeReport()))
	require.NoError(t, err)

	msg.Header.SetInt(tagMsgSeqNum, 4711)

	require.Contains(t, msg.String(), "\x0134=4711\x01")
	require.Equal(t, msg.String(), string(msg.Bytes()))
}

func TestParseMessageForResendKeepsTheBodyWhenCopied(t *testing.T) {
	msg, err := ParseMessageForResend([]byte(storedTradeReport()))
	require.NoError(t, err)

	copied := NewMessage()
	msg.CopyInto(copied)

	require.Equal(t, string(msg.build()), string(copied.build()))
}

func TestParseMessageForResendDoesNotShareMemoryWithItsInput(t *testing.T) {
	raw := []byte(storedTradeReport())
	msg, err := ParseMessageForResend(raw)
	require.NoError(t, err)
	want := string(msg.build())

	for i := range raw {
		raw[i] = 'x'
	}

	require.Equal(t, want, string(msg.build()))
}

func TestParseMessageForResendRefusesWhatIsNotAMessage(t *testing.T) {
	for name, raw := range map[string]string{
		"empty":        "",
		"not fix":      "hello",
		"without body": assembleStored("8=FIX.4.4", []string{"35=0", "49=A", "56=B"}, nil),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ParseMessageForResend([]byte(raw))

			require.Error(t, err)
		})
	}
}
