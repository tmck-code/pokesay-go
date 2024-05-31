package pokedex

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"embed"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func EntryFpath(subdir string, idx int) string {
	return path.Join(subdir, fmt.Sprintf("%d.cow", idx))
}

func MetadataFpath(subdir string, idx int) string {
	return path.Join(subdir, fmt.Sprintf("%d.metadata", idx))
}

func CategoryDirpath(subdir string, cat string) string {
	return path.Join(subdir, cat)
}

func CategoryFpath(subdir string, category string, fname string) string {
	return path.Join(subdir, category, fname)
}

func GatherMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0)
	for k, _ := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func ReadStructFromBytes[T any](data []byte) T {
	var d T
	gob.NewDecoder(bytes.NewBuffer(data)).Decode(&d)
	// Check(err)

	return d
}

func WriteStructToFile(obj interface{}, fpath string) {
	ostream, err := os.Create(fpath)
	Check(err)

	writer := bufio.NewWriter(ostream)
	gob.NewEncoder(writer).Encode(obj)

	writer.Flush()
	ostream.Close()
}

func StructToJSON(obj interface{}, indentation ...int) string {
	if len(indentation) == 1 {
		json, err := json.MarshalIndent(obj, "", strings.Repeat(" ", indentation[0]))
		Check(err)
		return string(json)
	} else {
		json, err := json.Marshal(obj)
		Check(err)
		return string(json)
	}
}

func WriteBytesToFile(data []byte, fpath string, compress bool) {
	ostream, err := os.Create(fpath)
	Check(err)

	writer := bufio.NewWriter(ostream)
	if compress {
		writer.WriteString(string(Compress(data)))
	} else {
		writer.WriteString(string(data))
	}
	writer.Flush()
	ostream.Close()
}

func WriteIntToFile(n int, fpath string) {
	WriteBytesToFile([]byte(strconv.Itoa(n)), fpath, false)
}

func ReadIntFromBytes(bs []byte) int {
	total, err := strconv.Atoi(string(bs))
	Check(err)

	return total
}

func Compress(data []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err := gz.Write(data)
	Check(err)

	err = gz.Close()
	Check(err)

	return b.Bytes()
}

func Decompress(data []byte) []byte {
	buf := bytes.NewBuffer(data)

	reader, err := gzip.NewReader(buf)
	Check(err)

	var resB bytes.Buffer

	_, err = resB.ReadFrom(reader)
	Check(err)

	return resB.Bytes()
}

func CreateNameMetadata(idx string, key string, name PokemonName, rootDir string, fpaths []string) *PokemonMetadata {
	entryCategories := make(map[int][][]string, 0)

	for i, fpath := range fpaths {
		localFpath := strings.TrimPrefix(fpath, rootDir)

		if strings.Contains(localFpath, strings.ToLower(name.Slug)) {
			data, err := os.ReadFile(fpath)
			Check(err)
			cats := createCategories(localFpath, data)
			entryCategories[i] = append(entryCategories[i], cats)
		}
	}
	return NewMetadata(
		idx,
		name.English,
		name.Japanese,
		name.JapanesePhonetic,
		entryCategories,
	)
}

func CreateCategoryStruct(rootDir string, metadata []PokemonMetadata, debug bool) []string {
	uniqueCategories := make(map[string]bool)
	for i, m := range metadata {
		for j, entry := range m.Entries {
			for _, cat := range entry.Categories {
				uniqueCategories[cat] = true
				destDir := fmt.Sprintf(
					"build/assets/categories/%s",
					cat,
				)
				os.MkdirAll(destDir, 0755)
				WriteBytesToFile([]byte(fmt.Sprintf("%d/%d", i, j)), fmt.Sprintf("%s/%02d%s", destDir, i, ".cat"), false)
			}
		}
	}
	return GatherMapKeys(uniqueCategories)
}

func createCategories(fpath string, data []byte) []string {
	// split the path into parts, e.g. "gen7x/shiny/"
	parts := strings.Split(fpath, "/")
	// Get the height of the cowfile by counting the number of lines
	height := sizeCategory(len(strings.Split(string(data), "\n")))

	// split the fname into parts, e.g. "charizard-mega-y.cow"
	// fmt.Println(fpath)
	fpathParts := strings.Split(fpath, "/")
	basename := fpathParts[len(fpathParts)-1]
	names := strings.SplitN(strings.TrimSuffix(basename, ".cow"), "-", 2)
	// fmt.Printf("created names from %s: %v\n", fpath, names)

	// return a slice of the parts
	// height, names from the 2nd to the 2nd to last, and parts from the 1st to the 2nd to last

	c := []string{height}
	if len(names) > 1 {
		c = append(c, names[1:]...)
	}
	c = append(c, parts[0:len(parts)-1]...)
	fmt.Printf("fpath %s, names: %v, categories: %v\n", fpath, names, c)
	return c
}

func sizeCategory(height int) string {
	if height <= 13 {
		return "small"
	} else if height <= 19 {
		return "medium"
	}
	return "big"
}

func ReadPokemonCow(embeddedData embed.FS, fpath string) []byte {
	d, err := embeddedData.ReadFile(fpath)
	Check(err)

	return Decompress(d)
}
