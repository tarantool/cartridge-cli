package rpm

func genRpmLead(name string) []byte {
	// struct rpmlead {
	//   unsigned char magic[4];
	//   unsigned char major, minor;
	//   short type;
	//   short archnum;
	//   char name[66];
	//   short osnum;
	//   short signature_type;
	//   char reserved[16];
	// } ;

	var rpmLeadName [66]uint8
	for i, nameByte := range []uint8(name) {
		rpmLeadName[i] = nameByte
	}

	rpmLead := packValues(
		[4]uint8{0xed, 0xab, 0xee, 0xdb}, // magic
		uint8(3),                         // major
		uint8(0),                         // minor
		int16(0),                         // type
		int16(1),                         // archnum
		rpmLeadName,                      // name
		int16(1),                         // osnum
		int16(5),                         // signature_type
		[16]int8{},                       // reserved
	)

	return rpmLead
}
