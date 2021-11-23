package core

var Profiles []Profile

func LoadProfiles() int {
	Profiles = getProfiles()
	return len(Profiles)
}

func (profile *Profile) GetProfilePublicKey() string {
	return EncPublicKey(MarshalPublicKey(&profile.PrivateKey.PublicKey))
}
