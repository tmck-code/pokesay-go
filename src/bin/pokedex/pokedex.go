package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/tmck-code/pokesay/src/bin"
	"github.com/tmck-code/pokesay/src/pokedex"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

// Strips the leading "./" from a path e.g. "./cows/ -> cows/"
func normaliseRelativeDir(dirPath string) string {
	return strings.TrimPrefix(dirPath, "./")
}

type PokedexArgs struct {
	FromDir           string
	FromMetadataFname string
	ToDir             string
	Debug             bool
	ToCategoryFname   string
	ToDataSubDir      string
	ToMetadataSubDir  string
	ToTotalFname      string
}

func parseArgs() PokedexArgs {
	fromDir := flag.String("from", "/tmp/cows", "from dir")
	fromMetadataFname := flag.String("fromMetadata", "/tmp/cows/pokemon.json", "metadata file")
	toDir := flag.String("to", "build/assets/", "to dir")

	toDataSubDir := flag.String("toDataSubDir", "cows/", "dir to write all binary (image) data to")
	toMetadataSubDir := flag.String("toMetadataSubDir", "metadata/", "dir to write all binary (metadata) data to")
	toCategoryFname := flag.String("toCategoryFpath", "pokedex.gob", "to fpath")
	toTotalFname := flag.String("toTotalFname", "total.txt", "file to write the number of available entries to")
	debug := flag.Bool("debug", false, "show debug logs")

	flag.Parse()

	args := PokedexArgs{
		FromDir:           normaliseRelativeDir(*fromDir),
		FromMetadataFname: *fromMetadataFname,
		ToDir:             normaliseRelativeDir(*toDir),
		ToCategoryFname:   *toCategoryFname,
		ToDataSubDir:      normaliseRelativeDir(*toDataSubDir),
		ToMetadataSubDir:  normaliseRelativeDir(*toMetadataSubDir),
		ToTotalFname:      *toTotalFname,
		Debug:             *debug,
	}
	if args.Debug {
		fmt.Printf("%+v\n", args)
	}
	return args
}

func mkDirs(dirPaths []string) {
	for _, dirPath := range dirPaths {
		err := os.MkdirAll(dirPath, 0755)
		check(err)
	}
}

// This function reads in the files given by the PokedexArgs, and generates the data that pokesay will use when running
// - The "category" struct
//   - contains category information, and the index of the corresponding metadata file
//
// - The "metadata" files
//   - named like 1.metadata, contains pokemon info like name, categories, japanese name
//
// - The "data" files
//   - contain the pokemon as gzipped text
//
// - The "total" file
//   - contains the total number of pokemon files, used for random selection
func main() {
	args := parseArgs()

	totalFpath := path.Join(args.ToDir, args.ToTotalFname)
	categoryFpath := path.Join(args.ToDir, args.ToCategoryFname)
	entryDirPath := path.Join(args.ToDir, args.ToDataSubDir)
	metadataDirPath := path.Join(args.ToDir, args.ToMetadataSubDir)

	// ensure that the destination directories exist
	mkDirs([]string{entryDirPath, metadataDirPath})

	// Find all the cowfiles
	cowfileFpaths := pokedex.FindFiles(args.FromDir, ".cow", make([]string, 0))
	fmt.Println("- Found", len(cowfileFpaths), "cowfiles")
	// Read pokemon names
	pokemonNames := pokedex.ReadNames(args.FromMetadataFname)
	fmt.Println("- Read", len(pokemonNames), "pokemon names from", args.FromMetadataFname)

	// 1. Create the category struct using the cowfile paths, pokemon names and indexes\
	fmt.Println("- Writing categories to file")
	pokedex.WriteStructToFile(
		pokedex.CreateCategoryStruct(args.FromDir, cowfileFpaths, args.Debug),
		categoryFpath,
	)

	// 2. Create the metadata files, containing name/category/japanese name info for each pokemon
	metadata := pokedex.CreateMetadata(args.FromDir, cowfileFpaths, pokemonNames, args.Debug)

	fmt.Println("- Writing metadata and entries to file")
	pbar := bin.NewProgressBar(len(metadata))
	for _, m := range metadata {
		pokedex.WriteBytesToFile(m.Data, pokedex.EntryFpath(entryDirPath, m.Index), true)
		pokedex.WriteStructToFile(m.Metadata, pokedex.MetadataFpath(metadataDirPath, m.Index))
		pbar.Add(1)
	}
	fmt.Println("- Writing total metadata to file")
	pokedex.WriteIntToFile(len(metadata), totalFpath)

	fmt.Println("✓ Complete! Indexed", len(cowfileFpaths), "total cowfiles")
	fmt.Println("✓ Wrote gzipped metadata to", metadataDirPath)
	fmt.Println("✓ Wrote gzipped cowfiles to", entryDirPath)
	fmt.Println("✓ Wrote 'total' metadata to", totalFpath)
	fmt.Println("✓ Wrote gzipped category trie to", categoryFpath)
}
