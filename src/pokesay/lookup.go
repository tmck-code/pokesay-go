package pokesay

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/tmck-code/pokesay/src/pokedex"
)

var (
	Rand rand.Source = rand.NewSource(time.Now().UnixNano())
)

func RandomInt(n int) int {
	if n <= 0 {
		return 0
	}
	return rand.New(Rand).Intn(n)
}

// ChooseByCategory chooses a pokemon via a requested category
// 1. It loads the category search structure and finds the name of a random Pokemon matching the entry
// e.g. if given the category "small", this function might pick the file `1.cat` in
// - categories/
//   - small/
//   - 1.cat
//   - 44.cat
//
// This file contains entries representing the <pokemon metadata index>/<the pokemon entry index>,
// e.g. "4/1" would represent 4.metadata, and the 2nd entry in that file
// 2. Using the indexes, load the corresponding metadata file and entry, and then return it
func ChooseByCategory(category string, categoryDir []fs.DirEntry, categoryFiles embed.FS, categoryRootDir string, metadataFiles embed.FS, metadataRootDir string) (pokedex.PokemonMetadata, pokedex.PokemonEntryMapping) {
	if len(categoryDir) == 0 {
		log.Fatalf("cannot find pokemon by category '%s'", category)
	}
	choice := categoryDir[RandomInt(len(categoryDir))]

	categoryMetadata, err := categoryFiles.ReadFile(
		pokedex.CategoryFpath(categoryRootDir, category, choice.Name()),
	)
	pokedex.Check(err)

	parts := strings.Split(string(categoryMetadata), "/")

	metadata := pokedex.ReadMetadataFromEmbedded(
		metadataFiles,
		path.Join(metadataRootDir, fmt.Sprintf("%s.metadata", parts[0])),
	)

	entryIndex, err := strconv.Atoi(string(parts[1]))
	pokedex.Check(err)

	return metadata, metadata.Entries[entryIndex]
}

func ListNames(names map[string][]int) []string {
	return pokedex.GatherMapKeys(names)
}

func fetchMetadataByName(names map[string][]int, nameToken string, metadataFiles embed.FS, metadataRootDir string) pokedex.PokemonMetadata {
	match := names[nameToken]
	if len(match) == 0 {
		log.Fatalf("cannot find pokemon by name '%s'", nameToken)
	}
	nameChoice := match[RandomInt(len(match))]

	metadata := pokedex.ReadMetadataFromEmbedded(
		metadataFiles,
		pokedex.MetadataFpath(metadataRootDir, nameChoice),
	)
	return metadata
}

func ChooseByIndex(idx int, entryIdx int, metadataFiles embed.FS, metadataRootDir string) (pokedex.PokemonMetadata, pokedex.PokemonEntryMapping, error) {
	metadata := pokedex.ReadMetadataFromEmbedded(
		metadataFiles,
		pokedex.MetadataFpath(metadataRootDir, idx),
	)
	for _, entry := range metadata.Entries {
		if entry.EntryIndex == entryIdx {
			return metadata, entry, nil
		}
	}
	return pokedex.PokemonMetadata{}, pokedex.PokemonEntryMapping{}, errors.New("could not find pokemon by index")
}

func ChooseByName(names map[string][]int, nameToken string, metadataFiles embed.FS, metadataRootDir string) (pokedex.PokemonMetadata, pokedex.PokemonEntryMapping) {
	metadata := fetchMetadataByName(
		names,
		nameToken,
		metadataFiles,
		metadataRootDir,
	)

	// pick a random entry
	choice := RandomInt(len(metadata.Entries))
	return metadata, metadata.Entries[choice]
}

func ChooseByNameAndCategory(names map[string][]int, nameToken string, metadataFiles embed.FS, metadataRootDir string, category string) (pokedex.PokemonMetadata, pokedex.PokemonEntryMapping) {
	// fetch the metadata of a pokemon matching the nameToken
	metadata := fetchMetadataByName(
		names,
		nameToken,
		metadataFiles,
		metadataRootDir,
	)

	// now try and find a metadata entry that matches the requested category
	matching := make([]pokedex.PokemonEntryMapping, 0)
	for _, entry := range metadata.Entries {
		for _, entryCategory := range entry.Categories {
			if entryCategory == category {
				matching = append(matching, entry)
			}
		}
	}

	// if the category is not found for this pokemon, return a random entry
	if len(matching) == 0 {
		return metadata, metadata.Entries[RandomInt(len(metadata.Entries))]
	} else {
		return metadata, matching[RandomInt(len(matching))]
	}
}

func ChooseByRandomIndex(totalInBytes []byte) (int, int) {
	total := pokedex.ReadIntFromBytes(totalInBytes)
	return total, RandomInt(total)
}
