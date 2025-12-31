package checksum_test

import (
	"os"
	"testing"

	"github.com/crumbyte/noxdir/command/checksum"

	"github.com/stretchr/testify/require"
)

func TestFileHash_Calculate(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "*")
	require.NoError(t, err)

	_, err = tempFile.WriteString(
		"FileHash allows calculating a file's content hash/checksum.",
	)
	require.NoError(t, err)

	tableData := []struct {
		hashType string
		encoding checksum.Encoding
		expected string
	}{
		{
			hashType: "md5",
			encoding: checksum.HexLower,
			expected: "e360f2198f3639c867d9707dba946404",
		},
		{
			hashType: "md5",
			encoding: checksum.HexUpper,
			expected: "E360F2198F3639C867D9707DBA946404",
		},
		{
			hashType: "md5",
			encoding: checksum.Base64,
			expected: "42DyGY82Ochn2XB9upRkBA==",
		},
		{
			hashType: "sha1",
			encoding: checksum.HexLower,
			expected: "6b9f9fd30f30a772e07b41c6628f100c2af75c15",
		},
		{
			hashType: "sha1",
			encoding: checksum.HexUpper,
			expected: "6B9F9FD30F30A772E07B41C6628F100C2AF75C15",
		},
		{
			hashType: "sha1",
			encoding: checksum.Base64,
			expected: "a5+f0w8wp3Lge0HGYo8QDCr3XBU=",
		},
		{
			hashType: "sha224",
			encoding: checksum.HexLower,
			expected: "820e9659b129060355b7cf6bf057ec61c942d440f1a8eb7e3fe987c8",
		},
		{
			hashType: "sha224",
			encoding: checksum.HexUpper,
			expected: "820E9659B129060355B7CF6BF057EC61C942D440F1A8EB7E3FE987C8",
		},
		{
			hashType: "sha224",
			encoding: checksum.Base64,
			expected: "gg6WWbEpBgNVt89r8FfsYclC1EDxqOt+P+mHyA==",
		},
		{
			hashType: "sha256",
			encoding: checksum.HexLower,
			expected: "c0ca15c2bbee99466084fb49bc24402feb90dda8bf98c6453eb0244dc59dcf6c",
		},
		{
			hashType: "sha256",
			encoding: checksum.HexUpper,
			expected: "C0CA15C2BBEE99466084FB49BC24402FEB90DDA8BF98C6453EB0244DC59DCF6C",
		},
		{
			hashType: "sha256",
			encoding: checksum.Base64,
			expected: "wMoVwrvumUZghPtJvCRAL+uQ3ai/mMZFPrAkTcWdz2w=",
		},
		{
			hashType: "sha384",
			encoding: checksum.HexLower,
			expected: "9571b35d7101532d2abf971fdf73f4243c0bbfc8793a02cf3f6ea7ef54bea2ab045a295cecab13faaa64e466ad52e3f8",
		},
		{
			hashType: "sha384",
			encoding: checksum.HexUpper,
			expected: "9571B35D7101532D2ABF971FDF73F4243C0BBFC8793A02CF3F6EA7EF54BEA2AB045A295CECAB13FAAA64E466AD52E3F8",
		},
		{
			hashType: "sha384",
			encoding: checksum.Base64,
			expected: "lXGzXXEBUy0qv5cf33P0JDwLv8h5OgLPP26n71S+oqsEWilc7KsT+qpk5GatUuP4",
		},
		{
			hashType: "sha512",
			encoding: checksum.HexLower,
			expected: "7fd85c466bf6eed8a86bd27e1a758c59f67b24406b9826c06ad1e18e85b717b54201c60b8b982e9683f6aa65d8b937fc19812e675c774db5e9d8e6aecfa0437f",
		},
		{
			hashType: "sha512",
			encoding: checksum.HexUpper,
			expected: "7FD85C466BF6EED8A86BD27E1A758C59F67B24406B9826C06AD1E18E85B717B54201C60B8B982E9683F6AA65D8B937FC19812E675C774DB5E9D8E6AECFA0437F",
		},
		{
			hashType: "sha512",
			encoding: checksum.Base64,
			expected: "f9hcRmv27tioa9J+GnWMWfZ7JEBrmCbAatHhjoW3F7VCAcYLi5guloP2qmXYuTf8GYEuZ1x3TbXp2Oauz6BDfw==",
		},
	}

	fh := checksum.NewFileHash()

	var (
		rawHash []byte
		fmtHash string
	)

	for _, data := range tableData {
		rawHash, err = fh.Calculate(data.hashType, tempFile.Name())
		require.NoError(t, err)

		fmtHash, err = fh.Format(rawHash, data.encoding)
		require.NoError(t, err)

		require.Equal(t, data.expected, fmtHash)
	}

	require.NoError(t, tempFile.Close())
	require.NoError(t, os.RemoveAll(tempFile.Name()))
}
