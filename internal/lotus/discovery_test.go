package lotus

const testMinerAddr = "t01000"

//func TestDiscoverer_Discover(t *testing.T) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	gateway := "api.chain.love"
//	disco, err := NewDiscoverer(gateway)
//	assert.NoError(t, err)
//
//	peerID, err := peer.Decode("12D3KooWRqmtFv7ccFfjR7RDcevoMEMXdCHNR8JNN8aNiH2dgk8Z")
//	assert.NoError(t, err)
//
//	d, err := disco.Discover(ctx, peerID, "f049911")
//	assert.NoError(t, err)
//	assert.NotNil(t, d)
//	t.Log(d)
//
//	d2, err := disco.Discover(ctx, peerID, "t01000")
//	assert.EqualError(t, err, "provider id mismatch")
//	assert.Nil(t, d2)
//
//}
