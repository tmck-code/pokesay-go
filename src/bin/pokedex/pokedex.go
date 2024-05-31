package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/tmck-code/pokesay/src/bin"
	"github.com/tmck-code/pokesay/src/pokedex"
)

// Strips the leading "./" from a path e.g. "./cows/ -> cows/"
func normaliseRelativeDir(dirPath string) string {
	return strings.TrimPrefix(dirPath, "./")
}

type PokedexArgs struct {
	FromDir           string
	FromMetadataFname string
	ToDir             string
	Debug             bool
	ToDataSubDir      string
	ToMetadataSubDir  string
	ToTotalFname      string
}

type PokedexPaths struct {
	EntryDirPath    string
	MetadataDirPath string
	TotalFpath      string
}

func NewPokedexPaths(args PokedexArgs) PokedexPaths {
	return PokedexPaths{
		EntryDirPath:    path.Join(args.ToDir, args.ToDataSubDir),
		MetadataDirPath: path.Join(args.ToDir, args.ToMetadataSubDir),
		TotalFpath:      path.Join(args.ToDir, args.ToTotalFname),
	}
}

func parseArgs() PokedexArgs {
	fromDir := flag.String("from", "/tmp/cows", "from dir")
	fromMetadataFname := flag.String("fromMetadata", "/tmp/cows/pokemon.json", "metadata file")
	toDir := flag.String("to", "build/assets/", "to dir")

	toDataSubDir := flag.String("toDataSubDir", "cows/", "dir to write all binary (image) data to")
	toMetadataSubDir := flag.String("toMetadataSubDir", "metadata/", "dir to write all binary (metadata) data to")
	toTotalFname := flag.String("toTotalFname", "total.txt", "file to write the number of available entries to")
	debug := flag.Bool("debug", false, "show debug logs")

	flag.Parse()

	args := PokedexArgs{
		FromDir:           normaliseRelativeDir(*fromDir),
		FromMetadataFname: *fromMetadataFname,
		ToDir:             normaliseRelativeDir(*toDir),
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
		pokedex.Check(err)
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
	paths := NewPokedexPaths(args)

	// ensure that the destination directories exist
	mkDirs([]string{paths.EntryDirPath, paths.MetadataDirPath})

	// Find all the cowfiles
	cowfileFpaths := pokedex.FindFiles(args.FromDir, ".cow", make([]string, 0))
	fmt.Println("- Found", len(cowfileFpaths), "cowfiles")
	// Read pokemon names
	pokemonNames := pokedex.ReadNames(args.FromMetadataFname)
	nameTokens := pokedex.GatherMapKeys(pokemonNames)
	sort.Strings(nameTokens)
	fmt.Println("names:", nameTokens)

	fmt.Println("- Read", len(pokemonNames), "pokemon names from", args.FromMetadataFname)

	// 1. For each pokemon name, write a metadata file, containing the name information, and
	// links to all of the matching cowfile indexes
	fmt.Println("- Writing metadata to file")
	pokemonMetadata := make([]pokedex.PokemonMetadata, 0)
	uniqueNames := make(map[string][]int)
	nameVariants := make(map[string][]string)
	i := 0
	pbar := bin.NewProgressBar(len(pokemonNames))
	for i, key := range nameTokens {
		name := pokemonNames[key]
		// add variant
		metadata := pokedex.CreateNameMetadata(fmt.Sprintf("%04d", i), key, name, args.FromDir, cowfileFpaths)
		pokedex.WriteStructToFile(metadata, pokedex.MetadataFpath(paths.MetadataDirPath, i))
		pokemonMetadata = append(pokemonMetadata, *metadata)
		uniqueNames[name.Slug] = append(uniqueNames[name.Slug], i)
		i++
		pbar.Add(1)
	}

	fmt.Println("- Writing entries to file")
	pbar = bin.NewProgressBar(len(cowfileFpaths))
	for i, fpath := range cowfileFpaths {
		data, err := os.ReadFile(fpath)
		pokedex.Check(err)
		entryFpath := pokedex.EntryFpath(paths.EntryDirPath, i)
		fmt.Println("writing", fpath, ">", entryFpath)

		fpathParts := strings.Split(fpath, "/")
		basename := fpathParts[len(fpathParts)-1]
		name := strings.SplitN(strings.TrimSuffix(basename, ".cow"), "-", 2)[0]
		fmt.Println("name", name)
		for i, key := range nameTokens {
			if name == key {
				fmt.Println("found", name, key, i, pokemonNames[key], pokemonMetadata[i])
				nameVariants[name] = append(nameVariants[name], basename)
				fmt.Println("name ++", name, fmt.Sprintf("%04d.%04d", i, len(nameVariants[name])))
				break
			}
		}

		pokedex.WriteBytesToFile(data, entryFpath, true)
		pbar.Add(1)
	}

	pokedex.WriteStructToFile(uniqueNames, "build/assets/names.txt")

	// 2. Create the category struct using the cowfile paths, pokemon names and indexes\
	fmt.Println("- Writing categories to file")
	categories := pokedex.CreateCategoryStruct(args.FromDir, pokemonMetadata, args.Debug)
	pokedex.WriteStructToFile(categories, "build/assets/category_keys.txt")

	fmt.Println("- Writing total metadata to file")
	pokedex.WriteIntToFile(len(pokemonMetadata), paths.TotalFpath)

	fmt.Println("✓ Complete! Indexed", len(cowfileFpaths), "total cowfiles")
	fmt.Println("wrote", i, "names to", "build/assets/names.txt")

	fmt.Println("✓ Wrote gzipped metadata to", paths.MetadataDirPath)
	fmt.Println("✓ Wrote gzipped cowfiles to", paths.EntryDirPath)
	fmt.Println("✓ Wrote 'total' metadata to", paths.TotalFpath, len(pokemonMetadata))
}
