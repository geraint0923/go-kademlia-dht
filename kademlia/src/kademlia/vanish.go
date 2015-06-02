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
	"strconv"
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

func CalculateSharedKeyLocations(accessKey int64, epoch int64, count int64) (ids []ID) {
	r := mathrand.New(mathrand.NewSource(accessKey + epoch))
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

const (
	TimePeriod  = int64(2)
	EpochCount  = int64(4)
	EpochPeriod = TimePeriod * EpochCount
)

func getCurrentEpoch() int64 {
	return time.Now().Unix() / EpochPeriod
}

func pushShareKeys(kadem *Kademlia, vdo VanishingDataObject, key []byte) (success int) {
	success = 0

	// split the key
	keyMap, err := sss.Split(vdo.NumberKeys, vdo.Threshold, key)
	if err != nil {
		err = errors.New("Error! Failed to split!")
		fmt.Println(err.Error())
		return
	}
	ids := CalculateSharedKeyLocations(vdo.AccessKey, getCurrentEpoch(), int64(vdo.NumberKeys))
	idx := 0
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
	return
}

func vdoMonitor(kadem *Kademlia, vdo VanishingDataObject, timeout int64) {
	sec := timeout * TimePeriod
	if sec > EpochPeriod {
		sec = EpochPeriod
	}
	prepareSec := int64(1)
	fmt.Println("begin sleep!")
	time.Sleep(time.Second * time.Duration(sec-prepareSec))
	fmt.Println("Extend!")
	_, originKey := UnvanishData(kadem, vdo, false)
	if originKey == nil {
		fmt.Println("Failed to reconstruct the original key when extending time")
		return
	}
	success := pushShareKeys(kadem, vdo, originKey)
	if success < int(vdo.Threshold) {
		fmt.Println("Failed to push share keys when extending time")
		return
	}
	timeout -= EpochCount
	if timeout > 0 {
		go vdoMonitor(kadem, vdo, timeout)
	}
}

func VanishData(kadem *Kademlia, data []byte, numberKeys byte,
	threshold byte, timeout int64) (vdo VanishingDataObject, err error) {
	err = nil
	// generate key for encryption
	key := GenerateRandomCryptoKey()

	// generate VDO for return
	vdo.AccessKey = GenerateRandomAccessKey()
	vdo.Ciphertext = encrypt(key, data)
	vdo.NumberKeys = numberKeys
	vdo.Threshold = threshold

	// push to other nodes
	success := pushShareKeys(kadem, vdo, key)
	fmt.Println("success => " + strconv.Itoa(success))
	fmt.Println("timeout => " + strconv.Itoa(int(timeout)))

	if success < int(vdo.Threshold) {
		err = errors.New("Could not store enough share keys")
	} else if timeout > 0 && timeout*TimePeriod > EpochPeriod {
		// TODO: start a new goroutine to extend tne timeout
		fmt.Println("hehe")
		go vdoMonitor(kadem, vdo, timeout)
	}
	fmt.Println("hehehe => " + strconv.Itoa((int(timeout * TimePeriod))) + " " + strconv.Itoa(int(EpochPeriod)))
	return
}

func UnvanishData(kadem *Kademlia, vdo VanishingDataObject, doDecrypt bool) (data []byte, key []byte) {
	data = nil
	key = nil
	currentEpoch := getCurrentEpoch()
	var success = 0
	// use the current and the neighbor epoch to find the keys
	for epoch := int64(-1); epoch <= 1; epoch++ {
		ids := CalculateSharedKeyLocations(vdo.AccessKey, currentEpoch+epoch, int64(vdo.NumberKeys))
		keyMap := make(map[byte][]byte)
		success = 0
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
			key = sss.Combine(keyMap)
			fmt.Println("current epoch => " + strconv.Itoa(int(epoch)))
			break
		}
	}
	fmt.Println("unv success => " + strconv.Itoa(success))
	if success >= int(vdo.Threshold) && doDecrypt {
		data = decrypt(key, vdo.Ciphertext)
	}
	return
}
