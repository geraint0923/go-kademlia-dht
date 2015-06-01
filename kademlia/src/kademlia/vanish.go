package kademlia

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	mathrand "math/rand"
	"sss"
	"time"
)

type VanishingDataObject struct {
	AccessKey  int64
	Ciphertext []byte
	NumberKeys byte
	Threshold  byte
}

func GenerateRandomCryptoKey() (ret []byte) {
	for i := 0; i < 32; i++ {
		ret = append(ret, uint8(mathrand.Intn(256)))
	}
	return
}

func GenerateRandomAccessKey() (accessKey int64) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	accessKey = r.Int63()
	return
}

func CalculateSharedKeyLocations(accessKey int64, count int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey))
	ids = make([]ID, count)
	for i := int64(0); i < count; i++ {
		for j := 0; j < IDBytes; j++ {
			ids[i][j] = uint8(r.Intn(256))
		}
	}
	return
}

func encrypt(key []byte, text []byte) (ciphertext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext = make([]byte, aes.BlockSize+len(text))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], text)
	return
}

func decrypt(key []byte, ciphertext []byte) (text []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext is not long enough")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return ciphertext
}

func VanishData(kadem *Kademlia, data []byte, numberKeys byte,
	threshold byte) (vdo VanishingDataObject, err error) {
	err = nil
	// generate key for encryption
	key := GenerateRandomCryptoKey()

	// generate VDO for return
	vdo.AccessKey = GenerateRandomAccessKey()
	vdo.Ciphertext = encrypt(key, data)
	vdo.NumberKeys = numberKeys
	vdo.Threshold = threshold

	// split the key
	keyMap, err := sss.Split(numberKeys, threshold, key)
	if err != nil {
		err = errors.New("Error! Failed to split!")
		fmt.Println(err.Error())
		return
	}
	ids := CalculateSharedKeyLocations(vdo.AccessKey, int64(vdo.NumberKeys))
	idx := 0
	success := 0
	for k, v := range keyMap {
		id := ids[idx]
		val := append([]byte{k}, v...)
		// TODO: call Kademlia's function to sprinkle the keys
		// TODO: consider synchronized or asynchronized methods
		_, cl := kadem.DoIterativeStore(id, val)
		if cl != nil && len(cl) > 0 {
			success++
		}
		idx += 1
	}
	if success < int(vdo.Threshold) {
		err = errors.New("Could not store enough share keys")
	}
	return
}

func UnvanishData(kadem *Kademlia, vdo VanishingDataObject) (data []byte) {
	data = nil
	ids := CalculateSharedKeyLocations(vdo.AccessKey, int64(vdo.NumberKeys))
	keyMap := make(map[byte][]byte)
	success := 0
	for _, id := range ids {
		// TODO: collect the shared keys
		// TODO: consider the synchronized and asynchronized methods
		_, val, _ := kadem.DoIterativeFindValue(id)
		if val != nil {
			k := val[0]
			v := val[1:]
			keyMap[k] = v
			success++
		}
	}
	if success >= int(vdo.Threshold) {
		key := sss.Combine(keyMap)
		data = decrypt(key, vdo.Ciphertext)
	}
	return
}
