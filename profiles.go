package core

var Profiles []ShowProfile

func LoadProfiles() int {
	Profiles = getProfiles()
	return len(Profiles)
}

func (cur *Profile) GetProfilePublicKey() string {
	return EncPublicKey(MarshalPublicKey(&cur.PrivateKey.PublicKey))
}
