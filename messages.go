package main

type opaque byte

type ProtocolVersion uint16
type Random [32]byte

type CipherSuite [2]uint8 /* Cryptographic suite selector */

type ClientHelloMessage struct {
	LegacyVersion            ProtocolVersion // = 0x0303;    /* TLS v1.2 */
	Random                   opaque
	LegacySessionID          opaque
	CipherSuites             CipherSuite
	LegacyCompressionMethods opaque
	// Extensions               Extension
}

type Extension struct {
	ExtensionType ExtensionType
	ExtensionData opaque
}

type ExtensionType uint8

const (
	ServerName                          ExtensionType = 0
	MaxFragmentLength                                 = 1
	StatusRequest                                     = 5
	SupportedGroups                                   = 10
	SignatureAlgorithms                               = 13
	UseSRTP                                           = 14
	Heartbeat                                         = 15
	ApplicationLayerProtocolNegotiation               = 16
	SignedCertificateTimestamp                        = 18
	ClientCertificateType                             = 19
	ServerCertificateType                             = 20
	Padding                                           = 21
	PreSharedKey                                      = 41
	EarlyData                                         = 42
	SupportedVersions                                 = 43
	Cookie                                            = 44
	PSKKeyExchangeModes                               = 45
	CertificateAuthoroties                            = 47
	OIDFilters                                        = 48
	PostHandshakeAuth                                 = 49
	SignatureAlgorithmsCert                           = 50
	KeyShares                                         = 51
)
